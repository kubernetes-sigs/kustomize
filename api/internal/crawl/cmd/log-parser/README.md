This binary takes as its input a json file including GKE logs (which can be
[exported](https://cloud.google.com/logging/docs/export/configure_export_v2) into 
[Cloud Storage](https://cloud.google.com/storage/docs/)), 
and extracts the `textPayload` field of each log entry.

Here is an log entry example:

{"insertId":"1sxuh4jg5lw6w10","labels":{"compute.googleapis.com/resource_name":"gke-crawler2-default-pool-5e55ea05-gzgv","container.googleapis.com/namespace_name":"default","container.googleapis.com/pod_name":"kustomize-stats-5bczg","container.googleapis.com/stream":"stdout"},"logName":"projects/haiyanmeng-gke-dev/logs/kustomize-stats","receiveTimestamp":"2020-01-06T23:33:07.012831742Z","resource":{"labels":{"cluster_name":"crawler2","container_name":"kustomize-stats","instance_id":"8183086081854184383","namespace_id":"default","pod_id":"kustomize-stats-5bczg","project_id":"haiyanmeng-gke-dev","zone":"us-central1-a"},"type":"container"},"severity":"INFO","textPayload":"The kustomize index already exists\n","timestamp":"2020-01-06T23:32:46.628930547Z"}
