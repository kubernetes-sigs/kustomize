# Exits with status 0 if it can be determined that the
# current PR should not trigger all travis checks.
#
# This could be done with a "git ...|grep -vqE" oneliner
# but as travis triggering is refined it's useful to check
# travis logs to see how branch files were considered.
function consider-early-travis-exit {
  if [ "$TRAVIS_PULL_REQUEST" == "false" ]; then
    echo "Unknown pull request."
    return
  fi
  # Might use this to improve checks on multi-commit PRs.
  echo "TRAVIS_COMMIT_RANGE=$TRAVIS_COMMIT_RANGE"
  echo "Branch Files ('T'==trigger tests, ' '=ignore):"
  echo "---"
  local triggers=0
  local invisibles=0
  for fn in $(git diff --name-only HEAD origin/master); do
    if [[ "$fn" =~ (\.md$)|(^docs/) ]]; then
      echo "     $fn"
      let invisibles+=1
    else
      echo "  T  $fn"
      let triggers+=1
    fi
  done
  echo "---"
  printf >&2 "%6d files invisible to travis.\n" $invisibles
  printf >&2 "%6d files trigger travis.\n" $triggers
  if [ $triggers -eq 0 ]; then
    echo "No files triggered travis test, exiting early."
    # see https://github.com/travis-ci/travis-build/blob/master/lib/travis/build/templates/header.sh
    travis_terminate 0
  fi
}
consider-early-travis-exit
unset -f consider-early-travis-exit
