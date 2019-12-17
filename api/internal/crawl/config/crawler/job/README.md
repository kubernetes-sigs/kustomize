There are three ways of running the crawler job.

# Crawling all the documents in the index and crawling all the kustomization files on Github

This is the default setting of the crawler job.

# Crawling all the documents in the index

Set the environment variable `CRAWL_INDEX_ONLY` to `true` like this:

```
        - name: CRAWL_INDEX_ONLY
          value: true
```

# Crawling all the kustomization files on Github

Set the environment variable `CRAWL_GITHUB_ONLY` to `true` like this:

```
        - name: CRAWL_GITHUB_ONLY
          value: true
```

# Crawling all the kustomization files in a Github repo

Add the environment variable `GITHUB_REPO` into the crawler container. For example:

```
        - name: GITHUB_REPO
          value: kubernetes-sigs/kustomize
```

# Crawling all the kustomization files in all the repositories of a Github user

Add the environment variable `GITHUB_USER` into the crawler container. For example:

```
        - name: GITHUB_USER
          value: kubernetes-sigs
```
