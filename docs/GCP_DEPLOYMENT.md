# GCP Deployment Guide

This guide explains how to deploy the tiny-bitly service to Google Cloud Platform for better load testing results using **managed services** (Cloud SQL and Memorystore).

## Architecture

- **Cloud SQL** (PostgreSQL) - Managed database ([official docs](https://cloud.google.com/sql/docs/postgres))
- **Memorystore** (Redis) - Managed Redis cache ([official docs](https://cloud.google.com/memorystore/docs/redis))
- **Compute Engine VMs** - For Go application servers ([official docs](https://cloud.google.com/compute/docs))
- **Cloud Load Balancer** (optional) - For production load balancing ([official docs](https://cloud.google.com/load-balancing/docs))

## Prerequisites

- GCP account with billing enabled
- `gcloud` CLI installed ([installation guide](https://cloud.google.com/sdk/docs/install))
- Basic familiarity with GCP concepts ([getting started guide](https://cloud.google.com/docs/get-started))

## Step-by-Step Setup

### 1. Create GCP Project and Enable APIs

```bash
# Authenticate with GCP
gcloud auth login
gcloud auth application-default login

# Create or select a project
gcloud projects create YOUR_PROJECT_ID --name="tiny-bitly" || \
gcloud config set project YOUR_PROJECT_ID

# Enable required APIs
gcloud services enable compute.googleapis.com
gcloud services enable sqladmin.googleapis.com
gcloud services enable redis.googleapis.com
```

**Official Tutorials:**
- [Creating and managing projects](https://cloud.google.com/resource-manager/docs/creating-managing-projects)
- [Enabling APIs](https://cloud.google.com/apis/docs/getting-started)

### 2. Create Cloud SQL PostgreSQL Instance

```bash
# Create PostgreSQL instance
# db-f1-micro is the smallest tier (~$7/month), upgrade for production
gcloud sql instances create tiny-bitly-db \
  --database-version=POSTGRES_16 \
  --tier=db-f1-micro \
  --region=us-central1 \
  --root-password=YOUR_SECURE_PASSWORD \
  --storage-type=SSD \
  --storage-size=20GB \
  --backup-start-time=03:00

# Create database
gcloud sql databases create tiny-bitly --instance=tiny-bitly-db

# Create application user
gcloud sql users create app_user \
  --instance=tiny-bitly-db \
  --password=YOUR_SECURE_PASSWORD
```

**Official Tutorials:**
- [Creating Cloud SQL instances](https://cloud.google.com/sql/docs/postgres/create-instance)
- [Connecting to Cloud SQL](https://cloud.google.com/sql/docs/postgres/connect-overview)

**Note:** For production, consider:
- Higher tier (e.g., `db-n1-standard-1` or higher)
- Read replicas for scaling reads
- Automated backups
- Private IP for better security

### 3. Create Memorystore Redis Instance

```bash
# Create Redis instance
# basic tier with 1GB is the smallest (~$30/month)
gcloud redis instances create tiny-bitly-redis \
  --size=1 \
  --region=us-central1 \
  --tier=basic \
  --redis-version=redis_7_0
```

**Official Tutorials:**
- [Creating Memorystore instances](https://cloud.google.com/memorystore/docs/redis/create-instance)
- [Connecting to Memorystore](https://cloud.google.com/memorystore/docs/redis/connect-redis-instance)

**Note:** For production, consider:
- Standard tier for high availability
- Larger memory size (5GB+)
- Redis Cluster for horizontal scaling

### 4. Create Compute Engine VM for Application

```bash
# Create VM for Go application servers
gcloud compute instances create tiny-bitly-app \
  --zone=us-central1-a \
  --machine-type=e2-standard-4 \
  --image-family=ubuntu-2204-lts \
  --image-project=ubuntu-os-cloud \
  --boot-disk-size=50GB \
  --boot-disk-type=pd-standard \
  --tags=http-server \
  --scopes=cloud-platform

# Allow HTTP traffic
gcloud compute firewall-rules create allow-http-8080 \
  --allow tcp:8080 \
  --source-ranges 0.0.0.0/0 \
  --target-tags http-server \
  --description="Allow HTTP traffic on port 8080"
```

**Official Tutorials:**
- [Creating VM instances](https://cloud.google.com/compute/docs/instances/create-start-instance)
- [Firewall rules](https://cloud.google.com/vpc/docs/firewalls)

### 5. Get Connection Details

```bash
# Get Cloud SQL connection details
CLOUD_SQL_CONNECTION_NAME=$(gcloud sql instances describe tiny-bitly-db \
  --format='get(connectionName)')
echo "Cloud SQL connection: $CLOUD_SQL_CONNECTION_NAME"

# Get Cloud SQL private IP (for VPC access)
CLOUD_SQL_IP=$(gcloud sql instances describe tiny-bitly-db \
  --format='get(ipAddresses[0].ipAddress)')
echo "Cloud SQL IP: $CLOUD_SQL_IP"

# Get Memorystore Redis IP
REDIS_IP=$(gcloud redis instances describe tiny-bitly-redis \
  --region=us-central1 \
  --format='get(host)')
echo "Redis IP: $REDIS_IP"

# Get VM external IP
VM_IP=$(gcloud compute instances describe tiny-bitly-app \
  --zone=us-central1-a \
  --format='get(networkInterfaces[0].accessConfigs[0].natIP)')
echo "VM external IP: $VM_IP"
```

### 6. Setup Application on VM

```bash
# SSH into the VM
gcloud compute ssh tiny-bitly-app --zone=us-central1-a

# On the VM, install dependencies
sudo apt-get update
sudo apt-get install -y docker.io docker-compose git golang-go

# Clone your repository
git clone YOUR_REPO_URL
cd tiny-bitly

# Install Cloud SQL Proxy (for secure connection to Cloud SQL)
# See: https://cloud.google.com/sql/docs/postgres/sql-proxy
wget https://storage.googleapis.com/cloud-sql-connectors/cloud-sql-proxy/v2.8.0/cloud-sql-proxy.linux.amd64 -O cloud-sql-proxy
chmod +x cloud-sql-proxy
sudo mv cloud-sql-proxy /usr/local/bin/
```

**Official Tutorials:**
- [Cloud SQL Proxy setup](https://cloud.google.com/sql/docs/postgres/sql-proxy)

### 7. Configure Environment Variables

Create `.env` file on the VM:

```bash
# Cloud SQL PostgreSQL (via Cloud SQL Proxy on localhost:5432)
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=tiny-bitly
POSTGRES_USER=app_user
POSTGRES_PASSWORD=YOUR_SECURE_PASSWORD

# Memorystore Redis (use private IP from step 5)
REDIS_HOST=REDIS_IP_FROM_STEP_5
REDIS_PORT=6379

# Application
API_PORT=8080
API_HOSTNAME=http://VM_IP_FROM_STEP_5:8080
LOG_LEVEL=info

# Rate limiting
RATE_LIMIT_REQUESTS_PER_SECOND=100
RATE_LIMIT_BURST=200

# Timeouts
TIMEOUT_REQUEST_MILLIS=10000
TIMEOUT_READ_MILLIS=5000
TIMEOUT_WRITE_MILLIS=5000
```

### 8. Start Cloud SQL Proxy

```bash
# On the VM, start Cloud SQL Proxy in background
# Replace CLOUD_SQL_CONNECTION_NAME with value from step 5
cloud-sql-proxy $CLOUD_SQL_CONNECTION_NAME &

# Verify connection
psql "host=localhost port=5432 dbname=tiny-bitly user=app_user password=YOUR_SECURE_PASSWORD" -c "SELECT 1;"
```

### 9. Run Database Migrations

```bash
# On the VM, run migrations
task migrate
```

### 10. Start Application

```bash
# On the VM, start the Go server
task start

# Or for multiple instances (horizontal scaling)
# ./scripts/start-servers.sh 4
```

### 11. Test Deployment

```bash
# From your laptop, test the health endpoint
curl http://VM_IP:8080/health

# Run load test
API_PORT=8080 go run ./cmd/test_load/main.go \
  -url=http://VM_IP:8080 \
  -users=1000 \
  -duration=60s \
  -read-only
```

## Alternative: Self-Managed Setup (Lower Cost)

If you prefer to run Postgres and Redis on the VM instead of managed services:

```bash
# On the VM, use docker-compose to run Postgres and Redis
docker-compose up -d postgres redis

# Update .env to use localhost
POSTGRES_HOST=localhost
POSTGRES_PORT=5434
REDIS_HOST=localhost
REDIS_PORT=6380
```

**Trade-offs:**
- ✅ Lower cost (~$120/month vs ~$150/month)
- ✅ More control over configuration
- ❌ You manage backups, updates, and scaling
- ❌ No automatic failover

## Load Testing from Your Laptop

```bash
# Update load test to point to GCP
API_PORT=8080 go run ./cmd/test_load/main.go \
  -url=http://EXTERNAL_IP:8080 \
  -users=10000 \
  -duration=60s \
  -user-interval=1s \
  -read-only
```

## Cost Estimates

### Managed Services (Recommended)
- **e2-standard-4 VM**: ~$0.15/hour = ~$110/month
- **Cloud SQL db-f1-micro**: ~$7/month ([pricing](https://cloud.google.com/sql/pricing))
- **Memorystore basic 1GB**: ~$30/month ([pricing](https://cloud.google.com/memorystore/pricing))
- **50GB disk**: ~$8/month
- **Network egress**: ~$0.12/GB (first 10GB free)
- **Total**: ~$155-180/month

### Self-Managed (Alternative)
- **e2-standard-4 VM**: ~$0.15/hour = ~$110/month
- **50GB disk**: ~$8/month
- **Network egress**: ~$0.12/GB (first 10GB free)
- **Total**: ~$120-150/month

**Note:** Use [GCP Pricing Calculator](https://cloud.google.com/products/calculator) for accurate estimates based on your usage.

## Performance Improvements Expected

1. **Better CPU**: Dedicated 4 vCPUs vs sharing laptop CPU
2. **More RAM**: 16GB vs laptop memory constraints
3. **Better Network**: GCP internal networking (low latency)
4. **No Resource Contention**: Services isolated from your laptop
5. **Higher Limits**: OS file descriptor limits much higher

Expected improvements:
- **10k users**: Should handle easily (vs 2.77% success on laptop)
- **Latency**: P50 should drop from 18s to <100ms
- **Throughput**: Should handle 10k+ RPS vs 630 RPS

## Monitoring

### GCP Console
- **VM metrics**: [Compute Engine Monitoring](https://console.cloud.google.com/compute/instances)
- **Cloud SQL metrics**: [Cloud SQL Monitoring](https://console.cloud.google.com/sql/instances)
- **Memorystore metrics**: [Memorystore Monitoring](https://console.cloud.google.com/memorystore/redis/instances)

### Command Line

```bash
# View VM metrics
gcloud compute instances describe tiny-bitly-app \
  --zone=us-central1-a \
  --format='get(status)'

# SSH and check resources
gcloud compute ssh tiny-bitly-app --zone=us-central1-a
htop
docker stats  # if using docker-compose
```

**Official Tutorials:**
- [Cloud Monitoring](https://cloud.google.com/monitoring/docs)
- [Setting up alerts](https://cloud.google.com/monitoring/alerts)

## Troubleshooting

### Cloud SQL Connection Issues
- **Problem**: Cannot connect to Cloud SQL
- **Solution**: Ensure Cloud SQL Proxy is running and connection name is correct
- **Docs**: [Troubleshooting Cloud SQL connections](https://cloud.google.com/sql/docs/postgres/troubleshooting)

### Memorystore Connection Issues
- **Problem**: Cannot connect to Redis
- **Solution**: Verify Redis IP, check firewall rules, ensure VM and Redis are in same region
- **Docs**: [Troubleshooting Memorystore](https://cloud.google.com/memorystore/docs/redis/troubleshooting)

### VM Performance Issues
- **Problem**: High CPU or memory usage
- **Solution**: Upgrade VM machine type, check application logs, monitor with `htop`
- **Docs**: [VM performance tuning](https://cloud.google.com/compute/docs/instances/optimizing-vm-performance)

## Cleanup

```bash
# Delete VM
gcloud compute instances delete tiny-bitly-app --zone=us-central1-a

# Delete Cloud SQL instance (⚠️ This deletes all data!)
gcloud sql instances delete tiny-bitly-db

# Delete Memorystore instance (⚠️ This deletes all data!)
gcloud redis instances delete tiny-bitly-redis --region=us-central1

# Delete firewall rule
gcloud compute firewall-rules delete allow-http-8080
```

**Warning:** Deleting Cloud SQL or Memorystore instances permanently deletes all data. Export backups first if needed.

## Additional Resources

- [GCP Documentation](https://cloud.google.com/docs)
- [Cloud SQL Best Practices](https://cloud.google.com/sql/docs/postgres/best-practices)
- [Memorystore Best Practices](https://cloud.google.com/memorystore/docs/redis/best-practices)
- [Compute Engine Best Practices](https://cloud.google.com/compute/docs/instances/best-practices)
- [GCP Architecture Center](https://cloud.google.com/architecture)
