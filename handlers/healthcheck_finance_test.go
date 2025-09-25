package handlers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitHandleHealthCheckFinance(t *testing.T) {

	cfg, _ := config.Get()
	now := time.Now()

	testCases := []struct {
		when                           string
		then                           string
		status                         int
		bodyStartsWith                 string
		weeklyDowntime                 bool
		plannedDowntime                bool
		plannedMaintenanceStartInvalid bool
		plannedMaintenanceEndInvalid   bool
	}{
		{
			when:           "When the system is healthy",
			then:           "Then the status should return 'OK'",
			status:         http.StatusOK,
			bodyStartsWith: `{"message":"HEALTHY"`,
			weeklyDowntime: false, plannedDowntime: false, plannedMaintenanceStartInvalid: false, plannedMaintenanceEndInvalid: false,
		},
		{
			when:           "When the system is not healthy due to weekly downtime",
			then:           "Then the status should return 'Service Unavailable'",
			status:         http.StatusServiceUnavailable,
			bodyStartsWith: `{"message":"UNHEALTHY - PLANNED MAINTENANCE","maintenance_end_time":`,
			weeklyDowntime: true, plannedDowntime: false, plannedMaintenanceStartInvalid: false, plannedMaintenanceEndInvalid: false,
		},
		{
			when:           "When the system is not healthy due to planned maintenance",
			then:           "Then the status should return 'Service Unavailable'",
			status:         http.StatusServiceUnavailable,
			bodyStartsWith: `{"message":"UNHEALTHY - PLANNED MAINTENANCE","maintenance_end_time":`,
			weeklyDowntime: false, plannedDowntime: true, plannedMaintenanceStartInvalid: false, plannedMaintenanceEndInvalid: false,
		},
		{
			when:           "When the Planned Maintenance Start Config value is invalid",
			then:           "Then the status should be 'Internal Server Error'",
			status:         http.StatusInternalServerError,
			weeklyDowntime: false, plannedDowntime: false, plannedMaintenanceStartInvalid: true, plannedMaintenanceEndInvalid: false,
		},
		{
			when:           "When the Planned Maintenance End Config value is invalid",
			then:           "Then the status should be 'Internal Server Error'",
			status:         http.StatusInternalServerError,
			bodyStartsWith: "",
			weeklyDowntime: false, plannedDowntime: false, plannedMaintenanceStartInvalid: false, plannedMaintenanceEndInvalid: true,
		},
	}

	Convey("Given I make a request to the healthcheck_finance endpoint", t, func() {

		for _, tc := range testCases {
			Convey(tc.when, func() {
				healthCheckFinanceTestConfigSetup(cfg, now, tc.weeklyDowntime, tc.plannedDowntime, tc.plannedMaintenanceStartInvalid, tc.plannedMaintenanceEndInvalid)
				req, _ := http.NewRequest("GET", "/penalty-payment-api/healthcheck/finance-system", nil)
				w := httptest.NewRecorder()
				HandleHealthCheckFinanceSystem(w, req)

				Convey(tc.then, func() {
					So(w.Code, ShouldEqual, tc.status)

					if tc.bodyStartsWith != "" {
						Convey("And the body of the message should be correct", func() {
							So(w.Body.String(), ShouldStartWith, tc.bodyStartsWith)
						})
					}
				})
			})
		}
	})
}

func healthCheckFinanceTestConfigSetup(cfg *config.Config, now time.Time, weeklyDowntime, plannedDowntime, plannedMaintenanceStartInvalid, plannedMaintenanceEndInvalid bool) {
	cfg.WeeklyMaintenanceStartTime = fmt.Sprintf("%02d00", now.Hour())
	cfg.WeeklyMaintenanceDay = now.Weekday()
	if weeklyDowntime {
		cfg.WeeklyMaintenanceEndTime = fmt.Sprintf("%02d00", now.Hour()+100)
	} else {
		cfg.WeeklyMaintenanceEndTime = fmt.Sprintf("%02d00", now.Hour()-100)
	}
	if plannedDowntime {
		cfg.PlannedMaintenanceStart = (now.AddDate(0, 0, -1)).Format("02 Jan 06 15:04 MST")
		cfg.PlannedMaintenanceEnd = (now.AddDate(0, 0, 1)).Format("02 Jan 06 15:04 MST")
	}
	if plannedMaintenanceStartInvalid {
		cfg.PlannedMaintenanceStart = "invalid"
	}
	if plannedMaintenanceEndInvalid {
		cfg.PlannedMaintenanceStart = (now).Format("02 Jan 06 15:04 MST")
		cfg.PlannedMaintenanceEnd = "invalid"
	}
}
