// Code generated by internal/gen/main.go; DO NOT EDIT.
package main

const (
	usageMsg = `
Helps when you have a git repository with multiple Go modules.

It handles tasks one might otherwise attempt with

'''
find ./ -name "go.mod" | xargs {some hack}
'''

Run it from a git repository root.

It walks the repository, reads 'go.mod' files, builds
a model of Go modules and intra-repo module
dependencies, then performs some operation.

Install:
'''
go get github.com/monopole/gorepomod
'''

## Usage

_Commands that change things (everything but 'list')
do nothing but log commands
unless you add the '--doIt' flag,
allowing the change._

#### 'gorepomod list'

Lists modules and intra-repo dependencies.

Use this to get module names for use in other commands.

#### 'gorepomod tidy'

Creates a change with mechanical updates
to 'go.mod' and 'go.sum' files.

#### 'gorepomod unpin {module}'

Creates a change to 'go.mod' files.

For each module _m_ in the repository,
if _m_ depends on a _{module}_,
then _m_'s dependency on it will be replaced by
a relative path to the in-repo module.

#### 'gorepomod pin {module} [{version}]'

Creates a change to 'go.mod' files.

The opposite of 'unpin'.
  
The change removes replacements and pins _m_ to a
specific, previously tagged and released version of _{module}_.

The argument _{version}_ defaults to recent version of _{module}_.

_{version}_ should be in semver form, e.g. 'v1.2.3'.


#### 'gorepomod release {module} [patch|minor|major]'

Computes a new version for the module, tags the repo
with that version, and pushes the tag to the remote.

The value of the 2nd argument, either 'patch' (the default),
'minor' or 'major', determines the new version.

If the existing version is _v1.2.7_, then the new version will be:
 - 'patch' -> _v1.2.8_
 - 'minor' -> _v1.3.0_
 - 'major' -> _v2.0.0_

After establishing the the version, the command looks for a branch named

> _release-{module}/-v{major}.{minor}_

If the branch doesn't exist, the command creates it and pushes it to the remote.

The command then creates a new tag in the form

> _{module}/v{major}.{minor}.{patch}_

The command pushes this tag to the remote.  This typically triggers
cloud activity to create release artifacts.

#### 'gorepomod unrelease {module}'

This undoes the work of 'release', by deleting the
most recent tag both locally and at the remote.

You can then fix whatever, and re-release.

This, however, must be done almost immediately.

If there's a chance someone (or some cloud robot) already
imported the module at the given tag, then don't do this,
because it will confuse module caches.

Do a new patch release instead.

`
)