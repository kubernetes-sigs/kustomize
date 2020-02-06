Retrieve information about all registered snapshot repositories:
```
curl -s -X GET "${ElasticSearchURL}:9200/_snapshot?pretty"
```

Retrieve information about a given snapshot repository, `kustomize-backup`:
```
curl -s -X GET "${ElasticSearchURL}:9200/_snapshot/kustomize-backup?pretty"
```

Verify a snapshot repository, `kustomize-backup`, manually:
```
curl -s -X POST "${ElasticSearchURL}:9200/_snapshot/kustomize-backup/_verify?pretty"
```

List all the snapshots in a given snapshot repository:
```
curl -s -X GET "${ElasticSearchURL}:9200/_cat/snapshots/kustomize-backup?v&s=id&pretty"
```

Retrieve a summary information about a given snapshot:
```
curl -s -X GET "${ElasticSearchURL}:9200/_snapshot/kustomize-backup/kustomize-snapshot?pretty"
```

Retrieve a detailed information about a given snapshot:
```
curl -s -X GET "${ElasticSearchURL}:9200/_snapshot/kustomize-backup/kustomize-snapshot/_status?pretty"
```
