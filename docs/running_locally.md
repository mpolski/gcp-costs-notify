# Running the function locally

This function has been written following [Google Cloud Functions Frameworks](https://cloud.google.com/functions/docs/running/function-frameworks).
Function Frameworks are open-source libraries used within Cloud Functions to unmarshal incoming HTTP requests into language-specific function invocations. You can use these to convert your function into a locally-runnable HTTP service.
Further details on Functions Framework for Go can be find [here](https://github.com/GoogleCloudPlatform/functions-framework-go#quickstart-hello-world-on-your-local-machine).

**Clone the repo**

```
git clone https://github.com/mpolski/gcp-costs-notify.git
```

**Provide the details about BigQuery dataset and Google Chat Space URL**

```
export PROJECT_ID=
export DATASET=
export BQ_TABLE_NAME=
export LOCATION=
export GOOGLE_CHAT_URL=
```

**Specify the function to be invoked by the Framework by providing the FUNCTION_TARGET environment variable**

```
export FUNCTION_TARGET=query
```

**Edit the go.mod file** 

```
cd query
sed -i '1 c module query/query' go.mod 
```

NOTE: Remember to revert this change before deploying to GCP otherwise the deployment will fail with the Error as follows:
```
ERROR: (gcloud.functions.deploy) OperationError: code=3, message=Build failed: error extracting package name: 2023/02/01 17:16:58 Unable to extract package name and imports: unable to find Go package in /workspace/serverless_function_source_code.
exit status 1 [id:7a420ccf]; Error ID: ce87980d
```

To revert the change:
```
sed -i '1 c module .query/query' go.mod
```

**5. Authenticate with Application Default Credentials if needed**

```
gcloud auth application-default login
```

**6. Run the function**

```
go run cmd/main.go
```

The function will starts

```
> go run cmd/main.go
Serving function: query
```

**7. Call the function**

```
curl localhost:8080
```