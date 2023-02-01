{{- define "apiversion" -}}
{{- if .Capabilities.APIVersions.Has "foo/v1" -}}
foo/v1
{{- else -}}
apps/v1
{{- end -}}
{{- end -}}