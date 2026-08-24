{{- define "pings.fullname" -}}
{{ .Release.Name }}
{{- end -}}

{{- define "pings.labels" -}}
app.kubernetes.io/name: pings
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}

{{- define "pings.selectorLabels" -}}
app.kubernetes.io/name: pings
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end -}}
