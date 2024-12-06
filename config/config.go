// Package config defines the environment variable and command-line flags
package config

import (
	"sync"
	"time"

	"github.com/companieshouse/gofigure"
)

var cfg *Config
var mtx sync.Mutex
var penaltyTypesMap map[string]PenaltyDetails

type PenaltyDetails struct {
	EReceivedAppId, EFilingDesc, EMsgType, PDesc, PDescId, PResourceKind, PProductType string
}

func init() {
	penaltyTypesMap = map[string]PenaltyDetails{
		"LP": {
			EReceivedAppId: "lfp-pay-api.late_filing_penalty_received_email",
			EFilingDesc:    "Late Filing Penalty",
			EMsgType:       "late_filing_penalty_received_email",
			PDesc:          "Late Filing Penalty",
			PDescId:        "late-filing-penalty",
			PResourceKind:  "late-filing-penalty#late-filing-penalty",
			PProductType:   "late-filing-penalty",
		},
		"PN": {
			EReceivedAppId: "cs.late_filing_penalty_received_email",
			EFilingDesc:    "C S Penalty",
			EMsgType:       "cs_received_email",
			PDesc:          "C S Penalty",
			PDescId:        "cs-filing-penalty",
			PResourceKind:  "cs-filing-penalty#cs-filing-penalty",
			PProductType:   "cs-filing-penalty",
		},
	}
}

// Config defines the configuration options for this service.
type Config struct {
	BindAddr                   string       `env:"BIND_ADDR"                      flag:"bind-addr"                       flagDesc:"Bind address"`
	E5APIURL                   string       `env:"E5_API_URL"                     flag:"e5-api-url"                      flagDesc:"Base URL for the E5 API"`
	E5Username                 string       `env:"E5_USERNAME"                    flag:"e5-username"                     flagDesc:"Username for the E5 API"`
	MongoDBURL                 string       `env:"MONGODB_URL"                    flag:"mongodb-url"                     flagDesc:"MongoDB server URL"`
	Database                   string       `env:"PPS_MONGODB_DATABASE"           flag:"mongodb-database"                flagDesc:"MongoDB database for data"`
	MongoCollection            string       `env:"PPS_MONGODB_COLLECTION"         flag:"mongodb-collection"              flagDesc:"The name of the mongodb collection"`
	BrokerAddr                 []string     `env:"KAFKA_BROKER_ADDR"              flag:"broker-addr"                     flagDesc:"Kafka broker address"`
	SchemaRegistryURL          string       `env:"SCHEMA_REGISTRY_URL"            flag:"schema-registry-url"             flagDesc:"Schema registry url"`
	CHSURL                     string       `env:"CHS_URL"                        flag:"chs-url"                         flagDesc:"CHS URL"`
	WeeklyMaintenanceStartTime string       `env:"WEEKLY_MAINTENANCE_START_TIME"  flag:"weekly-maintenance-start-time"   flagDesc:"The time of the day when Weekly E5 maintenance starts"`
	WeeklyMaintenanceEndTime   string       `env:"WEEKLY_MAINTENANCE_END_TIME"    flag:"weekly-maintenance-end-time"     flagDesc:"The time of the day when Weekly E5 maintenance ends"`
	WeeklyMaintenanceDay       time.Weekday `env:"WEEKLY_MAINTENANCE_DAY"         flag:"weekly-maintenance-day"          flagDesc:"The day on which Weekly E5 maintenance takes place"`
	PlannedMaintenanceStart    string       `env:"PLANNED_MAINTENANCE_START_TIME" flag:"planned-maintenance-start-time"  flagDesc:"The time of the day at which Planned E5 maintenance starts"`
	PlannedMaintenanceEnd      string       `env:"PLANNED_MAINTENANCE_END_TIME"   flag:"planned-maintenance-end-time"    flagDesc:"The time of the day at which Planned E5 maintenance ends"`
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

func GetMap() map[string]PenaltyDetails {
	copyMap := make(map[string]PenaltyDetails, len(penaltyTypesMap))
	for k, v := range penaltyTypesMap {
		copyMap[k] = v
	}
	return copyMap
}

func GetValue(key string) (PenaltyDetails, bool) {
	value, exists := penaltyTypesMap[key]
	return value, exists
}
