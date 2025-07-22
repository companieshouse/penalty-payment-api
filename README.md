# PENALTY PAYMENT API

An API for retrieving penalties from the E5 finance system and recording / viewing penalty payments

## Requirements
In order to run this API locally you will need to install the following:

- [Go](https://golang.org/doc/install)
- [Git](https://git-scm.com/downloads)

## Getting Started
1. Clone this repository: `go get github.com/companieshouse/penalty-payment-api`
2. Build the executable: `make build`

## Configuration
| Variable                                      | Default | Description                                                                 | Config Location                                                          |
|:----------------------------------------------|:-------:|:----------------------------------------------------------------------------|--------------------------------------------------------------------------|
| `API_KEY`                                     |   `-`   | API Key to call payments API                                                | Terraform Vault - To update, please create platform request              |
| `E5_USERNAME`                                 |   `-`   | E5 API Username                                                             | Terraform Vault - To update, please create platform request              |
| `MONGODB_URL`                                 |   `-`   | The mongo db connection string                                              | Terraform Vault - To update, please create platform request              |
| `E5_API_URL`                                  |   `-`   | E5 API Address                                                              | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `PPS_MONGODB_DATABASE`                        |   `-`   | The database name to connect to e.g. `financial_penalties`                  | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `PPS_MONGODB_PAYABLE_RESOURCES_COLLECTION`    |   `-`   | The collection name e.g. `payable_resources`                                | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `PPS_MONGODB_ACCOUNT_PENALTIES_COLLECTION`    |   `-`   | The collection name e.g. `account_penalties`                                | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `PPS_ACCOUNT_PENALTIES_TTL`                   |   `-`   | Account penalties cache time to live  e.g. `24h`                            | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `KAFKA_BROKER_ADDR`                           |   `_`   | Kafka Broker Address                                                        | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `KAFKA_ZOOKEEPER_ADDR`                        |   `_`   | Kafka Zookeeper Address                                                     | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `SCHEMA_REGISTRY_URL`                         |   `_`   | Schema Registry URL                                                         | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `EMAIL_SEND_TOPIC`                            |   `_`   | Kafka topic to send emails e.g. email-send                                  | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `PENALTY_PAYMENTS_PROCESSING_TOPIC`           |   `_`   | Kafka topic to process penalty payments to e.g. penalty-payments-processing | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `PENALTY_PAYMENTS_PROCESSING_MAX_RETRIES`     |   `_`   | The max retry attempts for transient errors e.g. 3                          | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `PENALTY_PAYMENTS_PROCESSING_RETRY_DELAY`     |   `_`   | The delay in seconds between retry attempts for transient errors e.g. 1     | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `PENALTY_PAYMENTS_PROCESSING_RETRY_MAX_DELAY` |   `_`   | The maximum delay time in seconds between retries for transient errors      | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `FEATURE_FLAG_PAYMENTS_PROCESSING_ENABLED`    |   `_`   | If the payments processing Kafka implementation is enabled                  | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `FEATURE_FLAG_SANCTIONS_CS_DISABLED`          | `false` | Feature flag to turn Sanctions - Confirmation Statement penalty OFF         | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE)  |
| `FEATURE_FLAG_SANCTIONS_ROE_DISABLED`         | `false` | Feature flag to turn Sanctions - ROE penalty OFF                            | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE)  |
| `API_URL`                                     |   `_`   | The application endpoint for the API, for go-sdk-manager integration        | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `PAYMENTS_API_URL`                            |   `_`   | The base path for the payments API, for go-sdk-manager integration          | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `CHS_URL`                                     |   `_`   | CHS URL                                                                     | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `WEEKLY_MAINTENANCE_START_TIME`               |   `_`   | Start time of weekly maintenance e.g. `1900`                                | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `WEEKLY_MAINTENANCE_END_TIME`                 |   `_`   | End time of weekly maintenance e.g. `1930`                                  | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `WEEKLY_MAINTENANCE_DAY`                      |   `_`   | Day of weekly maintenance e.g. `0` (zero for Sunday)                        | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `PLANNED_MAINTENANCE_START_TIME`              |   `_`   | Start time and date of planned maintenance e.g. `30 Jan 25 17:00 GMT`       | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |
| `PLANNED_MAINTENANCE_END_TIME`                |   `_`   | End time and date of planned maintenance e.g. `30 Jan 25 18:00 GMT`         | ecs-service-configs-dev(CIDEV) / ecs-service-configs-prod (STAGING/LIVE) |

## Endpoints

| Method    | Path                                                                | Description                                                           |
|:----------|:--------------------------------------------------------------------|:----------------------------------------------------------------------|
| **GET**   | `/penalty-payment-api/healthcheck`                                  | Standard healthcheck endpoint                                         |
| **GET**   | `/penalty-payment-api/healthcheck/finance-system`                   | Healthcheck endpoint to check whether the finance system is available |
| **GET**   | `/company/{customer_code}/penalties/late-filing`                    | List the late filing penalties for a company                          |
| **GET**   | `/company/{customer_code}/penalties/{penalty_reference_type}`       | List the financial penalties                                          |
| **POST**  | `/company/{customer_code}/penalties/payable`                        | Create a payable penalty resource                                     |
| **GET**   | `/company/{customer_code}/penalties/payable/{payable_ref}`          | Get a payable resource                                                |
| **GET**   | `/company/{customer_code}/penalties/payable/{payable_ref}/payment`  | List the cost items related to the penalty resource                   |
| **PATCH** | `/company/{customer_code}/penalties/payable/{payable_ref}/payment`  | Mark the resource as paid                                             |

## External Finance Systems
The only external finance system currently supported is E5.

## Docker support

Pull image from ch-shared-services registry by running `docker pull 416670754337.dkr.ecr.eu-west-2.amazonaws.com/penalty-payment-api:latest` command or run the following steps to build image locally:

1. `export SSH_PRIVATE_KEY_PASSPHRASE='[your SSH key passhprase goes here]'` (optional, set only if SSH key is passphrase protected)
2. `DOCKER_BUILDKIT=0 docker build --build-arg SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa)" --build-arg SSH_PRIVATE_KEY_PASSPHRASE -t 416670754337.dkr.ecr.eu-west-2.amazonaws.com/penalty-payment-api:latest .`

## Vulnerability Checks

1. Run the vulnerability scanner: `make security-check-summary`
2. To see details of any CVEs found run: `make security-check`
