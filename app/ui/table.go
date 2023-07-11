package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Event struct {
	Action string
	Arg    string
	No     int
}

type TableData struct {
	No         int
	Connected  string
	Name       string
	AllowedIps string
	Infos      *tview.TextView
	Iface      string
}

const (
	StatusUnknown   = "Unknown"
	StatusNo        = "Disconnected"
	StatusConnected = "Connected"
)

func InitTableBox(table *tview.Table) *tview.Flex {
	tableBox := tview.NewFlex()
	tableBox.SetBorder(true)
	tableBox.SetTitle("Tunnels")
	tableBox.AddItem(table, 0, 1, true)
	return tableBox
}

func CreateRows(files []string) []TableData {
	rows := make([]TableData, len(files))
	for i, file := range files {
		rows[i] = initRow(file)
		rows[i].No = i
	}
	return rows
}

func initRow(env string) TableData {
	textView := tview.NewTextView()
	textView.Write([]byte("Infos for " + env))
	textView.SetScrollable(true)

	row := TableData{
		Connected:  StatusNo,
		Name:       env,
		AllowedIps: StatusUnknown,
		Infos:      textView,
		Iface:      "",
	}

	return row
}

func CreateTable(data []TableData, infoBox *tview.Flex, eventChan chan Event) *tview.Table {
	table := tview.NewTable().SetBorders(false)

	// Add column headers
	table.SetCell(0, 0, tview.NewTableCell("Connection status").SetExpansion(1))
	table.SetCell(0, 1, tview.NewTableCell("Name").SetExpansion(2))
	table.SetCell(0, 2, tview.NewTableCell("AllowedIps").SetExpansion(5))
	table.SetEvaluateAllRows(true)
	table.SetFixed(1, 0)

	// Add rows from data
	for i, row := range data {
		connectedCell := tview.NewTableCell(row.Connected)
		if row.Connected == StatusNo {
			connectedCell.SetTextColor(tcell.ColorOrchid)
		} else {
			connectedCell.SetTextColor(tcell.ColorGreen)
		}
		table.SetCell(i+1, 0, connectedCell)
		table.SetCell(i+1, 1, tview.NewTableCell(row.Name))
		table.SetCell(i+1, 2, tview.NewTableCell(row.AllowedIps))
	}

	// Change selectioned row
	table.SetSelectionChangedFunc(func(row, column int) {
		infoBox.Clear()
		infoBox.AddItem(data[row-1].Infos, 0, 1, true)
	})

	table.SetSelectable(true, false) // Make rows selectable, not columns
	table.Select(1, 0)

	// Capture input events
	table.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		row, _ := table.GetSelection()

		// Get the total number of rows in the table
		totalRows := table.GetRowCount()

		// Prevent from moving outside the "Connected" column
		switch event.Key() {
		case tcell.KeyRight, tcell.KeyLeft:
			return nil
		case tcell.KeyUp:
			if row == 1 {
				return nil
			}
		case tcell.KeyDown:
			if row == totalRows-1 {
				return nil
			}
		}

		// Capture 'j' and 'k' keys for navigation
		switch event.Rune() {
		case 'j':
			if row < totalRows-1 {
				table.Select(row+1, 0)
			}
			return nil
		case 'k':
			if row > 1 {
				table.Select(row-1, 0)
			}
			return nil
		case 'c':
			eventChan <- Event{"up", data[row-1].Name, data[row-1].No}
		case 'd':
			eventChan <- Event{"down", data[row-1].Name, data[row-1].No}
		case 'q':
			eventChan <- Event{"quit", "", 0}
			return nil
		}

		// Propagate the event
		return event
	})

	return table
}
