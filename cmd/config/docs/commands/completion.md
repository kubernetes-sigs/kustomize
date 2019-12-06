## install-completion

Install shell completion.

### Synopsis

Install shell completion for kustomize commands and flags -- supports bash, fish and zsh.

    kustomize install-completion

Registers the completion command with known shells (e.g. .bashrc, .bash_profile, etc):

    complete -C /Users/USER/go/bin/kustomize kustomize

Because the completion command is embedded in kustomize directly, there is no need to update
it separately from the kustomize binary.

To uninstall shell completion run:

    COMP_UNINSTALL=1 kustomize install-completion
