## completion

Generate shell completion.

### Synopsis

Generate shell completion for `kustomize` -- supports bash, zsh, fish and powershell.

### Examples

    # load completion for Bash
    source <(kustomize completion bash)

    # install for Bash in Linux
    kustomize completion bash > /etc/bash_completion.d/kustomize

    # install for Bash in MacOS
    kustomize completion bash > /usr/local/etc/bash_completion.d/kustomize

    # package for Bash
    kustomize completion bash > /usr/share/bash-completion/completions/kustomize

    # package for zsh
    kustomize completion zsh > /usr/share/zsh/site-functions/_kustomize

