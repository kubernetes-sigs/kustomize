def run(r, fc):
  for resource in r:
    resource["metadata"]["annotations"]["a-string-value"] = fc["data"]["stringValue"]
    resource["metadata"]["annotations"]["a-int-value"] = fc["data"]["intValue"]
    resource["metadata"]["annotations"]["a-bool-value"] = fc["data"]["boolValue"]

run(ctx.resource_list["items"], ctx.resource_list["functionConfig"])
