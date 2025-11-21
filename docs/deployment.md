# Deployment

This document describes how to deploy the Bible API Service to Google Cloud Run.

## Prerequisites

-   A Google Cloud Platform (GCP) project.
-   The `gcloud` CLI installed and configured.
-   A container registry (e.g., Google Artifact Registry).

## Building and Pushing the Image

1.  **Enable the Artifact Registry API:**
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

## Deploying to Cloud Run

Deploy the service using `gcloud`. Ensure all necessary environment variables are set.

```bash
gcloud run deploy bible-api-service \
    --image <gcp-region>-docker.pkg.dev/<gcp-project>/<repository-name>/bible-api-service \
    --platform managed \
    --region <gcp-region> \
    --allow-unauthenticated \
    --set-env-vars="API_KEY=<your-api-key>,GCP_PROJECT_ID=<your-project-id>,LLM_PROVIDERS=openai,gemini,OPENAI_API_KEY=<your-openai-key>,GEMINI_API_KEY=<your-gemini-key>"
```

**Note**: The `--allow-unauthenticated` flag makes the service publicly accessible URL-wise, but the application itself still enforces authentication via the `X-API-KEY` header.

## Environment Variables Reference

| Variable | Description | Required? |
| :--- | :--- | :--- |
| `API_KEY` | Secret key for client authentication. | **Yes** |
| `GCP_PROJECT_ID` | Google Cloud Project ID (for Secret Manager). | **Yes** (for Prod) |
| `LLM_PROVIDERS` | Comma-separated list of providers (e.g., `openai,gemini,deepseek`). | **Yes** (for LLM features) |
| `OPENAI_API_KEY` | API key for OpenAI. | If using OpenAI |
| `GEMINI_API_KEY` | API key for Google Gemini. | If using Gemini |
| `DEEPSEEK_API_KEY` | API key for DeepSeek. | If using DeepSeek |
| `OPENAI_CUSTOM_BASE_URL` | Base URL for custom OpenAI-compatible provider. | If using Custom |
