
def reconcile(items):
  key = ctx.resource_list["functionConfig"]["spec"]["key"]
  value = ctx.resource_list["functionConfig"]["spec"]["value"]

  for resource in items:
    if resource.get("metadata") == None:
      resource["metadata"] = {}
    if resource["metadata"].get("labels") == None:
      resource["metadata"]["labels"] = {}
    resource["metadata"]["labels"][key] = value

reconcile(ctx.resource_list["items"])
