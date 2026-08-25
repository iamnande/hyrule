{{- define "app.pod" -}}
metadata:
  labels:
    {{- include "app.selectorLabels" . | nindent 4 }}
spec:
  serviceAccountName: {{ include "app.serviceAccountName" . }}
  {{- with .Values.imagePullSecrets | default (.Values.global).imagePullSecrets }}
  imagePullSecrets:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  {{- with .Values.initContainers }}
  initContainers:
    {{- toYaml . | nindent 4 }}
  {{- end }}
  containers:
    - name: {{ include "app.name" . }}
      image: {{ include "app.image" . }}
      imagePullPolicy: {{ .Values.image.pullPolicy }}
      ports:
        - name: http
          containerPort: {{ .Values.service.port }}
      {{- if or .Values.env .Values.extraEnv }}
      env:
        {{- range $key, $value := .Values.env }}
        - name: {{ $key }}
          value: {{ $value | quote }}
        {{- end }}
        {{- with .Values.extraEnv }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
      {{- end }}
      {{- with .Values.extraEnvFrom }}
      envFrom:
        {{- toYaml . | nindent 8 }}
      {{- end }}
      {{- if .Values.probes.startup.enabled }}
      startupProbe:
        httpGet:
          path: {{ .Values.probes.startup.path }}
          port: http
        failureThreshold: 30
        periodSeconds: 1
      {{- end }}
      {{- if .Values.probes.liveness.enabled }}
      livenessProbe:
        httpGet:
          path: {{ .Values.probes.liveness.path }}
          port: http
      {{- end }}
      {{- if .Values.probes.readiness.enabled }}
      readinessProbe:
        httpGet:
          path: {{ .Values.probes.readiness.path }}
          port: http
      {{- end }}
      resources:
        {{- toYaml .Values.resources | nindent 8 }}
      {{- with .Values.extraVolumeMounts }}
      volumeMounts:
        {{- toYaml . | nindent 8 }}
      {{- end }}
    {{- with .Values.additionalContainers }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
  {{- with .Values.extraVolumes }}
  volumes:
    {{- toYaml . | nindent 4 }}
  {{- end }}
{{- end -}}
