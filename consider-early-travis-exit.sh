# Exits with status 0 if it can be determined that the
# current PR should not trigger all travis checks.
#
# This could be done with a "git ...|grep -vqE" oneliner
# but as travis triggering is refined it's useful to check
# travis logs to see how branch files were considered.
function consider-early-travis-exit {
  if [ "$TRAVIS_PULL_REQUEST" == "false" ]; then
    echo "Not a travis pull request."
    return
  fi
  if [ -z "$TRAVIS_BRANCH" ]; then
    echo "Unknown travis branch."
    return
  fi
  echo "TRAVIS_BRANCH=$TRAVIS_BRANCH"
  local branchFiles=$(git diff --name-only FETCH_HEAD...$TRAVIS_BRANCH)
  local invisibles=0
  local triggers=0
  echo "Branch Files (X==invisible to travis):"
  echo "---"
  for fn in $branchFiles; do
    if [[ "$fn" =~ (\.md$)|(^docs/) ]]; then
      echo "  X  $fn"
      let invisibles+=1
    else
      echo "     $fn"
      let triggers+=1
    fi
  done
  echo "---"
  printf >&2 "%6d files invisible to travis.\n" $invisibles
  printf >&2 "%6d files trigger travis.\n" $triggers
  if [ $triggers -eq 0 ]; then
    echo "Exiting travis early."
    # see https://github.com/travis-ci/travis-build/blob/master/lib/travis/build/templates/header.sh
    travis_terminate 0
  fi
}
consider-early-travis-exit
unset -f consider-early-travis-exit
