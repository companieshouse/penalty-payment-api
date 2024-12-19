package config

import (
	"os"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitLoadPenaltyDetails(t *testing.T) {
	Convey("File does not exist", t, func() {
		_, err := LoadPenaltyDetails("pen_details.yml")
		So(err.Error(), ShouldEqual, "open pen_details.yml: no such file or directory")
	})
	Convey("YAML in incorrect format", t, func() {
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

		So(err.Error(), ShouldEqual, "yaml: line 2: found character that cannot start any token")

	})
	Convey("Load Penalty PenaltyDetails", t, func() {
		testYaml := []byte(`
name: penalty details
details:
  LP:
    EmailReceivedAppId: "lfp-pay-api.late_filing_penalty_received_email"
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

		So(err, ShouldBeNil)
		So(penaltyDetailsMap.Name, ShouldEqual, "penalty details")
		So(penaltyDetailsMap.Details["LP"].EmailReceivedAppId, ShouldEqual, "lfp-pay-api.late_filing_penalty_received_email")
	})
}

func TestUnitGetAllowedTransactions(t *testing.T) {
	Convey("File does not exist", t, func() {
		_, err := GetAllowedTransactions("pen_types.yml")
		So(err.Error(), ShouldEqual, "open pen_types.yml: no such file or directory")
	})
	Convey("YAML in incorrect format", t, func() {
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

		So(err.Error(), ShouldEqual, "yaml: line 2: found character that cannot start any token")

	})
	Convey("Get allowed transactions", t, func() {
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

		So(err, ShouldBeNil)
		So(allowedTransactionsMap.Description, ShouldEqual, "transaction types and subtypes of allowed penalties")
		So(allowedTransactionsMap.Types["1"]["C1"], ShouldEqual, true)
	})
}

func TestUnitGet(t *testing.T) {
	Convey("same instance for multiple calls", t, func() {
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
}
