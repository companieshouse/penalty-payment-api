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

Pull image from private CH registry by running `docker pull 169942020521.dkr.ecr.eu-west-2.amazonaws.com/local/penalty-payment-api:latest` command or run the following steps to build image locally:

1. `export SSH_PRIVATE_KEY_PASSPHRASE='[your SSH key passhprase goes here]'` (optional, set only if SSH key is passphrase protected)
2. `DOCKER_BUILDKIT=0 docker build --build-arg SSH_PRIVATE_KEY="$(cat ~/.ssh/id_rsa)" --build-arg SSH_PRIVATE_KEY_PASSPHRASE -t 169942020521.dkr.ecr.eu-west-2.amazonaws.com/local/penalty-payment-api:latest .`

## Terraform ECS

### What does this code do?

The code present in this repository is used to define and deploy a dockerised container in AWS ECS.
This is done by calling a [module](https://github.com/companieshouse/terraform-modules/tree/main/aws/ecs) from terraform-modules. Application specific attributes are injected and the service is then deployed using Terraform via the CICD platform 'Concourse'.


Application specific attributes | Value                                | Description
:---------|:-----------------------------------------------------------------------------|:-----------
**ECS Cluster**        | company-requests                                  | ECS cluster (stack) the service belongs to
**Load balancer**      | {env}-chs-internalapi                             | The load balancer that sits in front of the service
**Concourse pipeline**     |[Pipeline link](https://ci-platform.companieshouse.gov.uk/teams/team-development/pipelines/penalty-payment-api) <br> [Pipeline code](https://github.com/companieshouse/ci-pipelines/blob/master/pipelines/ssplatform/team-development/penalty-payment-api)                                  | Concourse pipeline link in shared services


### Contributing
- Please refer to the [ECS Development and Infrastructure Documentation](https://companieshouse.atlassian.net/wiki/spaces/DEVOPS/pages/4390649858/Copy+of+ECS+Development+and+Infrastructure+Documentation+Updated) for detailed information on the infrastructure being deployed.

### Testing
- Ensure the terraform runner local plan executes without issues. For information on terraform runners please see the [Terraform Runner Quickstart guide](https://companieshouse.atlassian.net/wiki/spaces/DEVOPS/pages/1694236886/Terraform+Runner+Quickstart).
- If you encounter any issues or have questions, reach out to the team on the **#platform** slack channel.

### Vault Configuration Updates
- Any secrets required for this service will be stored in Vault. For any updates to the Vault configuration, please consult with the **#platform** team and submit a workflow request.

### Useful Links
- [ECS service config dev repository](https://github.com/companieshouse/ecs-service-configs-dev)
- [ECS service config production repository](https://github.com/companieshouse/ecs-service-configs-production)