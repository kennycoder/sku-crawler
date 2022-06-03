# Intro

***sku-crawler*** is a demo boilerplate project written in go to scrape prices off HTML pages and send them to BigQuery for further processing (not in the scope of this repo). It is intended to run on [Cloud Run Jobs](https://cloud.google.com/run/docs/create-jobs "Cloud Run Jobs"), however can be used as a standalone too without any modifications

## Installation

1) Create a BigQuery dataset and a table with the following schema:

| Name        | Type           | Mode  |
| ------------- |:-------------:| -----:|
| name | STRING | REQUIRED |
| price | FLOAT | REQUIRED |
| source | STRING | NULLABLE |
| timestamp | TIMESTAMP | NULLABLE  |

2) Create a service account for the application that has the following permissions:

```
roles/bigquery.dataEditor
```

*optional* If you are planning on running this on [Cloud Run Jobs](http://https://cloud.google.com/run/docs/create-jobs "Cloud Run Jobs"), add the following permissions:
```
roles/artifactregistry.writer
roles/run.invoker
```


3) To build, you need docker

```bash
docker build --tag crawler .
```

## Usage locally

To run this locally you need to export the keys for the service account you created earlier.

```
export GOOGLE_APPLICATION_CREDENTIALS="/path/to/gcp/credentials.json"

docker run \
-e PROJECT_ID='project' \
-e BQ_DATASET='foo' \
-e BQ_TABLE='foo' \
-e GOOGLE_APPLICATION_CREDENTIALS=/tmp/credentials.json \
-v $GOOGLE_APPLICATION_CREDENTIALS:/tmp/credentials.json:ro \
crawler
```
## Usage with Cloud Run Jobs (*preview*)

First you need to [push your image to Artifact Registry](http://https://cloud.google.com/artifact-registry/docs/docker/pushing-and-pulling "push your image to Artifact Registry").

```bash
gcloud beta run jobs create crawl-job \
--image region-docker.pkg.dev/project/artifact-registry/crawler:latest \
--service-account=your-service-account@project.iam.gserviceaccount.com \
--set-env-vars=PROJECT_ID=foo,BQ_DATASET=foo,BQ_TABLE=foo

gcloud beta run jobs execute crawl-job
```
## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

## License
[MIT](https://choosealicense.com/licenses/mit/)