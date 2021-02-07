// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package types

// KvPairSources defines places to obtain key value pairs.
type KvPairSources struct {
	// LiteralSources is a list of literal
	// pair sources. Each literal source should
	// be a key and literal value, e.g. `key=value`
	LiteralSources []string `json:"literals,omitempty" yaml:"literals,omitempty"`

	// FileSources is a list of file "sources" to
	// use in creating a list of key, value pairs.
	// A source takes the form:  [{key}=]{path}
	// If the "key=" part is missing, the key is the
	// path's basename. If they "key=" part is present,
	// it becomes the key (replacing the basename).
	// In either case, the value is the file contents.
	// Specifying a directory will iterate each named
	// file in the directory whose basename is a
	// valid configmap key.
	FileSources []string `json:"files,omitempty" yaml:"files,omitempty"`

	// EnvSources is a list of file paths.
	// The contents of each file should be one
	// key=value pair per line, e.g. a Docker
	// or npm ".env" file or a ".ini" file
	// (wikipedia.org/wiki/INI_file)
	EnvSources []string `json:"envs,omitempty" yaml:"envs,omitempty"`

	// AgeIdentitySources is a list of file paths.
	// The contents of each files should be an AGE identity file.
	// AGE decryption will only apply to sources which have a '.age' suffix, e.g.:
	//
	//    envs:
	//    - foo.env.age
	//    file:
	//    - assets/id_rsa.age
	//    - id_rsa=assets/id_rsa.age
	//    literals:
	//    - |
	//      id_rsa.age=-----BEGIN AGE ENCRYPTED FILE-----
	//      YWdlLWVuY3J5cHRpb24ub3JnL3YxCi0+IFgyNTUxOSBPUjFuSDFaUStEaE5SQVhl
	//      ...
	//      SUrmocOpPl1j/aosTU0opGJ67FKiDu/XgVkJ9e4Sln386eTAwSFap1Z5FW7BFiZC
	//      -----END AGE ENCRYPTED FILE-----
	//
	// The `.age` suffix of the generated keys is removed.
	//
	// By default kustomize will also try do decrypt encrypted values with OpenSSH
	// keys found at ~/.ssh/id_rsa and ~/.ssh/id_ed25519. If either one of these
	// keys can decrypt your age encrypted data you can leave this property empty.
	AgeIdentitySources []string `json:"ageIdentities,omitempty" yaml:"ageIdentities,omitempty"`
}
