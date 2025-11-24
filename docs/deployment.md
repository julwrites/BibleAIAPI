# Deployment

This document describes how to deploy the Bible API Service to Google Cloud Run.

## Prerequisites

-   A Google Cloud Platform (GCP) project.
-   The `gcloud` CLI installed and configured.
-   A container registry (e.g., Google Artifact Registry).

## Building and Pushing the Image

1.  **Enable APIs:**
    ```bash
    gcloud services enable artifactregistry.googleapis.com firestore.googleapis.com
    ```

2.  **Create a Docker repository (if not exists):**
    ```bash
    gcloud artifacts repositories create <repository-name> \
        --repository-format=docker \
        --location=<gcp-region>
    ```

3.  **Configure Docker Auth:**
    ```bash
    gcloud auth configure-docker <gcp-region>-docker.pkg.dev
    ```

4.  **Build and Push:**
    ```bash
    docker build -t <gcp-region>-docker.pkg.dev/<gcp-project>/<repository-name>/bible-api-service .
    docker push <gcp-region>-docker.pkg.dev/<gcp-project>/<repository-name>/bible-api-service
    ```

## Database Setup (Required)

This service uses Google Cloud Firestore for API key management and rate limiting.

1.  **Create Firestore Database:**
    Go to the Google Cloud Console -> Firestore.
    Create a database in **Native Mode**.
    Select a location (ideally the same region as your Cloud Run service).

2.  **Permissions:**
    Ensure the Cloud Run service account has the `Cloud Datastore User` role.

## Secret Manager Setup

For security, sensitive keys are stored in Secret Manager.

1.  **Create Secrets:**
    ```bash
    # Legacy/Fallback API Key (Optional)
    gcloud secrets create API_KEY --replication-policy="automatic"
    echo -n "your-legacy-key" | gcloud secrets versions add API_KEY --data-file=-

    # Admin Password for Key Generation (Required)
    gcloud secrets create ADMIN_PASSWORD --replication-policy="automatic"
    echo -n "your-secure-admin-password" | gcloud secrets versions add ADMIN_PASSWORD --data-file=-
    ```

2.  **Grant Access to Cloud Run Service Account:**
    ```bash
    gcloud projects add-iam-policy-binding <your-project-id> \
        --member="serviceAccount:<your-cloud-run-sa-email>" \
        --role="roles/secretmanager.secretAccessor"
    ```

## Deploying to Cloud Run

Deploy the service using `gcloud`.

```bash
gcloud run deploy bible-api-service \
    --image <gcp-region>-docker.pkg.dev/<gcp-project>/<repository-name>/bible-api-service \
    --platform managed \
    --region <gcp-region> \
    --allow-unauthenticated \
    --max-instances 10 \
    --set-env-vars="GCP_PROJECT_ID=<your-project-id>,LLM_PROVIDERS=openai,gemini,OPENAI_API_KEY=<your-openai-key>,GEMINI_API_KEY=<your-gemini-key>"
```

## DDOS Protection & Rate Limiting

### Cloud Run Settings
-   **`--max-instances`**: Caps the maximum number of container instances. Setting this to a reasonable limit (e.g., `10`) prevents cost explosions during a DDOS attack.
-   **`--concurrency`**: Controls how many requests a single instance handles (default 80). Lowering this can improve isolation but increases instance count.

### Application Rate Limiting
The application uses Firestore to enforce daily rate limits per API Key.
-   **Pros**: Flexible, tracks usage per client, cheap for low/medium traffic.
-   **Costs**: Firestore charges ~$0.18 per 100k writes. Each API request = 1 read + 1 write.

### Advanced Protection: Google Cloud Armor
For high-volume services or strict DDOS mitigation, consider **Google Cloud Armor**.
-   **Cost**: Standard tier is ~$5-10/month per policy + $0.75 per million requests.
-   **Benefits**: Blocks attacks at the edge (before they reach Cloud Run), supports IP bans, Geo-blocking, and WAF rules.
-   **Setup**: Requires setting up a Global Load Balancer in front of Cloud Run.

## Environment Variables Reference

| Variable | Description | Required? |
| :--- | :--- | :--- |
| `GCP_PROJECT_ID` | Google Cloud Project ID (for Secrets/Firestore). | **Yes** |
| `API_KEY` | Legacy secret key. | Optional |
| `ADMIN_PASSWORD` | Password for `/admin` dashboard. | **Yes** (Secret) |
| `LLM_PROVIDERS` | Comma-separated list of providers. | **Yes** |
