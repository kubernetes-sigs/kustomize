# Kustomize Enhancement Proposal Processes

So you want to propose an enhancement to Kustomize—awesome! Choose the option below that best fits the scope of your idea. Before you get started, it's a good idea to review the list of [eschewed features](https://kubectl.docs.kubernetes.io/faq/kustomize/eschewedfeatures).

[SIG CLI]: https://github.com/kubernetes/community/tree/master/sig-cli
[Enhancements repo]: https://github.com/kubernetes/enhancements

### Option 1: Github issue

Small, straightforward enhancements can be proposed in regular GitHub issues. As a rule of thumb, the enhancement should be resolvable in a single PR that is at most size L.

**Example features**:
- a new Kustomization field that does something very straightforward, like annotating resources
- a new option for an existing built-in transformer

**Instructions**: [Open an issue](https://github.com/kubernetes-sigs/kustomize/issues/new?labels=kind%2Ffeature&template=feature_request.md)

### Option 2: Mini (In-Repo) Enhancement Proposal

If your feature may be controversial or has a lot of details to explain, you should write it up as a mini enhancement proposal on this repo. This process is still relatively lightweight, but allows for more in-depth discussion of multiple details. Because it is submitted as a PR, reviewers can comment on individual lines of the proposal, facilitating discussion. The PR will be merged whether the feature is accepted or rejected, creating a record of the decisions.

Since the proposal template is a subset of what's required for the full KEP process, this can also be a good option if you're unsure whether your proposal is of general interest to [SIG-CLI]. You can use it to get preliminary feedback from the Kustomize team before bringing a full-fledged KEP to the SIG.

**Example features**:
- a new built-in transformer with several configurable options
- a feature that will bring in a significant new dependency
- a feature that introduces a new class of behavior, such as manipulating data within an opaque resource field
- a new build process that does not affect `kubectl kustomize`

**Instructions**:
1. Make a copy of [00-00-template.md](00-00-template.md) and rename it with the current date (e.g. 21-08) and a succinct title.
1. Fill out the template.
1. Submit it for review as a PR.
1. (Optional) Present your proposal at a biweekly [SIG-CLI] meeting. This can be a good way to get more traction for your proposal. Your presentation should be a quick summary to help folks understand whether the proposal is relevant to them.

### Option 3: Kubernetes Enhancement Proposal (KEP)

If your feature changes behaviour in a way that has significance for `kubectl kustomize`, particularly in terms of security, you will need to follow the full KEP process and get buy-in from [SIG-CLI] leadership in addition to Kustomize maintainers. Note that you can still submit a mini (in-repo) enhancement proposal as a first step to get preliminary feedback, including on whether a full KEP is required.

**Example features**:
- a feature with significant privacy or security implications to work out
- a feature that is not purely localized to the Kustomize binary on the end user's machine (e.g. downloads something remote, executes something external)

**Instructions**:
1. Follow the process on the [Enhancements repo]. Be sure to put your KEP in the directory for SIG-CLI.
1. (Strongly recommended) Send a link to your KEP to the [SIG-CLI] mailing list.
1. (Strongly recommended) Present your KEP at a biweekly [SIG-CLI] meeting. Your presentation should be a quick summary to help folks understand whether the KEP is relevant to them.
1. After your KEP is accepted, remember to update its metadata as your feature proceeds through the release process.
