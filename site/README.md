# kustomize website: the alpha of the alpha

## Overview
This is just an example overview of what the new kustomize website might look like. It is forked from the [docsy example website](https://example.docsy.dev/) and heavy based on that.

I'm not a frontend dev but I was mostly successful at creating what I thought would be a good outline. However, I couldn't get rid of that picture of porridge with blueberries on it on the landing page! So ignore that and imagine it's something more nautical.

I put the most effort into the `Documentation` section. The left-menu bar has the custom structure that is my creation, based on the docsy example, the content of the current kustomize documentation sources and a general vibe of documentation sites I find easy to read.

The top bar is customized with the sections I think make sense to split. However, I have customized nothing else inside the `Community`, `Contribute` and `Blog` sections.

## Building

Build and run using Docker or Hugo, then access the site at `http://localhost:1313`.

### Docker
Dependencies:
* [docker](https://docs.docker.com/engine/install/)
* [docker-compose](https://docs.docker.com/compose/install/)
```bash
docker-compose build
docker-compomse up -d
```

### hugo
Building using the `hugo` command requires the following dependencies:
* [hugo CLI](https://gohugo.io/getting-started/installing/)
* [Go](https://go.dev/learn/)
* [Node.js](https://nodejs.org/en/)
* npm dependencies
   ```bash
   npm install -D autoprefixer
   npm install -D postcss-cli
   npm install -D postcss
   ```
Start in development mode:
```bash
hugo serve -D
```
