// Package config defines the environment variable and command-line flags
package config

import (
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v2"

	"github.com/companieshouse/gofigure"
	"github.com/companieshouse/penalty-payment-api-core/finance_config"
)

var cfg *Config
var mtx sync.Mutex

// Config defines the configuration options for this service.
type Config struct {
	BindAddr                               string       `env:"BIND_ADDR"                                    flag:"bind-addr"                                flagDesc:"Bind address"`
	E5APIURL                               string       `env:"E5_API_URL"                                   flag:"e5-api-url"                               flagDesc:"Base URL for the E5 API"`
	E5Username                             string       `env:"E5_USERNAME"                                  flag:"e5-username"                              flagDesc:"Username for the E5 API" json:"-"`
	MongoDBURL                             string       `env:"MONGODB_URL"                                  flag:"mongodb-url"                              flagDesc:"MongoDB server URL" json:"-"`
	Database                               string       `env:"PPS_MONGODB_DATABASE"                         flag:"mongodb-database"                         flagDesc:"MongoDB database for data"`
	PayableResourcesCollection             string       `env:"PPS_MONGODB_PAYABLE_RESOURCES_COLLECTION"     flag:"mongodb-payable-resources-collection"     flagDesc:"The name of the mongodb payable resources collection"`
	AccountPenaltiesCollection             string       `env:"PPS_MONGODB_ACCOUNT_PENALTIES_COLLECTION"     flag:"mongodb-account-penalties-collection"     flagDesc:"The name of the mongodb account penalties collection"`
	AccountPenaltiesTTL                    string       `env:"PPS_ACCOUNT_PENALTIES_TTL"                    flag:"account-penalties-ttl"                    flagDesc:"The time to live for account penalties cache entry"`
	BrokerAddr                             []string     `env:"KAFKA_BROKER_ADDR"                            flag:"broker-addr"                              flagDesc:"Kafka broker address"`
	Kafka3BrokerAddr                       []string     `env:"KAFKA3_BROKER_ADDR"                           flag:"kafka3-broker-addr"                       flagDesc:"Kafka3 broker address"`
	SchemaRegistryURL                      string       `env:"SCHEMA_REGISTRY_URL"                          flag:"schema-registry-url"                      flagDesc:"Schema registry url"`
	EmailSendTopic                         string       `env:"EMAIL_SEND_TOPIC"                             flag:"email-send-topic"                         flagDesc:"Kafka topic to send emails"`
	PenaltyPaymentsProcessingTopic         string       `env:"PENALTY_PAYMENTS_PROCESSING_TOPIC"            flag:"penalty-payments-processing-topic"        flagDesc:"Penalty payments processing topic"`
	PenaltyPaymentsProcessingMaxRetries    string       `env:"PENALTY_PAYMENTS_PROCESSING_MAX_RETRIES"      flag:"penalty-payments-processing-max-retries"  flagDesc:"Penalty payments processing max retry attempts for transient errors"`
	PenaltyPaymentsProcessingRetryDelay    string       `env:"PENALTY_PAYMENTS_PROCESSING_RETRY_DELAY"      flag:"penalty-payments-processing-retry-delay"  flagDesc:"Penalty payments processing retry delay for transient errors"`
	PenaltyPaymentsProcessingRetryMaxDelay string       `env:"PENALTY_PAYMENTS_PROCESSING_RETRY_MAX_DELAY"  flag:"penalty-payments-processing-max-delay"    flagDesc:"Penalty payments processing max delay for a retry attempt for transient errors"`
	ConsumerGroupName                      string       `env:"CONSUMER_GROUP_NAME"                          flag:"consumer-group-name"                      flagDesc:"Consumer group name"`
	ConsumerRetryGroupName                 string       `env:"CONSUMER_RETRY_GROUP_NAME"                    flag:"consumer-retry-group-name"                flagDesc:"Consumer retry group name"`
	ConsumerRetryThrottleRate              int          `env:"CONSUMER_RETRY_THROTTLE_RATE"                 flag:"consumer-retry-throttle-rate"             flagDesc:"Consumer retry throttle rate in seconds for resilience"`
	ConsumerRetryMaxAttempts               int          `env:"CONSUMER_RETRY_MAX_ATTEMPTS"                  flag:"consumer-retry-max-attempts"              flagDesc:"Consumer retry max attempts for resilience"`
	FeatureFlagPaymentsProcessingEnabled   bool         `env:"FEATURE_FLAG_PAYMENTS_PROCESSING_ENABLED"     flag:"feature-flag-payments-processing-enabled" flagDesc:"If the payments processing Kafka implementation is enabled"`
	DisabledPenaltyTransactionSubtypes     string       `env:"DISABLED_PENALTY_TRANSACTION_SUBTYPES"        flag:"disabled-penalty-transaction-subtypes"    flagDesc:"Penalty transaction subtypes to be disabled"`
	CHSURL                                 string       `env:"CHS_URL"                                      flag:"chs-url"                                  flagDesc:"CHS URL"`
	WeeklyMaintenanceStartTime             string       `env:"WEEKLY_MAINTENANCE_START_TIME"                flag:"weekly-maintenance-start-time"            flagDesc:"The time of the day when Weekly E5 maintenance starts"`
	WeeklyMaintenanceEndTime               string       `env:"WEEKLY_MAINTENANCE_END_TIME"                  flag:"weekly-maintenance-end-time"              flagDesc:"The time of the day when Weekly E5 maintenance ends"`
	WeeklyMaintenanceDay                   time.Weekday `env:"WEEKLY_MAINTENANCE_DAY"                       flag:"weekly-maintenance-day"                   flagDesc:"The day on which Weekly E5 maintenance takes place"`
	PlannedMaintenanceStart                string       `env:"PLANNED_MAINTENANCE_START_TIME"               flag:"planned-maintenance-start-time"           flagDesc:"The time of the day at which Planned E5 maintenance starts"`
	PlannedMaintenanceEnd                  string       `env:"PLANNED_MAINTENANCE_END_TIME"                 flag:"planned-maintenance-end-time"             flagDesc:"The time of the day at which Planned E5 maintenance ends"`
}

// Namespace implements service.Config Namespace.
func (c *Config) Namespace() string {
	return "penalty-payment-api"
}

// PenaltyDetailsMap defines the struct to hold the map of penalty details.
type PenaltyDetailsMap struct {
	Name    string                    `yaml:"name"`
	Details map[string]PenaltyDetails `yaml:"details"`
}

// PenaltyDetails defines the struct to hold the penalty details.
type PenaltyDetails struct {
	Description        string `yaml:"Description"`
	DescriptionId      string `yaml:"DescriptionId"`
	ClassOfPayment     string `yaml:"ClassOfPayment"`
	ResourceKind       string `yaml:"ResourceKind"`
	ProductType        string `yaml:"ProductType"`
	EmailReceivedAppId string `yaml:"EmailReceivedAppId"`
	EmailMsgType       string `yaml:"EmailMsgType"`
}

// Get returns a pointer to a Config instance
// populated with values from environment or command-line flags
func Get() (*Config, error) {
	mtx.Lock()
	defer mtx.Unlock()

	if cfg != nil {
		return cfg, nil
	}

	cfg = &Config{}

	err := gofigure.Gofigure(cfg)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

type PenaltyConfig struct {
	PenaltyTypesConfig     []finance_config.FinancePenaltyTypeConfig
	PayablePenaltiesConfig []finance_config.FinancePayablePenaltyConfig
}

var penaltyConfig PenaltyConfig

func LoadPenaltyConfig() error {
	var financePenaltyTypesConfig finance_config.FinancePenaltyTypesConfig
	var financePayablePenaltiesConfig finance_config.FinancePayablePenaltiesConfig

	err := yaml.Unmarshal(finance_config.FinancePenaltyTypes, &financePenaltyTypesConfig)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(finance_config.FinancePayablePenalties, &financePayablePenaltiesConfig)
	if err != nil {
		return err
	}

	penaltyConfig.PenaltyTypesConfig = financePenaltyTypesConfig.FinancePenaltyTypes
	penaltyConfig.PayablePenaltiesConfig = financePayablePenaltiesConfig.FinancePayablePenalties

	return nil
}

func GetPayablePenaltiesConfig() []finance_config.FinancePayablePenaltyConfig {
	return penaltyConfig.PayablePenaltiesConfig
}

func GetPenaltyTypesConfig() []finance_config.FinancePenaltyTypeConfig {
	return penaltyConfig.PenaltyTypesConfig
}

func TransactionAllowed(transactionType string, transactionSubtype string) bool {
	for _, penaltyType := range GetPenaltyTypesConfig() {
		if !penaltyType.Disabled && penaltyType.TransactionType == transactionType && penaltyType.TransactionSubtype == transactionSubtype {
			return true
		}
	}

	return false
}

func GetReasonForPenalty(transactionSubtype string) string {
	reasonForPenalty := "Penalty" //Default reason for penalty
	for _, penaltyType := range GetPenaltyTypesConfig() {
		if penaltyType.TransactionSubtype == transactionSubtype {
			reasonForPenalty = penaltyType.Reason
		}
	}

	return reasonForPenalty
}

func PenaltyReferenceTypeEnabled(penaltyReferenceType string) bool {
	for _, payablePenalty := range GetPayablePenaltiesConfig() {
		penalty := payablePenalty.Penalty
		if penalty.ReferenceType == penaltyReferenceType && penaltyEnabled(penalty) {
			return true
		}
	}

	return false
}

func penaltyEnabled(penalty *finance_config.PenaltyConfig) bool {
	return penalty.EnabledTo.IsZero() || penalty.EnabledTo.After(time.Now())
}

func TransanctionSubTypeEnabled(transactionSubtype string) bool {
	for _, penaltyType := range GetPenaltyTypesConfig() {
		if penaltyType.TransactionSubtype == transactionSubtype && !penaltyType.Disabled {
			return true
		}
	}

	return false
}

func LoadPenaltyDetails(fileName string) (*PenaltyDetailsMap, error) {
	yamlFile, err := os.ReadFile(fileName)
	if err != nil {
		return nil, err
	}

	var penaltyDetailsMap PenaltyDetailsMap

	err = yaml.Unmarshal(yamlFile, &penaltyDetailsMap)
	if err != nil {
		return nil, err
	}

	return &penaltyDetailsMap, nil
}
