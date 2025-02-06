package api

import (
	"time"
)

type HealthcheckFinanceSystem interface {
	CheckScheduledMaintenance() (systemAvailableTime time.Time, systemUnavailable bool, parseError bool)
}
