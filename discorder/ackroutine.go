package discorder

import (
	"github.com/jonas747/discordgo"
	"log"
	"sync"
	"time"
)

type AckRoutine struct {
	App  *App
	In   chan *discordgo.Message
	stop chan *sync.WaitGroup
}

func NewAckRoutine(app *App) *AckRoutine {
	return &AckRoutine{
		App:  app,
		In:   make(chan *discordgo.Message),
		stop: make(chan *sync.WaitGroup),
	}
}

func (a *AckRoutine) Run() {
	// Send ack's every 15 seconds if any, with the latest message from the channel
	ticker := time.NewTicker(15 * time.Second)

	curAckBuffer := make([]*discordgo.Message, 0)

	for {
		select {
		case m := <-a.In:
			ts, err := time.Parse(DiscordTimeFormat, m.Timestamp)
			if err != nil {
				continue
			}
			found := false
			for k, v := range curAckBuffer {
				if v.ChannelID == m.ChannelID {
					found = true
					parsed, err := time.Parse(DiscordTimeFormat, v.Timestamp)
					if err != nil {
						log.Println("Error pasring timestamp", err)
					}
					if ts.After(parsed) {
						curAckBuffer[k] = m
					}
					break
				}
			}
			if !found {
				curAckBuffer = append(curAckBuffer, m)
			}
		case wg := <-a.stop:
			ticker.Stop()
			wg.Done()
			log.Println("Ackroutine shut down")
			return
		case <-ticker.C:
			for _, v := range curAckBuffer {
				a.AckMessage(v)
			}
			curAckBuffer = make([]*discordgo.Message, 0)
		}
	}
}

// Send a ack if we should, the read state check may be overkill? dunno, should check how the official client handles it
func (a AckRoutine) AckMessage(msg *discordgo.Message) {
	// Check the readstate first to verify if we already have ack'd this messaeg before
	state := a.App.session.State
	state.Lock()
	defer state.Unlock()

	if state.User.Bot {
		return // Bot accounts can't do acks
	}

	// Do we really need this check here? maybe move it to the history processing...
	shouldAck := true
	ackTs, err := time.Parse(DiscordTimeFormat, msg.Timestamp)
	if err == nil {
		for _, rs := range state.ReadState {
			if rs.ID == msg.ChannelID {
				// Check if we have already read this message
				if rs.LastMessageID == msg.ID {
					return
				}

				lastRead := rs.LastMessageID
				// Find the message
				state.Unlock()
				channel, err := state.Channel(msg.ChannelID)
				state.Lock()
				if err != nil {
					break
				}
				for _, cm := range channel.Messages {
					if cm.ID == lastRead {
						parsedTs, err := time.Parse(DiscordTimeFormat, cm.Timestamp)
						if err == nil {
							if ackTs.Before(parsedTs) {
								// Do not ack, this message is older than the last read message
								return
							}
						}
						break
					}
				}

				break
			}
		}
	}

	if !shouldAck {
		return
	}
	state.Unlock()
	channel, _ := a.App.session.State.Channel(msg.ChannelID)
	msgStr := msg.ChannelID
	if channel != nil {
		msgStr = GetChannelNameOrRecipient(channel)
	}
	err = a.App.session.ChannelMessageAck(msg.ChannelID, msg.ID)
	if err != nil {
		log.Println("Error sending ack: ", err)
	}

	if a.App.options.DebugEnabled {
		log.Println("Send ack!", msgStr, msg.ID)
	}

	state.Lock()
	a.SetReadState(msg)
}

// Sets the last read message, should also undo notifications bound to this channel and readstate?
func (a *AckRoutine) SetReadState(msg *discordgo.Message) {
	found := false
	for _, v := range a.App.session.State.ReadState {
		if v.ID == msg.ChannelID {
			v.LastMessageID = msg.ID
			found = true
			break
		}
	}
	if !found {
		readState := &discordgo.ReadState{
			ID:            msg.ChannelID,
			LastMessageID: msg.ID,
		}
		a.App.session.State.ReadState = append(a.App.session.State.ReadState, readState)
	}
}
