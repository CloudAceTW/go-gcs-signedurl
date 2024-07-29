# go-gcs-signedurl

This project demonstrates how to generate signed URLs for Google Cloud Storage (GCS) objects using Go. This allows you to share files stored in GCS without exposing the actual object's URL.

## Getting Started

### Prerequisites

* **Google Cloud Project:** You'll need a Google Cloud project with the following services enabled:
    * Google Cloud Storage (GCS)
    * Cloud Firestore
* **Google Cloud SDK:** Install the Google Cloud SDK and authenticate your account.
* **Go:** Ensure you have Go installed on your system.

### Setup

1. **Create a GCS Bucket:**
    * Create a new GCS bucket to store your image files.
    * Set appropriate permissions for the bucket to allow access for signed URL generation.

2. **Create a Firestore Database:**
    * Create a new Firestore database to store short URLs.
    * Define a collection to store the short URLs and their corresponding GCS object URLs.

3. **Configure Environment Variables:**
    * Copy the `.env.example` file to `.env`:
      ```bash
      cp .env.example .env
      ```
    * Update the `.env` file with your project's specific values:
      * `GCS_BUCKET_NAME`: The name of your GCS bucket.
      * `FIRESTORE_PROJECT_ID`: The ID of your Firestore project.
      * `FIRESTORE_COLLECTION_NAME`: The name of your Firestore collection.
      * `GOOGLE_APPLICATION_CREDENTIALS`: The path to your Google Cloud service account key file.

### Development

1. **Build for Development:**
  ```bash
  make dev-build
  ```

2. **Start Development Console:**
  ```bash
  make dev-console
  ```
3. **Run in Development Mode:**
  ```bash
  go run ./main.go
  ```

## Deployment to Production
1. Create a GKE Cluster:
    * Create a Google Kubernetes Engine (GKE) cluster in your Google Cloud project.
2. Configure Helm Chart:
    * Copy the values.yaml.example file to values.yaml:
      ```bash
      cp ./chart/values.yaml.example ./values.yaml
      ```
    * Update the values.yaml file with your project's specific values

3. Build and Push Container Image:
    * Use skaffold to build your container image and push it to Google Artifact Registry (GAR):
      ```bash
      skaffold build -p dev --file-output=artifacts.json -d <GAR_PATH> -t "<TAG>"
      ```
        * Replace <GAR_PATH> with the path to your GAR repository.
        * Replace <TAG> with the desired tag for your container image.
4. Deploy to GKE:
    * Use skaffold to deploy your application to your GKE cluster:
      ```bash
      skaffold deploy -p prod -a artifacts.json
      ```
