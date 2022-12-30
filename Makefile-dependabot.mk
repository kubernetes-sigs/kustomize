DOCKER_DIRS := $(shell find . -type f -name "Dockerfile" -exec dirname {} \; | sort -u | egrep  '^./' )

GOMOD_DIRS := $(shell find . -type f -name "go.mod" -exec dirname {} \; | sort -u | egrep  '^./' )

NPM_DIRS := $(shell find . -type f -name "package.json" -exec dirname {} \; | sort -u | egrep  '^./' )

DEPENDABOT_PATH=".github/dependabot.yml"

.PHONY: dependabot/update
dependabot/update:
	@echo "Add update rule for \"${PACKAGE}\" in \"${DIR}\"";
	@echo "  - package-ecosystem: \"${PACKAGE}\"" >> ${DEPENDABOT_PATH};
	@echo "    directory: \"${DIR}\"" >> ${DEPENDABOT_PATH};
	@echo "    schedule:" >> ${DEPENDABOT_PATH};
	@echo "      interval: \"weekly\"" >> ${DEPENDABOT_PATH};
	@echo "" >> ${DEPENDABOT_PATH};

# This target should run on /bin/bash since the syntax DIR=$${dir:1} is not supported by /bin/sh.
.PHONY: generate/dependabot
generate/dependabot: $(eval SHELL:=/bin/bash)
	@echo "Recreating ${DEPENDABOT_PATH} file"
	@echo "# File generated by \"make generate/dependabot\"; DO NOT EDIT." > ${DEPENDABOT_PATH}
	@echo "" >> ${DEPENDABOT_PATH}
	@echo "version: 2" >> ${DEPENDABOT_PATH}
	@echo "updates:" >> ${DEPENDABOT_PATH}
	@set -e; for dir in $(DOCKER_DIRS); do \
		$(MAKE) dependabot/update DIR=$${dir:1} PACKAGE="docker"; \
	done
	$(MAKE) dependabot/update DIR="/" PACKAGE="github-actions"
	@set -e; for dir in $(GOMOD_DIRS); do \
		$(MAKE) dependabot/update DIR=$${dir:1} PACKAGE="gomod"; \
	done
	@set -e; for dir in $(NPM_DIRS); do \
		$(MAKE) dependabot/update DIR=$${dir:1} PACKAGE="npm"; \
	done