package kv

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	"filippo.io/age"
	"filippo.io/age/agessh"
	"filippo.io/age/armor"
	"golang.org/x/crypto/ssh"
	"golang.org/x/term"
	yage "sylr.dev/yaml/age/v3"
	"sylr.dev/yaml/v3"
)

func decryptValueWithAge(value []byte, ids []age.Identity) ([]byte, error) {
	if len(ids) == 0 {
		return value, nil
	}

	var r io.Reader
	if bytes.HasPrefix(value, []byte(armor.Header)) {
		rr := bytes.NewBuffer(value)
		r = armor.NewReader(rr)
	} else {
		r = bytes.NewBuffer(value)
	}

	rd, err := age.Decrypt(r, ids...)
	if err != nil {
		return value, err
	}
	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(rd)
	if err != nil {
		return value, err
	}
	value = buf.Bytes()

	return value, nil
}

func decryptInPlaceYAMLWithAge(value []byte, ids []age.Identity) ([]byte, error) {
	if len(ids) == 0 {
		return value, nil
	}

	in := bytes.NewBuffer(value)

	node := yaml.Node{}
	w := yage.Wrapper{
		Value:      &node,
		Identities: ids,
	}

	out := new(bytes.Buffer)
	decoder := yaml.NewDecoder(in)
	encoder := yaml.NewEncoder(out)
	encoder.SetIndent(2)

	for {
		err := decoder.Decode(&w)

		if err == io.EOF {
			break
		} else if err != nil {
			return value, err
		}

		err = encoder.Encode(&node)

		if err != nil {
			return value, err
		}
	}

	return out.Bytes(), nil
}

// Following code has been copied from:
// - https://github.com/FiloSottile/age/blob/31e0d226807f679cce89b67dfde6201b62582158/cmd/age/parse.go
// - https://github.com/FiloSottile/age/blob/31e0d226807f679cce89b67dfde6201b62582158/cmd/age/encrypted_keys.go

func parseSSHIdentity(name string, pemBytes []byte) ([]age.Identity, error) {
	id, err := agessh.ParseIdentity(pemBytes)
	if sshErr, ok := err.(*ssh.PassphraseMissingError); ok {
		pubKey := sshErr.PublicKey
		if pubKey == nil {
			pubKey, err = readPubFile(name)
			if err != nil {
				return nil, err
			}
		}
		passphrasePrompt := func() ([]byte, error) {
			fmt.Fprintf(os.Stderr, "Enter passphrase for %q: ", name)
			pass, err := readPassphrase()
			if err != nil {
				return nil, fmt.Errorf("could not read passphrase for %q: %v", name, err)
			}
			return pass, nil
		}
		i, err := agessh.NewEncryptedSSHIdentity(pubKey, pemBytes, passphrasePrompt)
		if err != nil {
			return nil, err
		}
		return []age.Identity{i}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("malformed SSH identity in %q: %v", name, err)
	}

	return []age.Identity{id}, nil
}

func readPubFile(name string) (ssh.PublicKey, error) {
	if name == "-" {
		return nil, fmt.Errorf(`failed to obtain public key for "-" SSH key
Use a file for which the corresponding ".pub" file exists, or convert the private key to a modern format with "ssh-keygen -p -m RFC4716"`)
	}
	f, err := os.Open(name + ".pub")
	if err != nil {
		return nil, fmt.Errorf(`failed to obtain public key for %q SSH key: %v
Ensure %q exists, or convert the private key %q to a modern format with "ssh-keygen -p -m RFC4716"`, name, err, name+".pub", name)
	}
	defer f.Close()
	contents, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("failed to read %q: %v", name+".pub", err)
	}
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(contents)
	if err != nil {
		return nil, fmt.Errorf("failed to parse %q: %v", name+".pub", err)
	}
	return pubKey, nil
}

func readPassphrase() ([]byte, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		tty, err := os.Open("/dev/tty")
		if err != nil {
			return nil, fmt.Errorf("standard input is not available or not a terminal, and opening /dev/tty failed: %v", err)
		}
		defer tty.Close()
		fd = int(tty.Fd())
	}
	defer fmt.Fprintf(os.Stderr, "\n")
	p, err := term.ReadPassword(fd)
	if err != nil {
		return nil, err
	}
	return p, nil
}
