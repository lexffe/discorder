package discorder

import (
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

var logRoutine *LogRoutine

func InitLogging(logPath string) {
	logRoutine = NewLogRoutine(logPath)
	log.SetOutput(logRoutine)
	go logRoutine.Run()
}

type LogRoutine struct {
	lastLogTime time.Time
	sync.RWMutex

	In chan *LogMessage

	buffer []*LogMessage
	file   *os.File
}

func NewLogRoutine(logPath string) *LogRoutine {
	var logFile *os.File
	var err error
	if logPath != "" {
		logFile, err = os.Create(logPath)
		if err != nil {
			log.Println("Failed to create logfile:", err)
		}
	}

	in := make(chan *LogMessage, 100)
	return &LogRoutine{
		In:   in,
		file: logFile,
	}
}

func (l *LogRoutine) Run() {
	for {
		select {
		case msg := <-l.In:
			l.handleMsg(msg)
		}
	}
}

func (l *LogRoutine) handleMsg(msg *LogMessage) {
	l.Lock()
	if l.file != nil {
		l.file.Write([]byte(msg.Content + "\n"))
	}

	l.buffer = append(l.buffer, msg)
	l.lastLogTime = msg.Timestamp
	l.Unlock()
}

func (l *LogRoutine) Write(data []byte) (n int, err error) {
	split := strings.Split(string(data), "\n")
	now := time.Now()

	for _, splitStr := range split {
		if splitStr == "" {
			continue
		}
		msg := &LogMessage{
			Timestamp: now,
			Content:   splitStr,
		}
		l.In <- msg
	}
	return len(data), nil
}

func (l *LogRoutine) GetCopy() []*LogMessage {
	l.RLock()
	cop := make([]*LogMessage, len(l.buffer))
	copy(cop, l.buffer)
	l.RUnlock()
	return cop
}

func (l *LogRoutine) HasChangedSince(since time.Time) bool {
	l.RLock()
	changed := !since.Equal(l.lastLogTime)
	l.RUnlock()
	return changed
}

func (l *LogRoutine) Clear() {
	l.Lock()
	l.buffer = []*LogMessage{}
	l.Unlock()
}
