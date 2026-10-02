# Instructions for coding agents

These instructions apply to every AI coding agent working in this repository. Human contributors should use [CONTRIBUTING.md](CONTRIBUTING.md) as the canonical policy. If this file conflicts with `CONTRIBUTING.md` or Kubernetes contributor policy, follow the human-facing policy and report the conflict.

## Read before changing anything

1. Read [CONTRIBUTING.md](CONTRIBUTING.md), including its AI, duplicate-search, assignment, testing, and review requirements.
1. Read the upstream [Kubernetes AI Guidance](https://github.com/kubernetes/community/blob/main/contributors/guide/pull-requests.md#ai-guidance).
1. Read the nearest README, design document, Makefile, and test instructions for the area being changed. Do not invent commands or conventions.
1. Inspect the working tree and preserve unrelated human changes.

## Check for existing work first

Before implementation, use live GitHub data to search:

- open and closed issues;
- open, closed, and merged pull requests; and
- linked design proposals or Kubernetes Enhancement Proposals.

Use several queries based on the symptom, component, error text, and proposed behavior. Record the exact queries and the closest candidate URLs for the human handoff. Do not declare that an item is unique merely because its title differs.

If live access is unavailable, explicitly report that duplicate and assignment checks are unverified. Never fabricate search results, issue state, assignees, comments, approvals, test results, or links.

When work is associated with an existing issue:

- Read the complete issue discussion and check current assignees before editing.
- The human contributor must use `/assign` before implementation begins. If self-assignment fails, the human must leave a claim comment and wait for maintainer acknowledgement.
- If another account is assigned, stop. Work may resume only after the assignee publicly agrees to collaborate or a maintainer reassigns the issue. Record the discussion URL.
- Do not open a competing pull request or silently work around the assignment rule.

A feature implementation requires a `triage/accepted` issue or an approved applicable proposal. Ask for human direction when the required decision does not exist.

## Make reviewable changes

- Keep the change focused on the accepted problem. Do not add speculative cleanup, generated boilerplate, or unrelated fixes.
- Prefer the smallest coherent change that can be tested and reviewed independently. Do not split work merely to evade review limits.
- Add or update tests appropriate to the change. Bug fixes require a regression test and must follow the commit structure in `CONTRIBUTING.md`.
- Run relevant formatting, tests, and verification commands. Report the exact commands and outcomes; use `Not run: <reason>` for anything not run.
- Review the entire diff before handoff. Remove redundant comments, generated narration, accidental files, and unsupported claims.

## AI use and project communication

Follow the human-authored communication rules in `CONTRIBUTING.md`:

- You may assist privately with repository analysis and with code, test, or documentation changes, but a human remains responsible for every submitted line.
- Do not draft or rewrite substantive content for an issue or pull request description, issue or pull request comment, review, GitHub Discussion, project Slack message, or shared design discussion. Give the human concise facts and raw verification results so they can independently write the communication in their own words.
- After review feedback is received, do not analyze or summarize it, decide how to address it, implement the response, or draft, rewrite, or translate a reply. Hand the review round to the human author so they can interpret the feedback and engage directly with the reviewer without AI assistance.
- If a human supplies text and explicitly asks for light spelling, grammar, accessibility, or translation assistance, make only non-substantive edits and remind them to disclose that assistance. This exception does not apply to responses to review comments.
- Do not infer or allege AI use from writing style. Only report objective missing fields, explicit disclosures, prohibited trailers, or other verifiable evidence.
- Do not generate commit messages. Leave changes uncommitted unless the human supplies the exact message and explicitly asks you to commit. Do not add an AI system as author, co-author, signer, or in `assisted-by`, `co-developed-by`, `co-authored-by`, or similar trailers.

Do not create an issue, pull request, comment, review, label, assignment, or other external side effect unless the human explicitly authorizes that action. Even when authorized:

- Post an issue or pull request title and body only when the human has supplied the final text.
- Do not open any pull request until the human confirms that they reviewed every submitted line and validation result and can explain the change. Draft state does not replace this pre-submission review.
- When AI generated or materially rewrote code, tests, repository documentation, or substantive analysis in the diff, open the pull request as a GitHub Draft. Never mark it ready for review, remove its hold, or check the final human-review attestation on the human's behalf.
- Never post an AI-generated triage finding or suggested response. A human maintainer must verify the evidence and write any public communication.

## Handoff and policy self-check

Before handoff, use [.github/prompts/ai-contribution-triage.md](.github/prompts/ai-contribution-triage.md) as a private checklist. Its output is advisory and must not be pasted into a community space.

Provide the human with:

- changed files and a factual summary of behavior;
- duplicate-search queries and candidate URLs, or an explicit unverified status;
- issue assignment or coordination evidence when applicable;
- tests and checks run with their outcomes;
- risks, limitations, and anything not verified; and
- the AI tool provider, product, model/version (or `unknown/auto-selected`), and a brief factual description of how AI assisted.

Leave authorship decisions, the public description, commit message, Draft-to-ready transition, and reviewer conversation to the human.
