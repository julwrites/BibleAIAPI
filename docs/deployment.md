# Deployment

This document describes how to deploy the Bible API Service.

## Target Environment

The service is designed to be deployed as a containerized application on Google Cloud Run.

## Prerequisites

-   A Google Cloud Platform (GCP) project.
-   The `gcloud` CLI installed and configured.
-   A container registry (e.g., Google Artifact Registry).

## Building and Pushing the Image

1.  **Enable the Artifact Registry API:**
    ```bash
    gcloud services enable artifactregistry.googleapis.com
    ```

2.  **Create a Docker repository:**
    ```bash
    gcloud artifacts repositories create <repository-name> \
        --repository-format=docker \
        --location=<gcp-region>
    ```

3.  **Build the Docker image:**
    ```bash
    docker build -t <gcp-region>-docker.pkg.dev/<gcp-project>/<repository-name>/bible-api-service .
    ```

4.  **Push the image to the registry:**
    ```bash
    docker push <gcp-region>-docker.pkg.dev/<gcp-project>/<repository-name>/bible-api-service
    ```

## Deploying to Cloud Run

To deploy the service to Cloud Run, use the following `gcloud` command:

```bash
gcloud run deploy bible-api-service \
    --image <gcp-region>-docker.pkg.dev/<gcp-project>/<repository-name>/bible-api-service \
    --platform managed \
    --region <gcp-region> \
    --allow-unauthenticated \
    --set-env-vars="API_KEY=<your-api-key>,LLM_PROVIDERS=openai,gemini,OPENAI_API_KEY=<your-openai-key>,GEMINI_API_KEY=<your-gemini-key>,GCP_PROJECT_ID=<your-project-id>"
```

**Note**: The `--allow-unauthenticated` flag makes the service publicly accessible, but it is still protected by the API key authentication implemented in the application.

## Environment Variables

The following environment variables need to be configured for the deployment:

-   `API_KEY`: The secret API key for client authentication.
-   `GCP_PROJECT_ID`: The Google Cloud Project ID (required for secrets management).
-   `LLM_PROVIDERS`: A comma-separated list of LLM providers to use (e.g., "openai,gemini,deepseek").
-   `OPENAI_API_KEY`: The API key for the OpenAI service (required if using "openai").
-   `GEMINI_API_KEY`: The API key for the Google Gemini service (required if using "gemini").
-   `DEEPSEEK_API_KEY`: The API key for the DeepSeek service (required if using "deepseek").
-   `OPENAI_CUSTOM_BASE_URL`: The base URL for a custom OpenAI-compatible provider (required if using "openai-custom").
