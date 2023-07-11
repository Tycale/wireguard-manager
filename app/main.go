package main

import (
	"github.com/rivo/tview"

	"log"
	"time"

	"github.com/tycale/wireguard-manager/app/ui"
	"github.com/tycale/wireguard-manager/app/wg"
)

func main() {
	AutoSu()

	app := tview.NewApplication()

	ui.SetTheme()

	headerFlex := ui.CreateHeader()

	files, err := wg.CheckConfigFiles()
	if err != nil {
		log.Fatal("The directory ~/.wg/ does not exist ? Err: ", err)
	}

	if len(files) == 0 {
		log.Fatal("The directory ~/.wg/ does not contain any .conf file ")
	}

	// Status messages
	globalStatus := tview.NewTextView()
	statusChan := make(chan ui.StatusMessage)
	go ui.ProcessGlobalStatus(globalStatus, statusChan)

	eventChan := make(chan ui.Event)

	rows := ui.CreateRows(files)
	infoBox := ui.InitInfoBox()
	table := ui.CreateTable(rows, infoBox, eventChan)
	tableBox := ui.InitTableBox(table)

	// Flex layout for the main screen
	mainFlex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(headerFlex, 6, 0, false).
		AddItem(globalStatus, 1, 0, false).
		AddItem(tableBox, 0, 1, true).
		AddItem(infoBox, 0, 2, true)

	focused := tableBox // initially, tableBox is focused
	app.SetInputCapture(ui.HandleTabSwitchView(app, &focused, &infoBox, &tableBox))

	go wg.ListenWGChan(app, eventChan, rows, statusChan)
	for _, row := range rows {
		r := row
		// Fetching the iface is a lot of operations on Darwin
		go func(row *ui.TableData) {
			wg.RefreshInterface(row)
			ticker := time.NewTicker(5 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				wg.RefreshInterface(row)
			}
		}(&r)
		go func(row *ui.TableData) {
			ticker := time.NewTicker(1 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				wg.UpdateStatus(app, row, table)
			}
		}(&r)
	}

	if err := app.SetRoot(mainFlex, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}
}
