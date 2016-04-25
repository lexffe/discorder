package main

import (
	"github.com/jonas747/discorder/common"
	"github.com/jonas747/discordgo"
	"log"
	"time"
)

type AckRoutine struct {
	App  *App
	In   chan *discordgo.Message
	Stop chan bool
}

func NewAckRoutine(app *App) *AckRoutine {
	return &AckRoutine{
		App:  app,
		In:   make(chan *discordgo.Message),
		Stop: make(chan bool),
	}
}

func (a *AckRoutine) Run() {
	// Send ack's every 5th second if any, with the latest message from the channel
	ticker := time.NewTicker(5 * time.Second)

	curAckBuffer := make([]*discordgo.Message, 0)

	for {
		select {
		case m := <-a.In:
			ts, err := time.Parse(common.DiscordTimeFormat, m.Timestamp)
			if err != nil {
				log.Println("Error pasing timestamp", err)
				continue
			}
			found := false
			for k, v := range curAckBuffer {
				if v.ChannelID == m.ChannelID {
					found = true
					parsed, err := time.Parse(common.DiscordTimeFormat, v.Timestamp)
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
		case <-a.Stop:
			ticker.Stop()
			return
		case <-ticker.C:
			for _, v := range curAckBuffer {
				err := a.App.session.ChannelMessageAck(v.ChannelID, v.ID)
				if err != nil {
					log.Println("Error sending ack: ", err)
				}
				log.Println("Send ack!", v.ChannelID, v.ID)
			}
			curAckBuffer = make([]*discordgo.Message, 0)
		}
	}
}
