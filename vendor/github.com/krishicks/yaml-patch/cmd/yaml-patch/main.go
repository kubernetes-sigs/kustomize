package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	flags "github.com/jessevdk/go-flags"
	yamlpatch "github.com/krishicks/yaml-patch"
)

type opts struct {
	OpsFiles []FileFlag `long:"ops-file" short:"o" value-name:"PATH" description:"Path to file with one or more operations"`
}

func main() {
	var o opts
	_, err := flags.Parse(&o)
	if err != nil {
		log.Fatalf("error: %s\n", err)
	}

	placeholderWrapper := yamlpatch.NewPlaceholderWrapper("{{", "}}")

	var patches []yamlpatch.Patch
	for _, opsFile := range o.OpsFiles {
		var bs []byte
		bs, err = ioutil.ReadFile(opsFile.Path())
		if err != nil {
			log.Fatalf("error reading opsfile: %s", err)
		}

		var patch yamlpatch.Patch
		patch, err = yamlpatch.DecodePatch(placeholderWrapper.Wrap(bs))
		if err != nil {
			log.Fatalf("error decoding opsfile: %s", err)
		}

		patches = append(patches, patch)
	}

	doc, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("error reading from stdin: %s", err)
	}

	mdoc := placeholderWrapper.Wrap(doc)
	for _, patch := range patches {
		mdoc, err = patch.Apply(mdoc)
		if err != nil {
			log.Fatalf("error applying patch: %s", err)
		}
	}

	fmt.Printf("%s", placeholderWrapper.Unwrap(mdoc))
}
