# This is the configuration file for the etcd server.

# Human-readable name for this member.
name: etcd-{{ .Role }}

# Path to the data directory.
data-dir: {{ .DataDir }}

client-transport-security:

  # Path to the client server TLS cert file.
  cert-file: {{ .ServerCertPath }}

  # Path to the client server TLS key file.
  key-file: {{ .ServerKeyPath }}

  # Enable client cert authentication.
  client-cert-auth: true

  # Path to the client server TLS trusted CA cert file.
  trusted-ca-file: {{ .CACertPath }}

  # Client TLS using generated certificates
  auto-tls: false

# List of this member's client URLs to advertise to the public.
# The URLs needed to be a comma-separated list.
advertise-client-urls: https://0.0.0.0:{{ .ETCDServicePort}}

# List of comma separated URLs to listen on for client traffic.
listen-client-urls: https://0.0.0.0:{{ .ETCDServicePort}}

# Initial cluster token for the etcd cluster during bootstrap.
initial-cluster-token: 'new'

# Initial cluster state ('new' or 'existing').
initial-cluster-state: 'new'

# Number of committed transactions to trigger a snapshot to disk.
snapshot-count: 75000

# Raise alarms when backend size exceeds the given quota. 0 means use the
# default quota.
quota-backend-bytes: 8589934592

# metrics configuration
metrics: basic
