# PENALTY PAYMENT API

Penalty Payment Service (PPS) API which provides an interface for Creating, Getting, and Patching Penalties.

## Requirements
In order to run this API locally you will need to install the following:

- [Go](https://golang.org/doc/install)
- [Git](https://git-scm.com/downloads)

## Getting Started
1. Clone this repository: `go get github.com/companieshouse/penalty-payment-api`
1. Build the executable: `make build`

## Configuration
| Variable                                  | Default | Description                                                           |
|:------------------------------------------|:-------:|:----------------------------------------------------------------------|
| `E5_API_URL`                              |   `-`   | E5 API Address                                                        |
| `E5_USERNAME`                             |   `-`   | E5 API Username                                                       |
| `BIND_ADDR`                               |   `-`   | The host:port to bind to                                              |
| `MONGODB_URL`                             |   `-`   | The mongo db connection string                                        |
| `PPS_MONGODB_DATABASE`                    |   `-`   | The database name to connect to e.g. `late_filing_penalties`          |
| `PPS_MONGODB_COLLECTION`                  |   `-`   | The collection name e.g. `payable_resources`                          |
| `KAFKA_BROKER_ADDR`                       |   `_`   | Kafka Broker Address                                                  |
| `SCHEMA_REGISTRY_URL`                     |   `_`   | Schema Registry URL                                                   |
| `API_URL`                                 |   `_`   | The application endpoint for the API, for go-sdk-manager integration  |
| `PAYMENTS_API_URL`                        |   `_`   | The base path for the payments API, for go-sdk-manager integration    |
| `CHS_URL`                                 |   `_`   | CHS URL                                                               |
| `WEEKLY_MAINTENANCE_START_TIME`           |   `_`   | Start time of weekly maintenance e.g. `0700`                          |
| `WEEKLY_MAINTENANCE_END_TIME`             |   `_`   | End time of weekly maintenance e.g. `0730`                            |
| `WEEKLY_MAINTENANCE_DAY`                  |   `_`   | Day of weekly maintenance e.g. `0` (zero for Sunday)                  |
| `PLANNED_MAINTENANCE_START_TIME`          |   `_`   | Start time and date of planned maintenance e.g. `01 Jan 19 15:04 BST` |
| `PLANNED_MAINTENANCE_END_TIME`            |   `_`   | End time and date of planned maintenance e.g. `31 Jan 19 16:59 BST`   |

## Endpoints
| Method    | Path                                                                   | Description                                                           |
|:----------|:-----------------------------------------------------------------------|:----------------------------------------------------------------------|
| **GET**   | `/penalty-payment-api/healthcheck`                                     | Standard healthcheck endpoint                                         |
| **GET**   | `/penalty-payment-api/healthcheck/finance-system`                      | Healthcheck endpoint to check whether the finance system is available |
| **GET**   | `/company/{company_number}/penalties/late-filing`                      | List the Late Filing Penalties for a company                          |
| **POST**  | `/company/{company_number}/penalties/late-filing/payable`              | Create a payable penalty resource                                     |
| **GET**   | `/company/{company_number}/penalties/late-filing/payable/{id}`         | Get a payable resource                                                |
| **GET**   | `/company/{company_number}/penalties/late-filing/payable/{id}/payment` | List the cost items related to the penalty resource                   |
| **PATCH** | `/company/{company_number}/penalties/late-filing/payable/{id}/payment` | Mark the resource as paid                                             |

## External Finance Systems
The only external finance system currently supported is E5.

## Docker support

Pull image from ch-shared-services registry by running `docker pull 416670754337.dkr.ecr.eu-west-2.amazonaws.com/penalty-payment-api:latest` command or run the following steps to build image locally:

1. `export SSH_PRIVATE_KEY_PASSPHRASE='[your SSH key passhprase goes here]'` (optional, set only if SSH key is passphrase protected)
2. `DOCKER_BUILDKIT=0 docker build --build-arg SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa)" --build-arg SSH_PRIVATE_KEY_PASSPHRASE -t 416670754337.dkr.ecr.eu-west-2.amazonaws.com/penalty-payment-api:latest .`
