package api

import (
	"time"
)

type HealthcheckFinanceSystem interface {
	CheckScheduledMaintenance() (time.Time, bool, bool)
}
