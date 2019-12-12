#!/usr/bin/env sh
set -e

# We assume one go file per plugin, and the desired name of the plugin .so
# matches the basename of said go file.
# Example: ReplacementTransformer.go -> ReplacementTransformer.so
#
# We also assume the script is run from a kustomize plugin directory.
# Usually one of:
#   - ${HOME}/.config/kustomize/plugin
#   - ${kustomize_git_repo}/plugin
plugins_dir="${PWD}"
command="${1}"

function buildPlugin {
  pushd "${1}" >& /dev/null
  plugin_src="$(find "${1}" -name '*.go'|grep -Ev '.*_test.go'|head -n1)"
  plugin_bin="$(echo "${plugin_src}"|sed 's/\(.*\).go/\1.so/')"
  if [ "${command}" = "clean" ]; then
    echo "Deleting ${plugin_bin}"
    rm -f "${plugin_bin}"
    return
  fi
  echo "Building ${plugin_bin}"
  go build -buildmode plugin -o "${plugin_bin}" "${plugin_src}"
  popd >& /dev/null
}

for goMod in $(find "${plugins_dir}" -name 'go.mod'); do
  d=$(dirname "${goMod}")
  # Skip "builtin" plugins in kustomize repo.
  if [ ! -z "$(echo "${goMod}" | grep -F "/plugin/builtin/")" ]; then
    continue
  fi
  # Skip plugins with only test Go files.
  if [ -z "$(find "${d}" -maxdepth 1 -name '*.go' | grep -Ev ".*_test.go")" ]; then
    continue
  fi
  buildPlugin "${d}"
done
