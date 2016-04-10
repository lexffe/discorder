package common

import (
	"time"
)

type LogMessage struct {
	Timestamp time.Time
	Content   string
}
