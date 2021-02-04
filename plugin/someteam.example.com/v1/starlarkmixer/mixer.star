def run(r, fc):
  add = []
  remove = []
  for resource in r:
    if resource.get("metadata") == None:
      resource["metadata"] = {}
    if resource["metadata"].get("annotations") == None:
      resource["metadata"]["annotations"] = {}

    # Flag for deletion
    if resource["metadata"]["name"] == "delete-me":
      remove.append(resource)
      continue

    # Deep-ish copy the resource
    cp = dict(resource)
    cp["metadata"] = dict(cp["metadata"])
    cp["metadata"]["annotations"] = dict(cp["metadata"]["annotations"])
    cp["metadata"]["name"] = resource["metadata"]["name"]+"-copy"
    add.append(cp)

    resource["metadata"]["annotations"]["modified-by"] = fc["metadata"]["name"]

  # Add something
  new = {
    "kind": "ConfigMap",
    "apiVersion": "v1",
    "metadata": {
      "name": "net-new"
    }
  }
  r.extend(add)
  r.append(new)
  for resource in remove:
    r.remove(resource)

run(ctx.resource_list["items"], ctx.resource_list["functionConfig"])
