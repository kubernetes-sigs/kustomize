Find the document with the given `_id`:
```
curl -s -X GET "${ElasticSearchURL}:9200/${INDEXNAME}/_search?pretty" -H 'Content-Type: application/json' -d'
{
  "query": {
    "terms": {
      "_id": [ "b3a03f3327841617db696e2d6abc30e1a1bd653f1a2bbce05637f7dcae1a43f7" ] 
    }
  }
}
'
```
