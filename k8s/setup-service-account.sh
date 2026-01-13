#!/bin/bash
# Setup script for GKE service account and IAM permissions
# Run this before deploying to GKE

set -e

PROJECT_ID=$(gcloud config get-value project)
GSA_NAME=tiny-bitly-gsa
KSA_NAME=tiny-bitly-sa
NAMESPACE=tiny-bitly

echo "Setting up service account for Cloud SQL access..."

# Create Google Service Account
gcloud iam service-accounts create $GSA_NAME \
  --display-name="Tiny Bitly GSA" \
  --project=$PROJECT_ID || echo "Service account may already exist"

# Grant Cloud SQL Client role
gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member="serviceAccount:${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com" \
  --role="roles/cloudsql.client"

# Get GKE cluster name and region
CLUSTER_NAME=$(gcloud container clusters list --format="value(name)" --limit=1)
REGION=$(gcloud container clusters list --format="value(location)" --limit=1)

if [ -z "$CLUSTER_NAME" ]; then
  echo "Error: No GKE cluster found. Please create a cluster first."
  exit 1
fi

echo "Using cluster: $CLUSTER_NAME in $REGION"

# Enable Workload Identity on the cluster (if not already enabled)
gcloud container clusters update $CLUSTER_NAME \
  --region=$REGION \
  --workload-pool=$PROJECT_ID.svc.id.goog || echo "Workload Identity may already be enabled"

# Create Kubernetes service account
kubectl create namespace $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -
kubectl create serviceaccount $KSA_NAME -n $NAMESPACE --dry-run=client -o yaml | kubectl apply -f -

# Allow the Kubernetes service account to impersonate the Google service account
gcloud iam service-accounts add-iam-policy-binding \
  --role roles/iam.workloadIdentityUser \
  --member "serviceAccount:${PROJECT_ID}.svc.id.goog[${NAMESPACE}/${KSA_NAME}]" \
  ${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com

# Annotate the Kubernetes service account
kubectl annotate serviceaccount $KSA_NAME \
  -n $NAMESPACE \
  iam.gke.io/gcp-service-account=${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com

echo "Service account setup complete!"
echo ""
echo "Update k8s/service-account.yaml with:"
echo "  PROJECT_ID: $PROJECT_ID"
echo "  GSA: ${GSA_NAME}@${PROJECT_ID}.iam.gserviceaccount.com"
