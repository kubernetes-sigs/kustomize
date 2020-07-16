---
title: "Writing Docs"
linkTitle: "Writing Docs"
type: docs
weight: 30
description: >
    How to make Kustomize docs contributions
---

Kustomize uses [Docsy](https://www.docsy.dev) for the site, and was
forked from the [docsy-example](https://github.com/google/docsy-example)

## Prerequisites

- [Install hugo](https://gohugo.io/getting-started/installing/#fetch-from-github)
- Clone kustomize
  - `git clone git@github.com:kubernetes-sigs/kustomize && cd kustomize/`

## Development

The doc input files are in the `site` directory.  The site can be hosted locally using
`hugo serve`.

```shell script
cd site/
hugo serve
```

```shell script
...
Running in Fast Render Mode. For full rebuilds on change: hugo server --disableFastRender
Web Server is available at http://localhost:1313/kustomize/ (bind address 127.0.0.1)
```

## Publishing

Hugo compiles the files under `site` Hugo into html which it puts in the `docs` folder:

```shell script
cd site/
hugo
```

```shell script
                   | EN  
-------------------+-----
  Pages            | 99  
  Paginator pages  |  0  
  Non-page files   |  0  
  Static files     | 47  
  Processed images |  0  
  Aliases          |  2  
  Sitemaps         |  1  
  Cleaned          |  0  
```

Add the `site/` and `docs/` folders to a commit, then create a PR.

## Publishing docs to your kustomize fork

It is possible to have the kustomize docs published to your forks github pages.

### Setup GitHub Pages for the fork

1. Go to the *forked repo's* **Settings** tab
   - e.g. [https://github.com/pwittrock/kustomize](https://github.com/pwittrock/kustomize)
2. Go to the **GitHub Pages** section
3. Set the source to master branch **/docs folder**

### Publish to the fork's GitHub Pages

{{% pageinfo color="info" %}}
Changes must be pushed to the fork's **master branch** to be served as the fork's GitHub Page.
{{% /pageinfo %}}

1. Make a change to a file under `site/content`
2. Run `hugo` from the `site/` directory
3. Add the `site` and `docs` directories to the **master branch**
4. Commit and push the changes to the *remote fork's* **master branch**
5. After a few minutes, the docs should be served from the fork's GitHub Page
   - e.g. [https://pwittrock.github.io/kustomize/](https://pwittrock.github.io/kustomize/)
