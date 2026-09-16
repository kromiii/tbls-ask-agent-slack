{{/*
Expand the name of the chart.
*/}}
{{- define "tbls-ask-agent-slack.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "tbls-ask-agent-slack.fullname" -}}
{{- if .Values.fullnameOverride }}
{{- .Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $name := default .Chart.Name .Values.nameOverride }}
{{- if contains $name .Release.Name }}
{{- .Release.Name | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "tbls-ask-agent-slack.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "tbls-ask-agent-slack.labels" -}}
helm.sh/chart: {{ include "tbls-ask-agent-slack.chart" . }}
{{ include "tbls-ask-agent-slack.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "tbls-ask-agent-slack.selectorLabels" -}}
app.kubernetes.io/name: {{ include "tbls-ask-agent-slack.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
app: {{ include "tbls-ask-agent-slack.name" . }}
{{- end }}

{{/*
Get the secret name
*/}}
{{- define "tbls-ask-agent-slack.secretName" -}}
{{- if .Values.secret.existingSecret }}
{{- .Values.secret.existingSecret }}
{{- else if .Values.secret.name }}
{{- .Values.secret.name }}
{{- else }}
{{- include "tbls-ask-agent-slack.fullname" . }}
{{- end }}
{{- end }}

{{/*
Get the configmap name
*/}}
{{- define "tbls-ask-agent-slack.configMapName" -}}
{{- if .Values.schemaConfig.existingConfigMap }}
{{- .Values.schemaConfig.existingConfigMap }}
{{- else }}
{{- printf "%s-schemas" (include "tbls-ask-agent-slack.fullname" .) }}
{{- end }}
{{- end }}
