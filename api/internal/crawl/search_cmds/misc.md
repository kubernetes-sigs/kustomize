Check the health status of an ElasticSearch cluster:
```
curl -X GET "${ElasticSearchURL}:9200/_cat/health?v&pretty"
```

Check the indices in an ElasticSearch cluster:
```
curl "${ElasticSearchURL}:9200/_cat/indices?v"
```

Get the mapping of the index:
```
curl -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_mapping?pretty"
```

Delete the kustomize index from the ElasticSearch cluster (**Use this command with caution**):
```
curl -X DELETE "${ElasticSearchURL}:9200/${INDEXNAME}?pretty"
```