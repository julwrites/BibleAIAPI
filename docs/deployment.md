# Deployment

This document describes how to deploy the Bible API Service to Google Cloud Run.

## Prerequisites

-   A Google Cloud Platform (GCP) project.
-   The `gcloud` CLI installed and configured.
-   A container registry (e.g., Google Artifact Registry).

## Building and Pushing the Image

1.  **Enable APIs:**
    ```bash
    gcloud services enable artifactregistry.googleapis.com
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

## Secret Manager Setup

For security, sensitive keys are stored in Secret Manager.

1.  **Create Secrets:**
    ```bash
    # API Keys (Required for Auth Middleware)
    # Stores a JSON mapping of ClientID to API Key
    gcloud secrets create API_KEYS --replication-policy="automatic"
    echo -n '{"telegram_bot": "secret-123", "web_portal": "secret-456"}' | gcloud secrets versions add API_KEYS --data-file=-

    # LLM Provider Keys (As needed)
    gcloud secrets create OPENAI_API_KEY --replication-policy="automatic"
    echo -n "sk-..." | gcloud secrets versions add OPENAI_API_KEY --data-file=-

    gcloud secrets create GEMINI_API_KEY --replication-policy="automatic"
    echo -n "..." | gcloud secrets versions add GEMINI_API_KEY --data-file=-
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

*Note: The `--allow-unauthenticated` flag is used because the service implements its own API Key authentication via `X-API-KEY` header.*

## DDOS Protection & Rate Limiting

### Cloud Run Settings
-   **`--max-instances`**: Caps the maximum number of container instances. Setting this to a reasonable limit (e.g., `10`) prevents cost explosions during a DDOS attack.
-   **`--concurrency`**: Controls how many requests a single instance handles (default 80). Lowering this can improve isolation but increases instance count.

### Advanced Protection: Google Cloud Armor
For high-volume services or strict DDOS mitigation, consider **Google Cloud Armor**.
-   **Cost**: Standard tier is ~$5-10/month per policy + $0.75 per million requests.
-   **Benefits**: Blocks attacks at the edge (before they reach Cloud Run), supports IP bans, Geo-blocking, and WAF rules.
-   **Setup**: Requires setting up a Global Load Balancer in front of Cloud Run.

## Environment Variables Reference

| Variable | Description | Required? |
| :--- | :--- | :--- |
| `GCP_PROJECT_ID` | Google Cloud Project ID (for Secrets). | **Yes** |
| `API_KEYS` | JSON string of client keys (if not using Secret Manager). | Optional (Fallback) |
| `LLM_PROVIDERS` | Comma-separated list of providers. | **Yes** |
| `OPENAI_API_KEY` | API Key for OpenAI. | Optional (if using OpenAI) |
| `GEMINI_API_KEY` | API Key for Gemini. | Optional (if using Gemini) |
| `DEEPSEEK_API_KEY` | API Key for DeepSeek. | Optional (if using DeepSeek) |
| `BIBLE_PROVIDER` | Bible data source (`biblegateway`, `biblehub`, `biblenow`). Default: `biblegateway` | Optional |
