# gcp-costs-notify
Cloud Function to run GCP Billing notifications

## About this function 
This function queries a dataset in BigQuery that contains billing information exported from a billing account.
The query returns consumption on per project basis from the day before.
The results will be pushed on to Google Chat Space via a provided Webhook. 

## Prerequisites
#### 1. Export your billing data to BigQuery
Follow the guide on how to [Set up Cloud Billing data export to BigQuery](https://cloud.google.com/billing/docs/how-to/export-data-bigquery-setup) to expoert your billing information to BigQuery.

#### 2. Configure a webhook in Google Chat Space
To register the incoming webhook:
- Open Google Chat in a web browser
- Go to the space to which you want to add a webhook
- At the top, next to space title, click arrow_drop_down Down Arrow > The icon for manage webhooks Manage webhooks
- If this space already has other webhooks, click Add another. Otherwise, skip this step
- Enter the name you like
- For Avatar URL, enter https://developers.google.com/chat/images/chat-product-icon.png
- Click SAVE.
- Click content_copy Copy to copy the full webhook URL to use it in the following step

#### 3. Development environment
You need a [Go](https://go.dev/doc/install) development environment or use [Cloud Shell](https://cloud.google.com/shell/docs/launching-cloud-shell).

#### 4. Clone the repo
```
git clone https://github.com/mpolski/gcp-costs-notify.git
```


## Deploying the function

#### Running the function locally (optional)
Before deploying the function to GCP one may want to test it locally.
See [instructions](https://github.com/mpolski/gcp-costs-notify/blob/main/docs/running_locally.md) on how to run the function locally.

#### Deploy to GCP
Enable Google Cloud Functions API in your project:

```
gcloud services enable cloudfunctions.googleapis.com
```

Fill in the `./query/.env.yaml` file with your values of the project your BigQuery dataset is in, the BigQuery dataset details (both from step 1) as weel as your Google Chat Space webhook link (created in step 2 above):

```
PROJECT_ID: <project_name_where_BQ_dataset_lives>
DATASET: <dataset name>
BQ_TABLE_NAME: <table name>
LOCATION: <region of the dataset>
GOOGLE_CHAT_URL: <webhook URL of your Google Chat space>
EOF
```

Deploy the function:

```
REGION=<region_name>
FUNCTION=<function_name>

cd query

gcloud functions deploy $FUNCTION \
--region=$REGION \
--entry-point=query \
--runtime go116 \
--trigger-http \
--max-instances=1 \
--env-vars-file=.env.yaml
```
When asked about allowing unauthenticated invocations, answer `No`.

#### Verify the function has been deployed

```
gcloud functions list
```

Call the function to test it

```
gcloud functions call $FUNCTION --region=$REGION
```

At this point a message should appear in you Google Chat Space similar to the image below:

![alt image](https://github.com/mpolski/gcp-costs-notify/blob/main/images/example.png?raw=true)


To run the function regularly one may use Google Cloud Scheduler.

## Scheduling the function

#### Create a Service Account
First, create a Service Account and assign it a role so that it can invoke the Cloud Function.

```
PROJECT_ID=$(gcloud config list --format='value(core.project)')
SA=queryscheduler
SA_EMAIL=$SA@$PROJECT_ID.iam.gserviceaccount.com

gcloud iam service-accounts create $SA \
  --description="Scheduler Service Account to trigger Cloud Function" \
  --display-name="Scheduler Service Account for Cloud Function"

gcloud projects add-iam-policy-binding $PROJECT_ID \
  --member serviceAccount:$SA_EMAIL \
  --role roles/cloudfunctions.invoker
```
#### Create the Scheduler Job
Next, create the Scheduler Job to trigger the function at a desired intervals.

```
JOB=run-$FUNCTION
TZ="Europe/Warsaw"
URI=$(gcloud functions describe $FUNCTION --region $REGION --format="value(httpsTrigger.url)")

gcloud scheduler jobs create http $JOB \
  --schedule="0 8 * * *" \
  --time-zone="Europe/Warsaw" \
  --uri=$URI \
  --oidc-service-account-email=$SA_EMAIL
```
Run the job now to test it.

```
gcloud scheduler jobs run $JOB
```

## Cleaning up

```
gcloud scheduler jobs delete $JOB --quiet
gcloud iam service-accounts delete $SA_EMAIL --quiet
gcloud functions delete $FUNCTION --region=$REGION --quiet
```

### License
Released under the [Apache license](https://github.com/mpolski/gcp-costs-notify/blob/main/LICENSE.md). It is distributed as-is, without warranties or conditions of any kind.
