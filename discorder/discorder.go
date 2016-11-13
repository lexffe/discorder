package discorder

import (
	"fmt"
	"time"
)

const (
	DiscordTimeFormat = "2006-01-02T15:04:05-07:00"

	VERSION_MAJOR = 0
	VERSION_MINOR = 4
	VERSION_PATCH = 0
	VERSION_NOTE  = "Git-Steaming"
)

var (
	VERSION = fmt.Sprintf("%d.%d.%d-%s", VERSION_MAJOR, VERSION_MINOR, VERSION_PATCH, VERSION_NOTE)
)

type LogMessage struct {
	Timestamp time.Time
	Content   string
}
