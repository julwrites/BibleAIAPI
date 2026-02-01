# Staging Environment Setup Instructions

To enable the staging environment for the API service, the following changes are required in Google Cloud Platform.

## 1. Create Staging Cloud Run Service

Create a new Cloud Run service that mirrors the production configuration but serves as the staging target.

*   **Service Name:** `bible-api-service-staging`
*   **Region:** `asia-southeast1` (Must match the region in `.github/workflows/cicd.yml`)
*   **Authentication:** `Allow unauthenticated invocations` (Matches current production config in `cicd.yml`)
*   **Container Image:** You can deploy a placeholder image (e.g., `us-docker.pkg.dev/cloudrun/container/hello`) initially. The CI/CD pipeline will overwrite this on the first successful staging build.

## 2. Configure Runtime Service Account

The staging service must use the **same** User-Managed Service Account as the production service to ensure it has identical access to Google Secrets Manager and other resources.

1.  Identify the service account currently used by `bible-api-service`.
2.  Assign this service account to `bible-api-service-staging` during creation or by editing the service "Security" settings.

## 3. Configure Deployment Permissions

The GitHub Actions pipeline uses a specific Service Account (authenticated via the `GCP_SA_KEY` secret) to perform deployments. This identity needs permission to deploy to the new staging service.

1.  Identify the Service Account email used for CI/CD (this is the account associated with the JSON key in your GitHub `GCP_SA_KEY` secret).
2.  Grant the following roles to this account on the new `bible-api-service-staging` resource (or project-wide):
    *   **Cloud Run Developer** (`roles/run.developer`)
    *   **Service Account User** (`roles/iam.serviceAccountUser`) - This must be granted specifically on the *Runtime Service Account* identified in Step 2, to allow the deployer to "act as" the runtime account.

## 4. Artifact Registry

Ensure the existing Artifact Registry repository (defined in `GAR_REPOSITORY`, e.g., `bible-ai-api`) allows writing new images. No new repository is strictly needed as we will use the `bible-api-service-staging` image name within the existing repository, but ensure no retention policies prevent this.
