The crawler job can run in one of the following mode:

# Crawling all the documents in the index and crawling all the kustomization files on Github

This is the default setting of the crawler job. The `command` and `args` field
of the container should be:

```
        command: ["/crawler"]
```

Or

```
        command: ["/crawler"]
        args: ["--mode=index+github"]
```

# Crawling all the documents in the index

The `command` and `args` field of the container should be:

```
        command: ["/crawler"]
        args: ["--mode=index"]
```

# Crawling all the kustomization files on Github

The `command` and `args` field of the container should be:

```
        command: ["/crawler"]
        args: ["--mode=github"]
```

# Crawling all the kustomization files in a Github repo

The `command` and `args` field of the container should be like:

```
        command: ["/crawler"]
        args: ["--mode=github-repo", "--github-repo=kubernetes-sigs/kustomize"]
```

# Crawling all the kustomization files in all the repositories of a Github user

The `command` and `args` field of the container should be like:

```
        command: ["/crawler"]
        args: ["--github-user", "--github-user=kubernetes-sigs"]
```
