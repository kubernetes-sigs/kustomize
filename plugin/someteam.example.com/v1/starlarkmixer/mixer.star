def run(r, fc):
  for resource in r:
    if resource.get("metadata") == None:
      resource["metadata"] = {}
    if resource["metadata"].get("annotations") == None:
      resource["metadata"]["annotations"] = {}
    resource["metadata"]["annotations"]["modified-by"] = fc["metadata"]["name"]

  new = {
    "kind": "ConfigMap",
    "apiVersion": "v1",
    "metadata": {
      "name": "net-new"
    }
  }
  r.append(new)

run(ctx.resource_list["items"], ctx.resource_list["functionConfig"])
