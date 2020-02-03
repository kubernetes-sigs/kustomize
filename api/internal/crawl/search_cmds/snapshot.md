Retrieve information about all registered snapshot repositories:
```
curl -X GET "${ElasticSearchURL}:9200/_snapshot?pretty"
```

Retrieve information about a given snapshot repository, `gcskustomize`:
```
curl -X GET "${ElasticSearchURL}:9200/_snapshot/gcskustomize?pretty"
```

Verify a snapshot repository, `gcskustomize`, manually:
```
curl -X POST "${ElasticSearchURL}:9200/_snapshot/gcskustomize/_verify?pretty"
```

Retrieve a summary information about a given snapshot:
```
curl -X GET "${ElasticSearchURL}:9200/_snapshot/gcskustomize/kustomize-snapshot?pretty"
```

Retrieve a detailed information about a given snapshot:
```
curl -X GET "${ElasticSearchURL}:9200/_snapshot/gcskustomize/kustomize-snapshot/_status?pretty"
```
