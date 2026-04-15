{{/*
  _helpers.tpl — shared template helpers for FleetOS Helm chart
*/}}

{{- define "fleetos.fullname" -}}
{{- printf "%s-%s" .Release.Name .Chart.Name | trunc 63 | trimSuffix "-" }}
{{- end }}

{{- define "fleetos.labels" -}}
helm.sh/chart: {{ .Chart.Name }}-{{ .Chart.Version }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "fleetos.selectorLabels" -}}
app.kubernetes.io/name: {{ .name }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{- define "fleetos.image" -}}
{{- if .Values.global.imageRegistry -}}
{{ .Values.global.imageRegistry }}/{{ .image }}:{{ .Values.global.imageTag }}
{{- else -}}
{{ .image }}:{{ .Values.global.imageTag }}
{{- end }}
{{- end }}

{{- define "fleetos.dbEnv" -}}
- name: DATABASE_URL
  valueFrom:
    secretKeyRef:
      name: {{ .Values.externalDatabase.existingSecret }}
      key: password
      optional: true
- name: DB_HOST
  value: {{ .Values.externalDatabase.host | quote }}
- name: DB_PORT
  value: {{ .Values.externalDatabase.port | quote }}
- name: DB_NAME
  value: {{ .Values.externalDatabase.name | quote }}
- name: DB_USER
  value: {{ .Values.externalDatabase.user | quote }}
{{- end }}

{{- define "fleetos.redisEnv" -}}
- name: REDIS_URL
  value: {{ printf "redis://%s:%d" .Values.externalRedis.host (.Values.externalRedis.port | int) | quote }}
{{- end }}

{{- define "fleetos.natsEnv" -}}
- name: NATS_URL
  value: "nats://nats:4222"
{{- end }}

{{- define "fleetos.jwtEnv" -}}
- name: JWT_PUBLIC_KEY_PATH
  value: /secrets/public.pem
- name: JWT_PRIVATE_KEY_PATH
  value: /secrets/private.pem
{{- end }}
