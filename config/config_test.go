package config

import (
	"encoding/json"
	"os"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api/common/utils"
	. "github.com/smartystreets/goconvey/convey"
)

// key constants
const (
	BindAddr                               = `BIND_ADDR`
	E5APIURL                               = `E5_API_URL`
	E5Username                             = `E5_USERNAME`
	MongoDBURL                             = `MONGODB_URL`
	Database                               = `PPS_MONGODB_DATABASE`
	PayableResourcesCollection             = `PPS_MONGODB_PAYABLE_RESOURCES_COLLECTION`
	AccountPenaltiesCollection             = `PPS_MONGODB_ACCOUNT_PENALTIES_COLLECTION`
	AccountPenaltiesTTL                    = `PPS_ACCOUNT_PENALTIES_TTL`
	BrokerAddr                             = `KAFKA_BROKER_ADDR`
	ZookeeperURL                           = `KAFKA_ZOOKEEPER_ADDR`
	Kafka3BrokerAddr                       = `KAFKA3_BROKER_ADDR`
	Kafka3ZookeeperURL                     = `KAFKA3_ZOOKEEPER_ADDR`
	SchemaRegistryURL                      = `SCHEMA_REGISTRY_URL`
	EmailSendTopic                         = `EMAIL_SEND_TOPIC`
	PenaltyPaymentsProcessingTopic         = `PENALTY_PAYMENTS_PROCESSING_TOPIC`
	PenaltyPaymentsProcessingMaxRetries    = `PENALTY_PAYMENTS_PROCESSING_MAX_RETRIES`
	PenaltyPaymentsProcessingRetryDelay    = `PENALTY_PAYMENTS_PROCESSING_RETRY_DELAY`
	PenaltyPaymentsProcessingRetryMaxDelay = `PENALTY_PAYMENTS_PROCESSING_RETRY_MAX_DELAY`
	ConsumerGroupName                      = `CONSUMER_GROUP_NAME`
	ConsumerRetryGroupName                 = `CONSUMER_RETRY_GROUP_NAME`
	ConsumerRetryThrottleRate              = `CONSUMER_RETRY_THROTTLE_RATE`
	ConsumerRetryMaxAttempts               = `CONSUMER_RETRY_MAX_ATTEMPTS`
	FeatureFlagPaymentsProcessingEnabled   = `FEATURE_FLAG_PAYMENTS_PROCESSING_ENABLED`
	CHSURL                                 = `CHS_URL`
	WeeklyMaintenanceStartTime             = `WEEKLY_MAINTENANCE_START_TIME`
	WeeklyMaintenanceEndTime               = `WEEKLY_MAINTENANCE_END_TIME`
	WeeklyMaintenanceDay                   = `WEEKLY_MAINTENANCE_DAY`
	PlannedMaintenanceStart                = `PLANNED_MAINTENANCE_START_TIME`
	PlannedMaintenanceEnd                  = `PLANNED_MAINTENANCE_END_TIME`
)

// value constants
const (
	bindAddrConst                               = `:1234`
	e5ApiUrlConst                               = `http://e5-finance.example.com`
	e5UsernameConst                             = `e5_username`
	mongoDbUrlConst                             = `localhost:12344`
	databaseConst                               = `penalties-db`
	payableResourcesCollectionConst             = `payable-resources-collection`
	accountPenaltiesCollectionConst             = `account-penalties-collection`
	accountPenaltiesTTLConst                    = `24h`
	brokerAddrConst                             = `kafka:9092`
	kafka3BrokerAddrConst                       = `kafka3:9092`
	SchemaRegistryURLConst                      = `http://schema.registry`
	EmailSendTopicConst                         = `email-send-topic`
	PenaltyPaymentsProcessingTopicConst         = `penalty-payments-processing-topic`
	PenaltyPaymentsProcessingMaxRetriesConst    = `3`
	PenaltyPaymentsProcessingRetryDelayConst    = `1`
	PenaltyPaymentsProcessingRetryMaxDelayConst = `5`
	ConsumerGroupNameConst                      = `penalty-payment-api-penalty-payments-processing`
	ConsumerRetryGroupNameConst                 = `penalty-payment-api-penalty-payments-processing-retry`
	ConsumerRetryThrottleRateConst              = `1`
	ConsumerRetryMaxAttemptsConst               = `3`
	FeatureFlagPaymentsProcessingEnabledConst   = `false`
	CHSURLConst                                 = `http://localhost:8080`
	WeeklyMaintenanceStartTimeConst             = `1900`
	WeeklyMaintenanceEndTimeConst               = `1930`
	WeeklyMaintenanceDayConst                   = `0`
	PlannedMaintenanceStartConst                = `30 Jan 25 17:00 GMT`
	PlannedMaintenanceEndConst                  = `30 Jan 25 18:00 GMT`
)

func TestUnitSensitiveConfig(t *testing.T) {
	os.Clearenv()
	var (
		err           error
		configuration *Config
		envVars       = map[string]string{
			BindAddr:                               bindAddrConst,
			E5APIURL:                               e5ApiUrlConst,
			E5Username:                             e5UsernameConst,
			MongoDBURL:                             mongoDbUrlConst,
			Database:                               databaseConst,
			PayableResourcesCollection:             payableResourcesCollectionConst,
			AccountPenaltiesCollection:             accountPenaltiesCollectionConst,
			AccountPenaltiesTTL:                    accountPenaltiesTTLConst,
			BrokerAddr:                             brokerAddrConst,
			Kafka3BrokerAddr:                       kafka3BrokerAddrConst,
			SchemaRegistryURL:                      SchemaRegistryURLConst,
			EmailSendTopic:                         EmailSendTopicConst,
			PenaltyPaymentsProcessingTopic:         PenaltyPaymentsProcessingTopicConst,
			PenaltyPaymentsProcessingMaxRetries:    PenaltyPaymentsProcessingMaxRetriesConst,
			PenaltyPaymentsProcessingRetryDelay:    PenaltyPaymentsProcessingRetryDelayConst,
			PenaltyPaymentsProcessingRetryMaxDelay: PenaltyPaymentsProcessingRetryMaxDelayConst,
			ConsumerGroupName:                      ConsumerGroupNameConst,
			ConsumerRetryGroupName:                 ConsumerRetryGroupNameConst,
			ConsumerRetryThrottleRate:              ConsumerRetryThrottleRateConst,
			ConsumerRetryMaxAttempts:               ConsumerRetryMaxAttemptsConst,
			FeatureFlagPaymentsProcessingEnabled:   FeatureFlagPaymentsProcessingEnabledConst,
			CHSURL:                                 CHSURLConst,
			WeeklyMaintenanceStartTime:             WeeklyMaintenanceStartTimeConst,
			WeeklyMaintenanceEndTime:               WeeklyMaintenanceEndTimeConst,
			WeeklyMaintenanceDay:                   WeeklyMaintenanceDayConst,
			PlannedMaintenanceStart:                PlannedMaintenanceStartConst,
			PlannedMaintenanceEnd:                  PlannedMaintenanceEndConst,
		}
		builtConfig = Config{
			BindAddr:                               bindAddrConst,
			E5APIURL:                               e5ApiUrlConst,
			E5Username:                             e5UsernameConst,
			MongoDBURL:                             mongoDbUrlConst,
			Database:                               databaseConst,
			PayableResourcesCollection:             payableResourcesCollectionConst,
			AccountPenaltiesCollection:             accountPenaltiesCollectionConst,
			AccountPenaltiesTTL:                    accountPenaltiesTTLConst,
			BrokerAddr:                             []string{brokerAddrConst},
			Kafka3BrokerAddr:                       []string{kafka3BrokerAddrConst},
			SchemaRegistryURL:                      SchemaRegistryURLConst,
			EmailSendTopic:                         EmailSendTopicConst,
			PenaltyPaymentsProcessingTopic:         PenaltyPaymentsProcessingTopicConst,
			PenaltyPaymentsProcessingMaxRetries:    PenaltyPaymentsProcessingMaxRetriesConst,
			PenaltyPaymentsProcessingRetryDelay:    PenaltyPaymentsProcessingRetryDelayConst,
			PenaltyPaymentsProcessingRetryMaxDelay: PenaltyPaymentsProcessingRetryMaxDelayConst,
			ConsumerGroupName:                      ConsumerGroupNameConst,
			ConsumerRetryGroupName:                 ConsumerRetryGroupNameConst,
			ConsumerRetryThrottleRate:              1,
			ConsumerRetryMaxAttempts:               3,
			FeatureFlagPaymentsProcessingEnabled:   false,
			CHSURL:                                 CHSURLConst,
			WeeklyMaintenanceStartTime:             WeeklyMaintenanceStartTimeConst,
			WeeklyMaintenanceEndTime:               WeeklyMaintenanceEndTimeConst,
			WeeklyMaintenanceDay:                   time.Sunday,
			PlannedMaintenanceStart:                PlannedMaintenanceStartConst,
			PlannedMaintenanceEnd:                  PlannedMaintenanceEndConst,
		}
		e5UsernameRegex = regexp.MustCompile(e5UsernameConst)
		mongoDbUrlRegex = regexp.MustCompile(mongoDbUrlConst)
	)

	// set test env variables
	for varName, varValue := range envVars {
		os.Setenv(varName, varValue)
		defer os.Unsetenv(varName)
	}

	Convey("Given an environment with no environment variables set", t, func() {

		Convey("Then configuration should be nil", func() {
			So(configuration, ShouldBeNil)
		})

		Convey("When the config values are retrieved", func() {

			Convey("Then there should be no error returned, and values are as expected", func() {
				configuration, err = Get()

				So(err, ShouldBeNil)
				So(configuration, ShouldResemble, &builtConfig)
			})

			Convey("The generated JSON string from configuration should not contain sensitive data", func() {
				jsonByte, err := json.Marshal(builtConfig)

				So(err, ShouldBeNil)
				So(e5UsernameRegex.Match(jsonByte), ShouldEqual, false)
				So(mongoDbUrlRegex.Match(jsonByte), ShouldEqual, false)
			})
		})
	})
}

func TestUnitLoadPenaltyDetails(t *testing.T) {
	Convey("Given the main method tries to load the penalty details yaml file", t, func() {
		Convey("When the file does not exist", func() {
			_, err := LoadPenaltyDetails("pen_details.yml")
			Convey("Then an error should be returned", func() {
				So(err.Error(), ShouldEqual, "open pen_details.yml: no such file or directory")
			})
		})
		Convey("When the yaml file is in an incorrect format", func() {
			testYaml := []byte(`
	name: penalty details
	` + "invalid_yaml")
			tmpFile, err := os.CreateTemp("", "config_*.yaml")
			if err != nil {
				t.Fatalf("Failed to create tmp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.Write(testYaml); err != nil {
				t.Fatalf("Failed to write tmp file: %v", err)
			}

			_, err = LoadPenaltyDetails(tmpFile.Name())

			Convey("Then an error should be returned", func() {
				So(err.Error(), ShouldEqual, "yaml: line 2: found character that cannot start any token")
			})

		})
		Convey("When the yaml file exists and is in the correct format", func() {
			testYaml := []byte(`
name: penalty details
details:
  LATE_FILING:
    EmailReceivedAppId: "penalty-payment-api.penalty_payment_received_email"
`)
			tmpFile, err := os.CreateTemp("", "config_*.yaml")
			if err != nil {
				t.Fatalf("Failed to create tmp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.Write(testYaml); err != nil {
				t.Fatalf("Failed to write tmp file: %v", err)
			}

			penaltyDetailsMap, err := LoadPenaltyDetails(tmpFile.Name())
			if err != nil {
				t.Fatalf("Expected no error but got: %v", err)
			}

			Convey("Then the penalty details should be returned", func() {
				So(err, ShouldBeNil)
				So(penaltyDetailsMap.Name, ShouldEqual, "penalty details")
				So(penaltyDetailsMap.Details[utils.LateFilingPenRef].EmailReceivedAppId, ShouldEqual, "penalty-payment-api.penalty_payment_received_email")
			})
		})
	})
}

func TestUnitGetAllowedTransactions(t *testing.T) {
	Convey("Given the main method tries to load the penalty types yaml file", t, func() {
		Convey("When the file does not exist", func() {
			_, err := GetAllowedTransactions("pen_types.yml")
			Convey("Then an error should be returned", func() {
				So(err.Error(), ShouldEqual, "open pen_types.yml: no such file or directory")
			})
		})
		Convey("When the yaml file is in an incorrect format", func() {
			testYaml := []byte(`
	name: penalty types
	` + "invalid_yaml")
			tmpFile, err := os.CreateTemp("", "config_*.yaml")
			if err != nil {
				t.Fatalf("Failed to create tmp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.Write(testYaml); err != nil {
				t.Fatalf("Failed to write tmp file: %v", err)
			}

			_, err = LoadPenaltyDetails(tmpFile.Name())

			Convey("Then an error should be returned", func() {
				So(err.Error(), ShouldEqual, "yaml: line 2: found character that cannot start any token")
			})

		})
		Convey("When the yaml file exists and is in the correct format", func() {
			testYaml := []byte(`
description: transaction types and subtypes of allowed penalties
allowed_transactions:
  1:
    C1:
      true
`)
			tmpFile, err := os.CreateTemp("", "config_*.yaml")
			if err != nil {
				t.Fatalf("Failed to create tmp file: %v", err)
			}
			defer os.Remove(tmpFile.Name())

			if _, err := tmpFile.Write(testYaml); err != nil {
				t.Fatalf("Failed to write tmp file: %v", err)
			}

			allowedTransactionsMap, err := GetAllowedTransactions(tmpFile.Name())
			if err != nil {
				t.Fatalf("Expected no error but got: %v", err)
			}

			Convey("Then the allowed transactions should be returned", func() {
				So(err, ShouldBeNil)
				So(allowedTransactionsMap.Description, ShouldEqual, "transaction types and subtypes of allowed penalties")
				So(allowedTransactionsMap.Types["1"]["C1"], ShouldEqual, true)
			})
		})
	})
}

func TestUnitGet(t *testing.T) {
	Convey("Given get config is called for multiple Go routines", t, func() {
		var wg sync.WaitGroup
		results := make(chan *Config, 10)

		// create multiple goroutines that will call the Get() function
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				cfg, err := Get()
				if err != nil {
					t.Errorf("Error getting config: %v", err)
				}
				results <- cfg
			}()
		}

		wg.Wait()
		close(results)

		Convey("Then the same instance should be returned for multiple calls", func() {
			var firstCfg *Config
			for cfg := range results {
				if firstCfg == nil {
					firstCfg = cfg
				} else {
					if firstCfg != cfg {
						t.Errorf("all instance are not the same")
					}
				}
			}
		})
	})
}
