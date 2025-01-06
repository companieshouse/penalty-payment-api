package config

import (
	"os"
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

		So(err, ShouldBeNil)
		So(penaltyDetailsMap.Name, ShouldEqual, "penalty details")
		So(penaltyDetailsMap.Details["LP"].EmailReceivedAppId, ShouldEqual, "penalty-payment-api.penalty_payment_received_email")
	})
}
