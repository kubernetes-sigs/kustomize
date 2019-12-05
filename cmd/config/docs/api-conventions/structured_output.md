## Structured Output for Kubernetes CLI Tools
Below is a protocol for emitting machine parseable structured output format (e.g. -o structured) from Kubernetes project tools. 
This protocol would provide an API for better command integration & composition of tools in the Kubernetes ecosystem.

## Structured Output Schema:
Define the following schema for rendering structured output.

```
{
  "$id": "https://example.com/person.schema.json",
  "$schema": "http://json-schema.org/draft-07/schema#",
  "title": "OutputMessage",
  "version": "0.1.0"
  "type": "object",
	"description": "The specification for a structured output message",
	"definitions": {
		"error_details":{
			"type": "object",
			"required": ["error"],
			"properties": {
				"error": {
					"type": "string",
					"description": "The exception or error being logged e.g. 'User Error'
					or ValidationException."
				},
				"context": {
					"type": "string",
					"description": "Any stack-trace or additional error context."
				}
			}
		}
	},
	"required": ["version", "timestamp", "body"],
	"properties": {
		"version": {
			"type": "string",
			"description": "Semantic version of the message format e.g. v1.0.0.
			Useful if message format changes before binary is updated."
		},
		"timestamp": {
			"type": "string",
			"description": "[RFC 3339](https://www.ietf.org/rfc/rfc3339.txt) encoded
			timestamp."

		},
		"body": {
			"type": "string",
			"description": "The actual content of the message. JSON/YAML structured
			content will be assumed to be resource content otherwise it will be
			assumed to be plain text."
		},
		"error_detail": {
			"$ref": "#/definitions/error_details",
			"description": "Only present if the message is related to an error."

		},
		"verbosity": {
			"enum": ["debug", "info", "warn", "error", "critical"]
		}
	}
}
```


## `structued` output format

For those tools (such as `kubectl`) that support the output flag (e.g.  `-o`, `--output`)
a new `structured` option should be added. This should change the behavior of all output logged or written to both stdout and stderr for a given command. 
This is particularly useful for commands that may return resources and log error messages on failure but other commands would be modified to also support 
an output flag so that output parsing could be done uniformly across the tool.

## Output Message Severity/Verbosity (Optional)

In addition to supporting the `structured` output format, tools that support outputting messages at different levels of severity should use the following
standard logging severities for easier filtering of generated messages:

- debug
- info
- warn
- error
- critical

## Message Versioning

To support message format extensions and backward compatibility, messages should include
a semantic version (e.g. `v0.1.0`) that will correspond to the version of the structured output json 
schema that was used to generate the message.
 
## Handling unstructured output from nested libraries

Some Kubernetes project tool commands  may invoke code from libraries which write output directly to stderr or stdout that is not structured. 

In these cases there are two main options:
- Do nothing and let whatever higher level context is invoking the tool provide 
  a 'catch-all' mechanism for dealing with any unstructured data it receives on
  stdout or stderr. 
  
- The Kubernetes project tool itself can provide such a mechanism by capturing
  stdout/stderr while invoking the nested library and then using that output to 
  build a structured message.  

## Examples

### Fetching resources:
`kubectl get pods -o structured`

```json
{
  "version": "0.1.0",
  "timestamp": "2019-11-20T00:00:00Z",
  "verbosity": "info",
  "body":
          {
            "apiVersion": "v1",
            "items": ...,
            "kind": "List",
            "metadata": {
                "resourceVersion": ... ,
                "selfLink": ...
            }
          }
}
```

### Fetching resources with errors:
`kubectl get pod foo -o structured`


```json
{
  "version": "0.1.0",
  "timestamp": "2019-11-20T00:00:00Z",
  "body": "",
  "verbosity": "error",
  "error_detail":
    {
      "error": "NotFound",
      "context": "pods 'foo' not found"
    }
}
```

### Executing command with non-resource output:
`kubectl apply -f bar.yaml -o structured`
```json
{
  "version": "0.1.0",
  "timestamp": "2019-11-20T00:00:00Z",
  "body": "kubectl apply should be used on resource created by either kubectl create --save-config or kubectl apply",
  "verbosity": "warning"
}

{
  "version": "0.1.0",
  "timestamp": "2019-11-20T00:00:00Z",
  "body": "namespace/bar-resource configured",
  "verbosity": "info"
}
```

