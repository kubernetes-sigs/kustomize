# Warning
#
# This makefile is provided only as a convenience to run
# the travis pre-commit script, which generates all the
# code that needs to be generated, and runs all the tests.
#
# This makefile won't be maintained until it replaces
# (or is used by) the travis/pre-commit.sh script

all:
	./travis/pre-commit.sh

.PHONY: all
