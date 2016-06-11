// Manages some requests like history
package discorder

import (
	"container/list"
	"github.com/jonas747/discordgo"
	"log"
	"sync"
	"time"
)

type Request interface {
	CheckDuplicate(Request) bool
	Do(finished chan error)
}

type HistoryRequest struct {
	App       *App
	Limit     int
	ChannelID string
	BeforeID  string
	AfterID   string
}

func NewHistoryRequest(app *App, channel string, limit int, before, after string) *HistoryRequest {
	return &HistoryRequest{
		App:       app,
		Limit:     limit,
		ChannelID: channel,
		BeforeID:  before,
		AfterID:   after,
	}
}

func (hq *HistoryRequest) CheckDuplicate(r Request) bool {
	cast, ok := r.(*HistoryRequest)
	if !ok {
		return false
	}
	if cast.ChannelID == hq.ChannelID {
		if hq.AfterID == cast.AfterID && hq.BeforeID == cast.BeforeID {
			return true
		}
	}

	return false
}

func (hq *HistoryRequest) Do(finished chan error) {
	var err error
	defer func() {
		hq.App.Lock()
		if hq.App.debug {
			log.Println("History processing complete")
		}
		hq.App.Unlock()

		finished <- err
	}()

	hq.App.RLock()
	state := hq.App.session.State
	hq.App.RUnlock()

	channel, err := state.Channel(hq.ChannelID)
	if err != nil {
		//log.Println("History error: ", err)
		return
	}

	// func (s *Session) ChannelMessages(channelID string, limit int, beforeID, afterID string) (st []*Message, err error)
	resp, err := hq.App.session.ChannelMessages(hq.ChannelID, hq.Limit, hq.BeforeID, hq.AfterID)
	if err != nil {
		//log.Println("History error: ", err)
		return
	}

	if len(resp) < 1 {
		hq.App.Lock()
		hq.App.firstMessages[hq.ChannelID] = hq.BeforeID
		hq.App.Unlock()
		return
	}

	if len(resp) < hq.Limit {
		// Looks like we've hit the first message of the channel
		hq.App.Lock()

		hq.App.firstMessages[hq.ChannelID] = resp[len(resp)-1].ID

		hq.App.Unlock()
	}

	state.Lock()
	if len(channel.Messages) < 1 && len(resp) > 0 {
		for i := len(resp) - 1; i >= 0; i-- {
			channel.Messages = append(channel.Messages, resp[i])
		}
		state.Unlock()
		return
	}

	newMessages := make([]*discordgo.Message, 0)
	nextNewMessageIndex := len(resp) - 1
	nextOldMessageIndex := 0

	for {
		newOut := false // new (response) is oob
		oldOut := false // old (current channel history) is oob
		var nextOldMessage *discordgo.Message
		if nextOldMessageIndex >= len(channel.Messages) {
			oldOut = true
		} else {
			nextOldMessage = channel.Messages[nextOldMessageIndex]
		}

		var nextNewMessage *discordgo.Message
		if nextNewMessageIndex < 0 {
			newOut = true
		} else {
			nextNewMessage = resp[nextNewMessageIndex]
		}

		if newOut && !oldOut {
			newMessages = append(newMessages, nextOldMessage)
			nextOldMessageIndex++
			continue
		} else if !newOut && oldOut {
			newMessages = append(newMessages, nextNewMessage)
			nextNewMessageIndex--
			continue
		} else if newOut && oldOut {
			break
		}

		if nextNewMessage.ID == nextOldMessage.ID {
			newMessages = append(newMessages, nextNewMessage)
			nextNewMessageIndex--
			nextOldMessageIndex++
			continue
		}

		parsedNew, _ := time.Parse(DiscordTimeFormat, nextNewMessage.Timestamp)
		parsedOld, _ := time.Parse(DiscordTimeFormat, nextOldMessage.Timestamp)

		if parsedNew.Before(parsedOld) {
			newMessages = append(newMessages, nextNewMessage)
			nextNewMessageIndex--
		} else {
			newMessages = append(newMessages, nextOldMessage)
			nextOldMessageIndex++
		}
	}
	channel.Messages = newMessages

	state.Unlock()

	if len(resp) > 0 {
		hq.App.ackRoutine.In <- resp[0]
	}
}

type RequestRoutine struct {
	sync.RWMutex
	queue *list.List
	In    chan Request
	stop  chan *sync.WaitGroup
}

func NewRequestRoutine() *RequestRoutine {
	return &RequestRoutine{
		queue: list.New(),
		In:    make(chan Request),
		stop:  make(chan *sync.WaitGroup),
	}
}

func (rr *RequestRoutine) Run() {
	finished := make(chan error)

	//cooldown := make(<-chan time.Time)
	var cooldown <-chan time.Time
	isFetching := false // Currently working on a request
	isReady := true     // Cooldown for next request over

	for {
		select {
		case req := <-rr.In:

			rr.addRequest(req)

			if isReady && !isFetching && rr.processNext(finished) {
				isFetching = true
				isReady = false
				cooldown = time.After(2 * time.Second)
			}

		case err := <-finished:

			if err != nil {
				log.Println("Request error:", err)
			}

			if isReady && rr.processNext(finished) {
				isFetching = true
				isReady = false
				cooldown = time.After(2 * time.Second)
			} else {
				isFetching = false
			}

		case <-cooldown:

			if !isFetching && rr.processNext(finished) {
				isFetching = true
				isReady = false
				cooldown = time.After(2 * time.Second)
			} else {
				isReady = true
			}
		case wg := <-rr.stop:
			wg.Done()
			log.Println("Request routine shut down")
			return
		}
	}
}

func (rr *RequestRoutine) processNext(finished chan error) bool {
	rr.Lock()
	defer rr.Unlock()

	if rr.queue.Len() < 1 {
		return false
	}

	v := rr.queue.Remove(rr.queue.Front())
	cast, ok := v.(Request)
	if !ok {
		log.Println("Unknown value in request queue?")
		return true
	}

	go cast.Do(finished)
	return true
}

func (rr *RequestRoutine) addRequest(req Request) {
	rr.Lock()
	defer rr.Unlock()
	for e := rr.queue.Front(); e != nil; e = e.Next() {
		cast, ok := e.Value.(Request)
		if !ok {
			log.Println("Something is very wrong, pleas nag me (jonas747) about this")
			continue // TODO: Remove it..
		}
		if req.CheckDuplicate(cast) {
			return
		}
	}

	rr.queue.PushBack(req)
}

func (rr *RequestRoutine) GetQueueLenth() int {
	rr.RLock()
	length := rr.queue.Len()
	rr.RUnlock()
	return length
}

func (rr *RequestRoutine) AddRequest(r Request) {
	rr.In <- r
}
