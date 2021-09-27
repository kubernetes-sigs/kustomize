---
title: "Doc Versioning"
date: 2020-02-02
weight: 4
description: >
   Customize navigation and banners for multiple versions of your docs.
---

Depending on your project's releases and versioning, you may want to let your
users access previous versions of your documentation. How you deploy the
previous versions is up to you. This page describes the Docsy features that you
can use to provide navigation between the various versions of your docs and
to display an information banner on the archived sites.

## Adding a version drop-down menu

If you add some `[params.versions]` in `config.toml`, the Docsy theme adds a 
version selector drop down to the top-level menu. You specify a URL and a name 
for each version you would like to add to the menu, as in the following example:

```
# Add your release versions here
[[params.versions]]
  version = "master"
  url = "https://master.kubeflow.org"

[[params.versions]]
  version = "v0.2"
  url = "https://v0-2.kubeflow.org"

[[params.versions]]
  version = "v0.3"
  url = "https://v0-3.kubeflow.org"
```

Remember to add your current version so that users can navigate back!

The default title for the version drop-down menu is **Releases**. To change the
title, change the `version_menu` parameter in `config.toml`:

```
version_menu = "Releases"
```

You can read more about Docsy menus in the guide to
[navigation and search](/docs/adding-content/navigation/).

## Displaying a banner on archived doc sites

If you create archived snapshots for older versions of your docs, you can add a 
note at the top of every page in the archived docs to let readers know that
they’re seeing an unmaintained snapshot and give them a link to the latest 
version.

For example, see the archived docs for 
[Kubeflow v0.6](https://v0-6.kubeflow.org/docs/):

<figure>
  <img src="/images/version-banner.png"
       alt="A text box explaining that this is an unmaintained snapshot of the docs."
       class="mt-3 mb-3 border border-info rounded" />
  <figcaption>Figure 1. The banner on the archived docs for Kubeflow v0.6
  </figcaption>
</figure>

To add the banner to your doc site, make the following changes in your
`config.toml` file:

1. Set the `archived_version` parameter to `true`:

    ```
    archived_version = true
    ```

1. Set the `version` parameter to the version of the archived doc set. For
  example, if the archived docs are for version 0.1:

    ```
    version = "0.1"
    ```

1. Make sure that `url_latest_version` contains the URL of the website that you
  want to point readers to. In most cases, this should be the URL of the latest 
  version of your docs:

    ```
    url_latest_version = "https://your-latest-doc-site.com"
    ```
