{{- define "app.pod" -}}
metadata:
  labels:
    {{- include "app.selectorLabels" . | nindent 4 }}
  {{- if include "app.hasSettings" . }}
  annotations:
    config-hash-{{ include "app.fullname" . }}: {{ merge ((.Values.global).runtimeSettings | default dict) (.Values.runtimeSettings | default dict) | toYaml | sha256sum }}
  {{- end }}
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
      {{- if include "app.hasSettings" . }}
      args:
        - "--config"
        - "/app/conf/{{ include "app.fullname" . }}.yml"
      {{- end }}
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
      {{- if or (include "app.hasSettings" .) .Values.extraVolumeMounts }}
      volumeMounts:
        {{- if include "app.hasSettings" . }}
        - name: config
          mountPath: /app/conf
        {{- end }}
        {{- with .Values.extraVolumeMounts }}
        {{- toYaml . | nindent 8 }}
        {{- end }}
      {{- end }}
    {{- with .Values.additionalContainers }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
  {{- if or (include "app.hasSettings" .) .Values.extraVolumes }}
  volumes:
    {{- if include "app.hasSettings" . }}
    - name: config
      configMap:
        name: {{ include "app.fullname" . }}-config
    {{- end }}
    {{- with .Values.extraVolumes }}
    {{- toYaml . | nindent 4 }}
    {{- end }}
  {{- end }}
{{- end -}}
