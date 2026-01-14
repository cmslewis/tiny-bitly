# Google Kubernetes Engine (GKE) Deployment Guide

This guide explains how to deploy the tiny-bitly service to **Google Kubernetes Engine (GKE)**, the industry-standard container orchestration platform. This approach provides auto-scaling, self-healing, rolling updates, and integrated monitoring/logging.

## Why GKE?

### Benefits Over VM-Based Deployment

- **Auto-scaling**: Automatically scale pods based on CPU, memory, or custom metrics
- **Self-healing**: Automatically restart failed containers
- **Zero-downtime deployments**: Rolling updates with health checks
- **Service discovery**: Services automatically find each other
- **Integrated monitoring**: Cloud Monitoring and Cloud Logging built-in
- **Declarative configuration**: Define desired state, Kubernetes makes it happen
- **Resource efficiency**: Better utilization than static VMs
- **Production-ready**: Industry standard for containerized applications

### Architecture

```
Internet
  ↓
GKE Ingress (Cloud Load Balancer)
  ↓
GKE Autopilot Cluster (Fully Managed)
  ├─ tiny-bitly Deployment (2-10 pods, auto-scaled)
  │   ├─ App Container (Go application)
  │   └─ Cloud SQL Proxy Sidecar (secure DB connection)
  ├─ Service (internal load balancer)
  └─ HorizontalPodAutoscaler (auto-scaling)
  ↓
Cloud SQL (PostgreSQL) - Managed
Memorystore (Redis) - Managed
```

**Note:** This guide uses GKE Autopilot, which automatically manages nodes, scaling, and infrastructure. All node management is handled by Google.

## Prerequisites

- GCP account with billing enabled
- `gcloud` CLI installed ([installation guide](https://cloud.google.com/sdk/docs/install))
- `kubectl` CLI installed (will be installed during setup)
- `docker` installed (for building container images)
- Basic familiarity with Kubernetes concepts ([Kubernetes basics](https://kubernetes.io/docs/tutorials/kubernetes-basics/))

## Step-by-Step Setup

### 1. Create GCP Project and Enable APIs

```bash
# Authenticate with GCP
gcloud auth login
gcloud auth application-default login

# Create or select a project
gcloud projects create YOUR_PROJECT_ID --name="tiny-bitly" || \
gcloud config set project YOUR_PROJECT_ID

# Verify project is set correctly
gcloud config get-value project
# If the above doesn't show your project ID, set it explicitly:
# gcloud config set project YOUR_PROJECT_ID

# Enable required APIs
gcloud services enable container.googleapis.com      # GKE
gcloud services enable sqladmin.googleapis.com       # Cloud SQL
gcloud services enable redis.googleapis.com          # Memorystore
gcloud services enable monitoring.googleapis.com     # Cloud Monitoring
gcloud services enable logging.googleapis.com        # Cloud Logging
```

**Official Tutorials:**
- [Creating and managing projects](https://cloud.google.com/resource-manager/docs/creating-managing-projects)
- [Enabling APIs](https://cloud.google.com/apis/docs/getting-started)

### 2. Create Cloud SQL PostgreSQL Instance

**Important:** Make sure your GCP project is set before running these commands. Verify with:
```bash
gcloud config get-value project
```

If it's not set, run:
```bash
gcloud config set project YOUR_PROJECT_ID
```

```bash
# Create PostgreSQL instance
# Note: If you get "Invalid project" error, ensure project is set (see above)
gcloud sql instances create tiny-bitly-db \
  --database-version=POSTGRES_16 \
  --tier=db-f1-micro \
  --region=us-central1 \
  --edition=ENTERPRISE \
  --root-password=YOUR_SECURE_ROOT_PASSWORD \
  --storage-type=SSD \
  --storage-size=20GB \
  --backup-start-time=03:00

# Create database
gcloud sql databases create tiny-bitly --instance=tiny-bitly-db

# Create application user
gcloud sql users create app_user \
  --instance=tiny-bitly-db \
  --password=YOUR_SECURE_APP_PASSWORD

# Get connection name (needed for Cloud SQL Proxy)
CLOUD_SQL_CONNECTION_NAME=$(gcloud sql instances describe tiny-bitly-db \
  --format='get(connectionName)')
echo "Cloud SQL connection: $CLOUD_SQL_CONNECTION_NAME"
```

**Note:** Save the connection name for later use in Kubernetes configuration.

**Official Tutorials:**
- [Creating Cloud SQL instances](https://cloud.google.com/sql/docs/postgres/create-instance)
- [Connecting to Cloud SQL](https://cloud.google.com/sql/docs/postgres/connect-overview)

### 3. Create Memorystore Redis Instance

```bash
# Create Redis instance
gcloud redis instances create tiny-bitly-redis \
  --size=1 \
  --region=us-central1 \
  --tier=basic \
  --redis-version=redis_7_0

# Get Redis IP (needed for application config)
REDIS_IP=$(gcloud redis instances describe tiny-bitly-redis \
  --region=us-central1 \
  --format='get(host)')
echo "Redis IP: $REDIS_IP"
```

**Official Tutorials:**
- [Creating Memorystore instances](https://cloud.google.com/memorystore/docs/redis/create-instance)
- [Connecting to Memorystore](https://cloud.google.com/memorystore/docs/redis/connect-redis-instance)

### 4. Create GKE Cluster

This guide uses **GKE Autopilot**, Google's fully managed Kubernetes service. Autopilot handles node management, scaling, and infrastructure automatically, so you can focus on your application.

**Benefits of Autopilot:**
- **Fully managed**: Google handles node provisioning, scaling, and maintenance
- **Cost-effective**: Pay only for requested pod resources, not entire nodes
- **Security**: Built-in security best practices and automatic updates
- **Simplified operations**: No need to configure node pools, machine types, or scaling policies
- **Production-ready**: SLA-backed with automatic high availability

**Alternative:** If you need more control over node configuration, you can use a Standard cluster instead (see note at end of this section).

```bash
# Set variables
PROJECT_ID=$(gcloud config get-value project)
CLUSTER_NAME=tiny-bitly-cluster
REGION=us-central1

# Create GKE Autopilot cluster
gcloud container clusters create-auto $CLUSTER_NAME \
  --region=$REGION \
  --logging=SYSTEM,WORKLOAD \
  --monitoring=SYSTEM \
  --release-channel=regular

# Get cluster credentials (this installs kubectl if needed)
gcloud container clusters get-credentials $CLUSTER_NAME --region=$REGION

# Verify cluster is running
kubectl cluster-info
kubectl get nodes
```

**Autopilot Configuration Explained:**
- **create-auto**: Creates an Autopilot cluster (fully managed)
- **region**: Multi-zone region for high availability (Autopilot automatically distributes pods)
- **logging**: `SYSTEM,WORKLOAD` enables logging for system components and workloads
- **monitoring**: `SYSTEM` enables monitoring for system components
- **release-channel**: `regular` provides a balance of new features and stability

**Note:** With Autopilot, you don't need to specify:
- Machine types (Google optimizes automatically)
- Node counts (scales automatically based on pod requirements)
- Node pool configurations (managed automatically)
- Auto-scaling settings (handled automatically)

**Autopilot Requirements:**
- All pods must have resource requests and limits defined (CPU and memory)
- Some advanced Kubernetes features may not be available (e.g., hostPath volumes, privileged containers)
- Pods are automatically scheduled across multiple zones for high availability
- See [Autopilot restrictions](https://cloud.google.com/kubernetes-engine/docs/concepts/autopilot-overview#autopilot_limitations) for full details

**Official Tutorials:**
- [Creating GKE clusters](https://cloud.google.com/kubernetes-engine/docs/how-to/creating-a-cluster)
- [Cluster autoscaling](https://cloud.google.com/kubernetes-engine/docs/how-to/cluster-autoscaler)

### 5. Build and Push Container Image

First, create a Dockerfile (see `Dockerfile` in project root):

```bash
# Build the container image
PROJECT_ID=$(gcloud config get-value project)
IMAGE_NAME=gcr.io/$PROJECT_ID/tiny-bitly
IMAGE_TAG=latest

# Build image
docker build -t $IMAGE_NAME:$IMAGE_TAG .

# Configure Docker to use gcloud as credential helper
gcloud auth configure-docker

# Push image to Google Container Registry
docker push $IMAGE_NAME:$IMAGE_TAG
```

**Alternative: Use Cloud Build** (recommended for CI/CD):

```bash
# Submit build to Cloud Build
gcloud builds submit --tag $IMAGE_NAME:$IMAGE_TAG
```

**Official Tutorials:**
- [Container Registry](https://cloud.google.com/container-registry/docs)
- [Cloud Build](https://cloud.google.com/build/docs)

### 6. Create Kubernetes Secrets

Store sensitive configuration (database passwords, etc.) in Kubernetes Secrets:

```bash
# Create secret for database credentials
kubectl create secret generic tiny-bitly-secrets \
  --from-literal=postgres-password=YOUR_SECURE_APP_PASSWORD \
  --from-literal=root-password=YOUR_SECURE_ROOT_PASSWORD
```

**Note:** For production, consider using [Secret Manager](https://cloud.google.com/secret-manager) instead.

### 8. Create Kubernetes ConfigMap

Store non-sensitive configuration:

```bash
# Create ConfigMap from environment variables
kubectl create configmap tiny-bitly-config \
  --from-literal=postgres-db=tiny-bitly \
  --from-literal=postgres-user=app_user \
  --from-literal=postgres-port=5432 \
  --from-literal=redis-host=$REDIS_IP \
  --from-literal=redis-port=6379 \
  --from-literal=api-port=8080 \
  --from-literal=log-level=info \
  --from-literal=rate-limit-requests-per-second=100 \
  --from-literal=rate-limit-burst=200
```

**Note:** Replace `$REDIS_IP` with the actual Redis IP from step 3.

### 9. Deploy Application to GKE

**Important:** Before deploying, update placeholders in the manifests:

```bash
# Replace PROJECT_ID in all manifests
sed -i '' 's/PROJECT_ID/YOUR_PROJECT_ID/g' k8s/*.yaml

# Or manually edit:
# - k8s/deployment.yaml (image path)
# - k8s/migration-job.yaml (image path)
# - k8s/service-account.yaml (PROJECT_ID)
```

Create Kubernetes manifests (see `k8s/` directory):

```bash
# Apply all Kubernetes manifests
kubectl apply -f k8s/

# Or apply individually:
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml
kubectl apply -f k8s/hpa.yaml
kubectl apply -f k8s/ingress.yaml
```

**Verify deployment:**

```bash
# Check pods are running
kubectl get pods -n tiny-bitly

# Check services
kubectl get services -n tiny-bitly

# Check deployment status
kubectl get deployment tiny-bitly -n tiny-bitly

# View pod logs
kubectl logs -f deployment/tiny-bitly -n tiny-bitly
```

### 10. Run Database Migrations

Run migrations using a Kubernetes Job:

```bash
# Create migration job (uses generateName, so use create not apply)
kubectl create -f k8s/jobs/migration-job.yaml

# Check job status
kubectl get jobs -n tiny-bitly

# View migration logs (Cloud SQL Proxy logs are redirected unless migrations fail)
kubectl logs -n tiny-bitly -l job-name --tail=200
```

### 11. Access Your Application

Get the external IP or load balancer address:

```bash
# If using LoadBalancer service
kubectl get service tiny-bitly -n tiny-bitly

# If using Ingress
kubectl get ingress tiny-bitly-ingress -n tiny-bitly
```

Access your application at the provided IP/URL.

## Kubernetes Manifests Overview

The deployment uses the following Kubernetes resources:

### Deployment (`k8s/deployment.yaml`)
- Defines the application container and Cloud SQL Proxy sidecar
- Sets resource limits and requests
- Configures health checks (liveness and readiness probes)
- Mounts secrets and configmaps

### Service (`k8s/service.yaml`)
- Exposes the application internally within the cluster
- Can be type `ClusterIP` (internal) or `LoadBalancer` (external)

### HorizontalPodAutoscaler (`k8s/hpa.yaml`)
- Automatically scales pods based on CPU/memory usage
- Example: Scale between 2-10 pods based on CPU > 70%

### Ingress (`k8s/ingress.yaml`)
- Provides HTTP(S) load balancing
- SSL/TLS termination
- Path-based routing

### Job (`k8s/jobs/migration-job.yaml`)
- Runs database migrations as a one-time job
- Uses the same container image as the application

## Monitoring and Logging

### Cloud Logging (Automatic)

GKE automatically collects logs from container stdout/stderr. View logs in:

- **GCP Console**: [Logs Explorer](https://console.cloud.google.com/logs)
- **Command line**: `kubectl logs -f deployment/tiny-bitly -n tiny-bitly`
- **Filter by resource**: `resource.type="k8s_container" AND resource.labels.cluster_name="tiny-bitly-cluster"`

### Cloud Monitoring (Automatic)

GKE automatically collects metrics. View metrics in:

- **GCP Console**: [Monitoring](https://console.cloud.google.com/monitoring)
- **Metrics include**: CPU, memory, network, disk usage per pod/container

### Custom Metrics (Prometheus)

Your application exposes Prometheus metrics at `/metrics`. To scrape them:

1. **Option A**: Use [GKE Managed Prometheus](https://cloud.google.com/stackdriver/docs/managed-prometheus)
2. **Option B**: Deploy Prometheus in the cluster (see `k8s/prometheus.yaml`)

### Creating Dashboards

1. Go to [Cloud Monitoring Dashboards](https://console.cloud.google.com/monitoring/dashboards)
2. Create custom dashboards using MQL (Monitoring Query Language)
3. Query metrics like:
   - `kubernetes.io/container/cpu/core_usage_time`
   - `kubernetes.io/container/memory/bytes_used`
   - Custom metrics from `/metrics` endpoint

## Scaling

### Manual Scaling

```bash
# Scale deployment to specific number of replicas
kubectl scale deployment tiny-bitly --replicas=5 -n tiny-bitly
```

### Auto-Scaling (HPA)

The HorizontalPodAutoscaler automatically scales based on metrics:

```bash
# Check HPA status
kubectl get hpa tiny-bitly-hpa -n tiny-bitly

# View HPA details
kubectl describe hpa tiny-bitly-hpa -n tiny-bitly
```

### Cluster Auto-Scaling (Autopilot)

With Autopilot, node scaling is fully automatic and managed by Google:

- Nodes scale up automatically when pods need resources
- Nodes scale down automatically when underutilized
- No configuration needed - Autopilot handles all node management
- You only need to configure pod resource requests/limits in your deployments

**Note:** With Standard clusters, you would configure `--min-nodes` and `--max-nodes` during cluster creation, but Autopilot handles this automatically.

## Updating the Application

### Rolling Update (Zero Downtime)

```bash
# Update container image
kubectl set image deployment/tiny-bitly \
  tiny-bitly=gcr.io/PROJECT_ID/tiny-bitly:v2.0.0 \
  -n tiny-bitly

# Watch rollout status
kubectl rollout status deployment/tiny-bitly -n tiny-bitly

# Rollback if needed
kubectl rollout undo deployment/tiny-bitly -n tiny-bitly
```

### Blue-Green Deployment

For zero-downtime deployments with instant rollback:

1. Deploy new version to separate deployment
2. Switch traffic using service selector
3. Keep old version running for instant rollback

## Troubleshooting

### Pods Not Starting

```bash
# Check pod status
kubectl get pods -n tiny-bitly

# Describe pod for events
kubectl describe pod POD_NAME -n tiny-bitly

# View pod logs
kubectl logs POD_NAME -n tiny-bitly

# Check previous container logs (if crashed)
kubectl logs POD_NAME --previous -n tiny-bitly
```

### Database Connection Issues

```bash
# Check Cloud SQL Proxy sidecar logs
kubectl logs POD_NAME -c cloud-sql-proxy -n tiny-bitly

# Verify Cloud SQL connection name
kubectl describe pod POD_NAME -n tiny-bitly | grep -A3 cloud-sql-proxy
```

### High Resource Usage

```bash
# Check resource usage
kubectl top pods -n tiny-bitly
kubectl top nodes

# Check HPA status
kubectl get hpa -n tiny-bitly
```

### Common Issues

1. **Image pull errors**: Verify image exists in GCR and credentials are correct
2. **CrashLoopBackOff**: Check pod logs for application errors
3. **Pending pods**: Check node resources and resource requests/limits
4. **Service not accessible**: Verify service type and ingress configuration

## Cost Estimates

### GKE Autopilot Cluster Costs

- **Control plane**: Free (managed by Google)
- **Pod resources**: Pay only for requested CPU/memory (not entire nodes)
  - Example: 2 pods with 500m CPU, 512Mi memory each: ~$30-40/month
  - Scales automatically - pay only for what you use
  - More cost-effective than Standard clusters for variable workloads
- **Load balancer**: ~$18/month (if using Ingress)
- **Persistent volumes**: ~$0.17/GB/month

### Managed Services (Same as VM deployment)

- **Cloud SQL db-f1-micro**: ~$7/month
- **Memorystore basic 1GB**: ~$30/month

### Total Estimated Cost

- **Minimum (2 pods)**: ~$85-95/month
- **With auto-scaling**: ~$85-150/month depending on load
- **vs Standard cluster**: Often cheaper due to pay-per-pod pricing
- **vs VM deployment**: Similar or lower cost, with better auto-scaling and reliability

**Note:** Autopilot pricing is based on requested pod resources (CPU/memory), not node capacity. This makes it more cost-effective for workloads that don't fully utilize node resources.

## Cleanup

To delete all resources:

```bash
# Delete Kubernetes resources
kubectl delete namespace tiny-bitly

# Delete GKE cluster
gcloud container clusters delete $CLUSTER_NAME --region=$REGION

# Delete Cloud SQL (⚠️ Deletes all data!)
gcloud sql instances delete tiny-bitly-db

# Delete Memorystore (⚠️ Deletes all data!)
gcloud redis instances delete tiny-bitly-redis --region=us-central1

# Delete container images (optional)
gcloud container images delete gcr.io/$PROJECT_ID/tiny-bitly:$IMAGE_TAG
```

## Next Steps

1. **CI/CD Pipeline**: Set up Cloud Build or GitHub Actions for automated deployments
2. **Production Hardening**: 
   - Use Secret Manager for secrets
   - Enable network policies
   - Set up backup strategies
   - Configure alerting
3. **Multi-Region**: Deploy to multiple regions for high availability
4. **Service Mesh**: Consider Istio for advanced traffic management
5. **GitOps**: Use ArgoCD or Flux for declarative deployments

## Additional Resources

- [GKE Documentation](https://cloud.google.com/kubernetes-engine/docs)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [GKE Best Practices](https://cloud.google.com/kubernetes-engine/docs/best-practices)
- [Cloud SQL Proxy](https://cloud.google.com/sql/docs/postgres/sql-proxy)
- [GKE Monitoring](https://cloud.google.com/kubernetes-engine/docs/how-to/monitoring)
