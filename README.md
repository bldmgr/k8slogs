# k8slogs

```cmd
go build -o k8s-support-bundle .
```

```
k8s-support-bundle -n namespace-name
OutputDir:     "./pod-logs"

=== Pull pods logs from Kubernetes ===
Collecting logs for pod: api-service-7c4bd7879d-22qwc
  ✓ Saved logs to: pod-logs/api-service-7c4bd7879d-22qwc_api-service_20250818_221944.log
Collecting logs for pod: api-service-7c4bd7879d-9mjbt
  ✓ Saved logs to: pod-logs/api-service-7c4bd7879d-9mjbt_api-service_20250818_221944.log
Collecting logs for pod: api-service-7c4bd7879d-hh9z2
  ✓ Saved logs to: pod-logs/api-service-7c4bd7879d-hh9z2_api-service_20250818_221945.log
Collecting logs for pod: api-service-7c4bd7879d-m7v5f
  ✓ Saved logs to: pod-logs/api-service-7c4bd7879d-m7v5f_api-service_20250818_221945.log
Collecting logs for pod: audit-log-service-75df8959fb-bqkkz
  ✓ Saved logs to: pod-logs/audit-log-service-75df8959fb-bqkkz_audit-log-service_20250818_221946.log
Collecting logs for pod: authentication-svc-authapi-7fd565f79-b69xk
  ✓ Saved logs to: pod-logs/authentication-svc-authapi-7fd565f79-b69xk_20250818_221946.log
  ✓ Saved logs to: pod-logs/authentication-svc-authapi-7fd565f79-b69xk_20250818_221947.log
  ✓ Saved logs to: pod-logs/authentication-svc-authapi-7fd565f79-b69xk_20250818_221947.log
Collecting logs for pod: workflows-conductor-grpc-5df7d544c5-gmgxl
  ✓ Saved logs to: pod-logs/workflows-conductor-grpc-5df7d544c5-gmgxl_workflows-conductor-grpc-handler_20250818_222136.log
  ✓ Saved logs to: pod-logs/workflows-conductor-grpc-5df7d544c5-gmgxl_20250818_222139.log
  ✓ Saved logs to: pod-logs/workflows-conductor-grpc-5df7d544c5-gmgxl_flyway-table-rename_20250818_222139.log
  ✓ Saved logs to: pod-logs/workflows-conductor-grpc-5df7d544c5-gmgxl_repair-migrations_20250818_222139.log
  ✓ Saved logs to: pod-logs/workflows-conductor-grpc-5df7d544c5-gmgxl_20250818_222139.log
  ✓ Saved logs to: pod-logs/workflows-conductor-grpc-5df7d544c5-gmgxl_dbmigrator-approvaljobservice_20250818_222139.log
Log collection completed!
```
