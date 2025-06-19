package api

import (
	"fmt"
	"testing"
	"time"

	"github.com/companieshouse/penalty-payment-api/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestUnitCheckScheduledMaintenance(t *testing.T) {
	cfg, _ := config.Get()
	currentTime := time.Now()

	Convey("No maintenance config", t, func() {
		// Given

		// When
		gotSystemAvailableTime, gotSystemUnavailable, gotParseError := CheckScheduledMaintenance()

		// Then
		So(gotSystemAvailableTime, ShouldEqual, time.Time{})
		So(gotSystemUnavailable, ShouldBeFalse)
		So(gotParseError, ShouldBeFalse)
	})

	Convey("Current time is before weekly maintenance times", t, func() {
		// Given
		startHour := currentTime.Hour() - 2
		endHour := startHour + 1
		cfg.WeeklyMaintenanceStartTime = fmt.Sprintf("%02d00", startHour)
		cfg.WeeklyMaintenanceEndTime = fmt.Sprintf("%02d00", endHour)
		cfg.WeeklyMaintenanceDay = currentTime.Weekday()

		// When
		gotSystemAvailableTime, gotSystemUnavailable, gotParseError := CheckScheduledMaintenance()

		// Then
		So(gotSystemAvailableTime, ShouldEqual, time.Time{})
		So(gotSystemUnavailable, ShouldBeFalse)
		So(gotParseError, ShouldBeFalse)
	})

	Convey("Current time is during weekly maintenance times", t, func() {
		// Given
		endHour := currentTime.Hour() + 1
		cfg.WeeklyMaintenanceStartTime = fmt.Sprintf("%02d00", currentTime.Hour())
		cfg.WeeklyMaintenanceEndTime = fmt.Sprintf("%02d00", endHour)
		cfg.WeeklyMaintenanceDay = currentTime.Weekday()

		// When
		gotSystemAvailableTime, gotSystemUnavailable, gotParseError := CheckScheduledMaintenance()

		// Then
		So(gotSystemAvailableTime, ShouldEqual, time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), endHour, 0, 0, 0, currentTime.Location()))
		So(gotSystemUnavailable, ShouldBeTrue)
		So(gotParseError, ShouldBeFalse)
	})

	Convey("Current time is after weekly maintenance times", t, func() {
		// Given
		startHour := currentTime.Hour() + 1
		endHour := currentTime.Hour() + 2
		cfg.WeeklyMaintenanceStartTime = fmt.Sprintf("%02d00", startHour)
		cfg.WeeklyMaintenanceEndTime = fmt.Sprintf("%02d00", endHour)
		cfg.WeeklyMaintenanceDay = currentTime.Weekday()

		// When
		gotSystemAvailableTime, gotSystemUnavailable, gotParseError := CheckScheduledMaintenance()

		// Then
		So(gotSystemAvailableTime, ShouldEqual, time.Time{})
		So(gotSystemUnavailable, ShouldBeFalse)
		So(gotParseError, ShouldBeFalse)
	})

	Convey("Planned maintenance times in wrong format", t, func() {
		// Given
		cfg.WeeklyMaintenanceStartTime = "1900"
		cfg.WeeklyMaintenanceEndTime = "1930"
		cfg.WeeklyMaintenanceDay = time.Sunday
		startTime := currentTime.Add(time.Hour * -4).Truncate(time.Minute).Round(0)
		cfg.PlannedMaintenanceStart = startTime.Format(time.RFC3339)
		endTime := currentTime.Add(time.Hour * -3).Truncate(time.Minute).Round(0)
		cfg.PlannedMaintenanceEnd = endTime.Format(time.RFC3339)

		// When
		gotSystemAvailableTime, gotSystemUnavailable, gotParseError := CheckScheduledMaintenance()

		// Then
		So(gotSystemAvailableTime, ShouldEqual, time.Time{})
		So(gotSystemUnavailable, ShouldBeFalse)
		So(gotParseError, ShouldBeTrue)
	})

	Convey("Planned maintenance end time is in wrong format", t, func() {
		// Given
		weeklyEndHour := currentTime.Hour() + 2
		cfg.WeeklyMaintenanceStartTime = fmt.Sprintf("%02d00", currentTime.Hour())
		cfg.WeeklyMaintenanceEndTime = fmt.Sprintf("%02d00", weeklyEndHour)
		cfg.WeeklyMaintenanceDay = currentTime.Weekday()
		plannedStartTime := currentTime.Add(time.Hour * -1).Truncate(time.Minute).Round(0)
		cfg.PlannedMaintenanceStart = plannedStartTime.Format(time.RFC822)
		cfg.PlannedMaintenanceEnd = "1111111111"

		// When
		gotSystemAvailableTime, gotSystemUnavailable, gotParseError := CheckScheduledMaintenance()

		// Then
		So(gotSystemAvailableTime, ShouldEqual, time.Time{})
		So(gotSystemUnavailable, ShouldBeFalse)
		So(gotParseError, ShouldBeTrue)
	})

	Convey("Current time is before planned maintenance times", t, func() {
		// Given
		cfg.WeeklyMaintenanceStartTime = "1900"
		cfg.WeeklyMaintenanceEndTime = "1930"
		cfg.WeeklyMaintenanceDay = time.Sunday
		startTime := currentTime.Add(time.Hour * 3).Truncate(time.Minute).Round(0)
		cfg.PlannedMaintenanceStart = startTime.Format(time.RFC822)
		endTime := currentTime.Add(time.Hour * 4).Truncate(time.Minute).Round(0)
		cfg.PlannedMaintenanceEnd = endTime.Format(time.RFC822)

		// When
		gotSystemAvailableTime, gotSystemUnavailable, gotParseError := CheckScheduledMaintenance()

		// Then
		So(gotSystemAvailableTime, ShouldEqual, time.Time{})
		So(gotSystemUnavailable, ShouldBeFalse)
		So(gotParseError, ShouldBeFalse)
	})

	Convey("Current time is during planned maintenance times", t, func() {
		// Given
		cfg.WeeklyMaintenanceStartTime = "1900"
		cfg.WeeklyMaintenanceEndTime = "1930"
		cfg.WeeklyMaintenanceDay = time.Sunday
		startTime := currentTime.Add(time.Minute * -30).Truncate(time.Minute).Round(0)
		cfg.PlannedMaintenanceStart = startTime.Format(time.RFC822)
		endTime := currentTime.Add(time.Minute * 30).Truncate(time.Minute).Round(0)
		cfg.PlannedMaintenanceEnd = endTime.Format(time.RFC822)

		// When
		gotSystemAvailableTime, gotSystemUnavailable, gotParseError := CheckScheduledMaintenance()

		// Then
		So(gotSystemAvailableTime, ShouldEqual, endTime)
		So(gotSystemUnavailable, ShouldBeTrue)
		So(gotParseError, ShouldBeFalse)
	})

	Convey("Current time is after planned maintenance times", t, func() {
		// Given
		cfg.WeeklyMaintenanceStartTime = "1900"
		cfg.WeeklyMaintenanceEndTime = "1930"
		cfg.WeeklyMaintenanceDay = time.Sunday
		startTime := currentTime.Add(time.Hour * -4).Truncate(time.Minute).Round(0)
		cfg.PlannedMaintenanceStart = startTime.Format(time.RFC822)
		endTime := currentTime.Add(time.Hour * -3).Truncate(time.Minute).Round(0)
		cfg.PlannedMaintenanceEnd = endTime.Format(time.RFC822)

		// When
		gotSystemAvailableTime, gotSystemUnavailable, gotParseError := CheckScheduledMaintenance()

		// Then
		So(gotSystemAvailableTime, ShouldEqual, time.Time{})
		So(gotSystemUnavailable, ShouldBeFalse)
		So(gotParseError, ShouldBeFalse)
	})

	Convey("Current time is during scheduled maintenance times, planned ends later", t, func() {
		// Given
		weeklyEndHour := currentTime.Hour() + 1
		cfg.WeeklyMaintenanceStartTime = fmt.Sprintf("%02d00", currentTime.Hour())
		cfg.WeeklyMaintenanceEndTime = fmt.Sprintf("%02d00", weeklyEndHour)
		cfg.WeeklyMaintenanceDay = currentTime.Weekday()
		plannedStartTime := currentTime.Add(time.Minute * -1).Truncate(time.Minute).Round(0)
		cfg.PlannedMaintenanceStart = plannedStartTime.Format(time.RFC822)
		plannedEndTime := currentTime.Add(time.Hour * 2).Truncate(time.Minute).Round(0)
		cfg.PlannedMaintenanceEnd = plannedEndTime.Format(time.RFC822)

		// When
		gotSystemAvailableTime, gotSystemUnavailable, gotParseError := CheckScheduledMaintenance()

		// Then
		So(gotSystemAvailableTime, ShouldEqual, plannedEndTime)
		So(gotSystemUnavailable, ShouldBeTrue)
		So(gotParseError, ShouldBeFalse)
	})

	Convey("Current time is during scheduled maintenance times, planned ends earlier", t, func() {
		// Given
		weeklyEndHour := currentTime.Hour() + 2
		cfg.WeeklyMaintenanceStartTime = fmt.Sprintf("%02d00", currentTime.Hour())
		cfg.WeeklyMaintenanceEndTime = fmt.Sprintf("%02d00", weeklyEndHour)
		cfg.WeeklyMaintenanceDay = currentTime.Weekday()
		plannedStartTime := currentTime.Add(time.Hour * -1).Truncate(time.Minute).Round(0)
		cfg.PlannedMaintenanceStart = plannedStartTime.Format(time.RFC822)
		plannedEndTime := currentTime.Add(time.Hour * 1).Truncate(time.Minute).Round(0)
		cfg.PlannedMaintenanceEnd = plannedEndTime.Format(time.RFC822)

		// When
		gotSystemAvailableTime, gotSystemUnavailable, gotParseError := CheckScheduledMaintenance()

		// Then
		So(gotSystemAvailableTime, ShouldEqual, time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), weeklyEndHour, 0, 0, 0, currentTime.Location()))
		So(gotSystemUnavailable, ShouldBeTrue)
		So(gotParseError, ShouldBeFalse)
	})
}

func TestUnit_checkWeeklyDownTime(t *testing.T) {
	Convey("Check weekly down time", t, func() {
		type args struct {
			cfg         *config.Config
			currentTime time.Time
		}
		now := time.Now()

		testCases := []struct {
			name                    string
			args                    args
			wantSystemAvailableTime time.Time
			wantSystemUnavailable   bool
		}{
			{
				name: "No config",
				args: args{
					cfg:         &config.Config{},
					currentTime: time.Time{},
				},
				wantSystemAvailableTime: time.Time{},
				wantSystemUnavailable:   false,
			},
			{
				name: "Current time is before weekly maintenance times",
				args: args{
					cfg: &config.Config{
						WeeklyMaintenanceStartTime: "1900",
						WeeklyMaintenanceEndTime:   "1930",
						WeeklyMaintenanceDay:       now.Weekday(),
					},
					currentTime: time.Date(now.Year(), now.Month(), now.Day(), 18, 50, 0, 0, now.Location()),
				},
				wantSystemAvailableTime: time.Time{},
				wantSystemUnavailable:   false,
			},
			{
				name: "Current time is during weekly maintenance times",
				args: args{
					cfg: &config.Config{
						WeeklyMaintenanceStartTime: "1900",
						WeeklyMaintenanceEndTime:   "1930",
						WeeklyMaintenanceDay:       now.Weekday(),
					},
					currentTime: time.Date(now.Year(), now.Month(), now.Day(), 19, 5, 0, 0, now.Location()),
				},
				wantSystemAvailableTime: time.Date(now.Year(), now.Month(), now.Day(), 19, 30, 0, 0, now.Location()),
				wantSystemUnavailable:   true,
			},
			{
				name: "Current time is after weekly maintenance times",
				args: args{
					cfg: &config.Config{
						WeeklyMaintenanceStartTime: "1900",
						WeeklyMaintenanceEndTime:   "1930",
						WeeklyMaintenanceDay:       now.Weekday(),
					},
					currentTime: time.Date(now.Year(), now.Month(), now.Day(), 19, 40, 0, 0, now.Location()),
				},
				wantSystemAvailableTime: time.Time{},
				wantSystemUnavailable:   false,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				gotSystemAvailableTime, gotSystemUnavailable := checkWeeklyDownTime(tc.args.cfg, tc.args.currentTime)

				So(gotSystemAvailableTime, ShouldEqual, tc.wantSystemAvailableTime)
				So(gotSystemUnavailable, ShouldEqual, tc.wantSystemUnavailable)
			})
		}
	})
}

func TestUnit_isPlannedMaintenanceCheckRequired(t *testing.T) {
	Convey("Is planned maintenance check required", t, func() {
		type args struct {
			cfg *config.Config
		}
		testCases := []struct {
			name string
			args args
			want bool
		}{
			{
				name: "No config",
				args: args{
					cfg: &config.Config{},
				},
				want: false,
			},
			{
				name: "Config with PlannedMaintenanceStart and PlannedMaintenanceEnd",
				args: args{
					cfg: &config.Config{
						PlannedMaintenanceStart: "6 Feb 25 17:00 GMT",
						PlannedMaintenanceEnd:   "6 Feb 25 18:00 GMT",
					},
				},
				want: true,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := isPlannedMaintenanceCheckRequired(tc.args.cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})
}

func TestUnit_isWeeklyMaintenanceTimeCheckRequired(t *testing.T) {
	Convey("Is weekly maintenance time check required", t, func() {
		type args struct {
			cfg *config.Config
		}
		testCases := []struct {
			name string
			args args
			want bool
		}{
			{
				name: "Weekly maintenance time check not required",
				args: args{
					cfg: &config.Config{},
				},
				want: false,
			},
			{
				name: "Weekly maintenance time check required",
				args: args{
					cfg: &config.Config{
						WeeklyMaintenanceStartTime: "1900",
						WeeklyMaintenanceEndTime:   "1930",
						WeeklyMaintenanceDay:       0,
					},
				},
				want: true,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := isWeeklyMaintenanceTimeCheckRequired(tc.args.cfg)

				So(got, ShouldEqual, tc.want)
			})
		}
	})
}

func TestUnit_isWithinMaintenanceTime(t *testing.T) {
	Convey("Is within maintenance time", t, func() {
		type args struct {
			weeklyMaintenanceEndTime   time.Time
			currentTime                time.Time
			weeklyMaintenanceStartTime time.Time
		}
		testCases := []struct {
			name string
			args args
			want bool
		}{
			{
				name: "Current time is before weekly maintenance times",
				args: args{
					weeklyMaintenanceEndTime:   time.Date(2025, 2, 9, 19, 30, 0, 0, time.UTC),
					currentTime:                time.Date(2025, 2, 6, 19, 5, 0, 0, time.UTC),
					weeklyMaintenanceStartTime: time.Date(2025, 2, 9, 19, 0, 0, 0, time.UTC),
				},
				want: false,
			},
			{
				name: "Current time is during weekly maintenance times",
				args: args{
					weeklyMaintenanceEndTime:   time.Date(2025, 2, 9, 19, 30, 0, 0, time.UTC),
					currentTime:                time.Date(2025, 2, 9, 19, 5, 0, 0, time.UTC),
					weeklyMaintenanceStartTime: time.Date(2025, 2, 9, 19, 0, 0, 0, time.UTC),
				},
				want: true,
			},
			{
				name: "Current time is after weekly maintenance times",
				args: args{
					weeklyMaintenanceEndTime:   time.Date(2025, 2, 9, 19, 30, 0, 0, time.UTC),
					currentTime:                time.Date(2025, 2, 10, 19, 5, 0, 0, time.UTC),
					weeklyMaintenanceStartTime: time.Date(2025, 2, 9, 19, 0, 0, 0, time.UTC),
				},
				want: false,
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := isWithinMaintenanceTime(tc.args.weeklyMaintenanceEndTime, tc.args.currentTime, tc.args.weeklyMaintenanceStartTime)

				So(got, ShouldEqual, tc.want)
			})
		}
	})
}

func TestUnit_returnWeeklyMaintenanceTime(t *testing.T) {
	Convey("Return weekly maintenance time", t, func() {
		type args struct {
			currentTime time.Time
			hour        string
			minute      string
		}
		now := time.Now()

		testCases := []struct {
			name string
			args args
			want time.Time
		}{
			{
				name: "Success WEEKLY_MAINTENANCE_START_TIME=1900",
				args: args{
					currentTime: now,
					hour:        "19",
					minute:      "00",
				},
				want: time.Date(now.Year(), now.Month(), now.Day(), 19, 0, 0, 0, now.Location()),
			},
			{
				name: "Success WEEKLY_MAINTENANCE_END_TIME=1930",
				args: args{
					currentTime: now,
					hour:        "19",
					minute:      "30",
				},
				want: time.Date(now.Year(), now.Month(), now.Day(), 19, 30, 0, 0, now.Location()),
			},
		}
		for _, tc := range testCases {
			Convey(tc.name, func() {
				got := returnWeeklyMaintenanceTime(tc.args.currentTime, tc.args.hour, tc.args.minute)

				So(got, ShouldEqual, tc.want)
			})
		}
	})
}
