package ui

import (
	"time"

	"github.com/rivo/tview"
)

type StatusMessage struct {
	Message string
	Timer   int
}

func ProcessGlobalStatus(globalStatus *tview.TextView, statusChan <-chan StatusMessage) {
	var displayTimer *time.Timer
	for {
		select {
		case msg := <-statusChan:
			if displayTimer != nil {
				displayTimer.Stop() // stop any active timer
			}
			globalStatus.SetDynamicColors(true).SetTextAlign(tview.AlignCenter).SetText(msg.Message)
			if msg.Timer > 0 {
				displayTimer = time.NewTimer(time.Duration(msg.Timer) * time.Second)
				go func() {
					<-displayTimer.C
					globalStatus.SetText("")
				}()
			}
		}
	}
}
