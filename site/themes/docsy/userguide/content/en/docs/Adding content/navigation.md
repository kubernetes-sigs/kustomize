---
title: "Navigation and Search"
date: 2017-01-05
weight: 3
description: >
   Customize site navigation and search for your Docsy site.
---

## Top-level menu

The top level menu (the one that appears in the top navigation bar for the entire site) uses your site's [`main` menu](https://gohugo.io/content-management/menus/). All Hugo sites have a `main` menu array of menu entries, accessible via the `.Site.Menus` site variable and populatable via page front matter or your site's `config.toml`. 

To add a page or section to this menu, add it to the site's `main` menu in either `config.toml` or in the destination page's front matter (in `_index.md` or `_index.html` for a section, as that's the section landing page). For example, here's how we added the Documentation section landing page to the main menu in this site:

```yaml
---
title: "Docsy Documentation"
linkTitle: "Documentation"
menu:
  main:
    weight: 20
---
```

The menu is ordered from left to right by page `weight`. So, for example, a section index or page with `weight: 30` would appear after the Documentation section in the menu, while one with `weight: 10` would appear before it.

If you want to add a link to an external site to this menu, add it in `config.toml`, specifying the `weight`.

```yaml
[[menu.main]]
    name = "GitHub"
    weight = 50
    url = "https://github.com/google/docsy/"
```

### Adding a version drop-down

If you add some `[params.versions]` in `config.toml`, the Docsy theme adds a
version selector drop down to the top-level menu.

You can find out more in the guide to 
[versioning your docs](/docs/adding-content/versioning/).

### Adding a language drop-down

If you configure more than one language in `config.toml`, the Docsy theme adds a language selector drop down to the top-level menu. Selecting a language takes the user to the translated version of the current page, or the home page for the given language.

You can find out more in [Multi-language support](/docs/language/).

## Section menu

The section menu, as shown in the left side of the `docs` section, is automatically built from the `content` tree. Like the top-level menu, it is ordered by page or section index `weight` (or by page creation `date` if `weight` is not set), with the page or index's `Title`, or `linkTitle` if different, as its link title in the menu. If a section subfolder has pages other than `_index.md` or `_index.html`, those pages will appear as a submenu, again ordered by `weight`. For example, here's the metadata for this page showing its `weight` and `title`:

```yaml
---
title: "Navigation and Search"
linkTitle: "Navigation and Search"
date: 2017-01-05
weight: 3
description: >
  Customize site navigation and search for your Docsy site.
---
```

To hide a page or section from the menu, set `toc_hide: true` in front matter.

By default, the section menu will show the current section fully expanded all the way down. This may make the left nav too long and difficult to scan for bigger sites. Try setting site param `ui.sidebar_menu_compact = true` in `config.toml`.

## Breadcrumb navigation

Breadcrumb navigation is enabled by default. To disable breadcrumb navigation, set site param `ui.breadcrumb_disable = true` in `config.toml`.

## Site search options

Docsy offers multiple options that let your readers search your site content, so you can pick one that suits your needs. You can choose from:

* [Google Custom Search Engine](#configure-search-with-a-google-custom-search-engine) (GCSE), the default option, which uses Google's index of your public site to generate a search results page.
* [Algolia DocSearch](#configure-algolia-docsearch), which uses Algolia's indexing and search mechanism, and provides an organized dropdown of search results when your readers use the search box. Algolia DocSearch is free for public documentation sites.
* [Local search with Lunr](#configure-local-search-with-lunr), which uses Javascript to index and search your site without the need to connect to external services. This option doesn't require your site to be public.

If you enable any of these search options in your `config.toml`, a search box displays in the right of your top navigation bar. By default a search box also displays at the top of the section menu in the left navigation pane, which you can disable if you prefer, or if you're using a search option that only works with the top search box.

Be aware that if you accidentally enable more than one search option in your `config.toml` you may get unexpected results (for example, if you have added the `.js` for Algolia DocSearch, you'll get Algolia results if you enable GCSE search but forget to disable Algolia search).

### Disabling the sidebar search box

By default, the search box appears in both the top navigation bar and at the top of the sidebar left navigation pane. If you don't want the sidebar search box, set `sidebar_search_disable` to `true` in `config.toml`:

```
sidebar_search_disable = true
```

## Configure search with a Google Custom Search Engine

By default Docsy uses a [Google Custom Search Engine](https://cse.google.com/cse/all) (GCSE) to search your site. To enable this feature, you'll first need to make sure that you have built a public production version of your site, as otherwise your site won't be crawled and indexed.

### Setting up site search

1.  Deploy your site and ensure that it's built with `HUGO_ENV="production"`, as Google will only crawl and index Docsy sites built with this setting (you probably don't want your not-ready-for-prime-time site to be searchable!). You can specify this variable as a command line flag to Hugo: 

    ```
    $ env HUGO_ENV="production" hugo
    ```

    Alternatively, if you're using Netlify, you can specify it as a Netlify [deployment setting](https://www.netlify.com/docs/continuous-deployment/#build-environment-variables) in `netlify.toml` or the Netlify UI, along with the Hugo version. It may take a day or so before your site has actual search results available.
2.  Create a Google Custom Search Engine for your deployed site by clicking **New search engine** on the [Custom Search page](https://cse.google.com/cse/all) and following the instructions. Make a note of the ID for your new search engine.
3.  Add any further configuration you want to your search engine using the **Edit search engine** options. In particular you may want to do the following:

    * Select **Look and feel**. Change from the default **Overlay** layout to **Results only**, as this option means your search results are embedded in your search page rather than appearing in a separate box. Click **Save** to save your changes.
    * Edit the default result link behavior so that search results from your site don't open in a new tab. To do this, select **Search Features** - **Advanced** - **Websearch Settings**. In the **Link Target** field, type "\_parent". Click **Save** to save your changes.
    
{{% alert title="Tip" %}}
Your site search results should show up within a couple of days. If it takes longer than that, you can manually request that your site is indexed by [submitting a sitemap through the Google Search Console](https://support.google.com/webmasters/answer/183668?hl=en).
{{% /alert %}}

### Adding the search page

Once you have your search engine set up, you can add the feature to your site:

1. Ensure you have a Markdown file in `content/en/search.md` (and one per other languages if needed) to display your search results. It only needs a title and `layout: search`, as in the following example:

    ```
    ---
    title: Search Results
    layout: search
    ---
    ```

1. Add your Google Custom Search Engine ID to the site params in `config.toml`. You can add different values per language if needed.

    ```
    # Google Custom Search Engine ID. Remove or comment out to disable search.
    gcs_engine_id = "011737558837375720776:fsdu1nryfng"
    ```

### Disabling GCSE search

If you don't specify a Google Custom Search Engine ID for your project and haven't enabled any other search options, the search box won't appear in your site. If you're using the default `config.toml` from the example site and want to disable search, just comment out or remove the relevant line.

## Configure Algolia DocSearch

As an alternative to GCSE, you can use [Algolia DocSearch](https://community.algolia.com/docsearch/) with this theme. Algolia DocSearch is free for public documentation sites.

### Sign up for Algolia DocSearch

Complete the form at [https://community.algolia.com/docsearch/#join-docsearch-program](https://community.algolia.com/docsearch/#join-docsearch-program).

If you are accepted to the program, you will receive the JavaScript code to add to your documentation site from Algolia by email.

### Adding Algolia DocSearch

1. Enable Algolia DocSearch in `config.toml`.

    ```
    # Enable Algolia DocSearch
    algolia_docsearch = true
    ```

2. Remove or comment out any GCSE ID in `config.toml` and ensure local search is set to `false` as you can only have one type of search enabled. See [Disabling GCSE search](#disabling-gcse-search).

3. Disable the sidebar search in `config.toml` as this is not currently supported for Algolia DocSearch. See [Disabling the sidebar search box](#disabling-the-sidebar-search-box).

3. Add the JavaScript code provided to you by Algolia to the head and body of every page on your site. See [Add code to head or before body end](/docs/adding-content/lookandfeel/#add-code-to-head-or-before-body-end) for details.

4. Update the `inputSelector` field in the body end Javascript with the appropriate CSS selector (e.g. `.td-search-input` to use the default CSS from this theme).

When you've completed these steps the Algolia search should be enabled on your site. Search results are displayed as a drop-down under the search box, so you don't need to add any search results page.

## Configure local search with Lunr

[Lunr](https://lunrjs.com/) is a Javascript-based search option that lets you index your site and make it searchable without the need for external, server-side search services. This is a good option particularly for smaller or non-public sites.

To add Lunr search to your Docsy site:

1. Enable local search in `config.toml`.

    ```
    # Enable local search
    offlineSearch = true
    ```

2. Remove or comment out any GCSE ID in `config.toml` and ensure Algolia DocSearch is set to `false`, as you can only have one type of search enabled. See [Disabling GCSE search](#disabling-gcse-search).

Once you've completed these steps, local search is enabled for your site and results appear in a drop down when you use the search box.

{{% alert title="Tip" %}}
If you're [testing this locally](/docs/deployment/#serving-your-site-locally) using Hugo’s local server functionality, you need to build your `offline-search-index.xxx.json` file first by running `hugo`. If you have the Hugo server running while you build `offline-search-index.xxx.json`, you may need to stop the server and restart it in order to see your search results.
{{% /alert %}}
