# Kubernetes Manifests

This directory contains Kubernetes manifests for deploying tiny-bitly to GKE.

## Files

- `namespace.yaml` - Creates the `tiny-bitly` namespace
- `service-account.yaml` - Service account for Cloud SQL Proxy access
- `deployment.yaml` - Main application deployment with Cloud SQL Proxy sidecar
- `service.yaml` - Service (LoadBalancer by default) for exposing the app publicly
- `hpa.yaml` - Horizontal Pod Autoscaler for auto-scaling
- `ingress.yaml` - External HTTP(S) load balancer (optional)
- `migration-job.yaml` - One-time job for running database migrations

## Setup Instructions

1. **Update placeholders** in the manifests:
   - Replace `PROJECT_ID` with your GCP project ID
   - Replace the Cloud SQL instance connection name placeholder:
     - `PROJECT_ID:us-central1:tiny-bitly-db` in `deployment.yaml` and `migration-job.yaml`
     - Format is `PROJECT_ID:REGION:INSTANCE_ID` (example: `tiny-bitly:us-central1:tiny-bitly-db`)
   - Replace `tiny-bitly.example.com` in `ingress.yaml` with your domain (or remove ingress if not needed)

2. **Create secrets and configmaps** (see main GKE_DEPLOYMENT.md guide)

3. **Apply manifests**:
   ```bash
   kubectl apply -f k8s/
   ```

## Customization

### Resource Limits

Adjust CPU and memory requests/limits in `deployment.yaml` based on your needs:

```yaml
resources:
  requests:
    cpu: "100m"      # Minimum guaranteed
    memory: "256Mi"
  limits:
    cpu: "1000m"     # Maximum allowed
    memory: "512Mi"
```

### Replica Count

Change initial replica count in `deployment.yaml`:

```yaml
spec:
  replicas: 2  # Change this
```

Or let HPA manage it automatically (recommended).

### Auto-Scaling

Adjust HPA settings in `hpa.yaml`:

```yaml
spec:
  minReplicas: 2    # Minimum pods
  maxReplicas: 10   # Maximum pods
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        averageUtilization: 70  # Scale when CPU > 70%
```

## Notes

- **Why this can feel “surprisingly hard”**
  Even though this is “one Go service on Kubernetes”, deploying it on **GKE Autopilot** with **Cloud SQL** and **Memorystore** pulls in a few systems that commonly create first-deploy edge cases:
  - **Managed services add identity plumbing**: Cloud SQL access via Cloud SQL Auth Proxy requires **Workload Identity + IAM roles** to be correct, otherwise the proxy starts but fails when the app first connects.
  - **Kubernetes doesn’t expand env vars in `command`/`args`**: strings like `$(INSTANCE_CONNECTION_NAME)` are passed literally; you must provide the Cloud SQL connection name directly (format: `PROJECT_ID:REGION:INSTANCE_ID`).
  - **Jobs + sidecars don’t mix by default**: a Kubernetes Job only reaches `Completed` when **all containers exit**. A proxy sidecar naturally runs forever, so the migration Job uses a pattern that starts/stops the proxy from within the migration container.
  - **Autopilot enforces stricter policies**: common “just kill the sidecar” patterns can be blocked by security constraints; prefer designs that don’t rely on cross-container process control.
  - **CPU architecture surprises**: local builds (often `arm64`) can fail to run on GKE nodes (`amd64`) unless you build/push the correct platform image (e.g. using `docker buildx --platform linux/amd64`).

- The Cloud SQL Proxy runs as a sidecar container in each pod
- The migration Job does **not** use a long-running sidecar (Jobs only complete when all containers exit). Instead, it runs the proxy inside the migration container so the Job can reach `Completed`.
- Health checks ensure pods are ready before receiving traffic
- All containers run as non-root users for security
- Secrets should be managed via Secret Manager in production
