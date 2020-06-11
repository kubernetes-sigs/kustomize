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

The docs are in the `site` directory

```shell script
cd site/
```

To view the docs run

```shell script
hugo serve
```

```shell script
...
Running in Fast Render Mode. For full rebuilds on change: hugo server --disableFastRender
Web Server is available at http://localhost:1313/kustomize/ (bind address 127.0.0.1)
```

and visit the URL that is printed

## Publishing

The `site` content is compiled by Hugo into the `docs` folder using the hugo command:

```shell script
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

Add the `site` and `docs` folders to a commit, and create a PR.