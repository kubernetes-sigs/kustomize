SHELL := /bin/bash -euo pipefail

config_file_name = kustomize.yaml
example_config = docs/$(config_file_name)

.PHONY: all
all: docs

# In a branch, run 'make docs' to update docs with
# generated code, then merge it to master.
docs: $(example_config)

# Use kustomize to create the standard kustomize configuration
# file that appears in the website's documentation.
$(example_config): /tmp/bin/kustomize
	rm -f TMP
	echo "# This is a generated example; do not edit.  Rebuild with 'make docs'." >> TMP
	echo " " >> TMP
	/tmp/bin/kustomize init
	cat $(config_file_name) >> TMP
	mv TMP $(example_config)
	rm $(config_file_name)

/tmp/bin/kustomize:
	go build -o /tmp/bin/kustomize kustomize.go
