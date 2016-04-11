package common

import (
	"time"
)

const (
	DiscordTimeFormat = "2006-01-02T15:04:05-07:00"
)

type LogMessage struct {
	Timestamp time.Time
	Content   string
}
