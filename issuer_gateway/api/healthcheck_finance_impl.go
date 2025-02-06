package api

import (
	"fmt"
	"strconv"
	"time"

	"github.com/companieshouse/chs.go/log"
	"github.com/companieshouse/penalty-payment-api/config"
)

var getConfig = func() (*config.Config, error) {
	return config.Get()
}

type IssuerGatewayHealthcheckFinanceSystem struct {
}

func (ig *IssuerGatewayHealthcheckFinanceSystem) CheckScheduledMaintenance() (systemAvailableTime time.Time, systemUnavailable bool, parseError bool) {

	cfg, err := getConfig()
	if err != nil {
		err = fmt.Errorf("error getting config for planned maintenance: [%v]", err)
		return time.Time{}, false, true
	}

	currentTime := time.Now()

	systemAvailableTime, systemUnavailable = checkWeeklyDownTime(cfg, currentTime)

	if isPlannedMaintenanceCheckRequired(cfg) {
		timeDateLayout := time.RFC822
		maintenanceStart, err := time.Parse(timeDateLayout, cfg.PlannedMaintenanceStart)
		if err != nil {
			log.Error(fmt.Errorf("error parsing Maintenance Start time: [%v]", err))
			return time.Time{}, false, true
		}
		maintenanceEnd, err := time.Parse(timeDateLayout, cfg.PlannedMaintenanceEnd)
		if err != nil {
			log.Error(fmt.Errorf("error parsing Maintenance End time: [%v]", err))
			return time.Time{}, false, true
		}

		if maintenanceEnd.After(currentTime) && maintenanceStart.Before(currentTime) && maintenanceEnd.After(systemAvailableTime) {
			systemAvailableTime = maintenanceEnd
			systemUnavailable = true
		}
	}
	return systemAvailableTime, systemUnavailable, false
}

func checkWeeklyDownTime(cfg *config.Config, currentTime time.Time) (systemAvailableTime time.Time, systemUnavailable bool) {
	if isWeeklyMaintenanceTimeCheckRequired(cfg) {
		// If the weekday is maintenance day
		if currentTime.Weekday() == cfg.WeeklyMaintenanceDay {

			weeklyMaintenanceStartTime := returnWeeklyMaintenanceTime(currentTime, cfg.WeeklyMaintenanceStartTime[:2], cfg.WeeklyMaintenanceStartTime[2:])

			weeklyMaintenanceEndTime := returnWeeklyMaintenanceTime(currentTime, cfg.WeeklyMaintenanceEndTime[:2], cfg.WeeklyMaintenanceEndTime[2:])

			// Check if time is within maintenance time
			if isWithinMaintenanceTime(weeklyMaintenanceEndTime, currentTime, weeklyMaintenanceStartTime) {
				systemAvailableTime = weeklyMaintenanceEndTime
				systemUnavailable = true
			}
		}
	}
	return systemAvailableTime, systemUnavailable
}

func isWeeklyMaintenanceTimeCheckRequired(cfg *config.Config) bool {
	return cfg.WeeklyMaintenanceStartTime != "" && cfg.WeeklyMaintenanceEndTime != ""
}

// returnWeeklyMaintenanceTime returns a time.Time format for the current date with the hour and minute set to the arguments passed
func returnWeeklyMaintenanceTime(currentTime time.Time, hour, minute string) time.Time {

	intHour, _ := strconv.Atoi(hour)
	timeDifferenceInHours := time.Duration(intHour - currentTime.Hour())

	intMinute, _ := strconv.Atoi(minute)
	timeDifferenceInMinutes := time.Duration(intMinute - currentTime.Minute())

	secondDuration := time.Duration(0 - currentTime.Second())
	nanosecondDuration := time.Duration(0 - currentTime.Nanosecond())

	return currentTime.Add(time.Hour*timeDifferenceInHours + time.Minute*timeDifferenceInMinutes + time.Second*secondDuration + time.Nanosecond*nanosecondDuration).Round(0)
}

func isWithinMaintenanceTime(weeklyMaintenanceEndTime time.Time, currentTime time.Time, weeklyMaintenanceStartTime time.Time) bool {
	return weeklyMaintenanceEndTime.After(currentTime) && weeklyMaintenanceStartTime.Before(currentTime)
}

func isPlannedMaintenanceCheckRequired(cfg *config.Config) bool {
	return cfg.PlannedMaintenanceStart != "" && cfg.PlannedMaintenanceEnd != ""
}
