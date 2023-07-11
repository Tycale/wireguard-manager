package ui

import (
	"github.com/rivo/tview"
)

func CreateHeader() *tview.Flex {
	asciiArt := tview.NewTextView().SetDynamicColors(true).SetText(`[red]
 __  _  ______   _____
 \ \/ \/ / ___\ /     \
  \     / /_/  >  Y Y  \
   \/\_/\___  /|__|_|  /
       /_____/       \/[white]`)

	// Create keyboard shortcuts help
	help := tview.NewTextView().SetDynamicColors(true).
		SetText("[yellow]Keyboard Shortcuts[white]\n\n" +
			"[blue]<tab>[white] Switch panel\n" +
			"[blue]c[white] Connect   \n" +
			"[blue]d[white] Disconnect\n" +
			"[blue]q[white] Quit\n").
		SetTextAlign(tview.AlignCenter)

	// Create a Flex layout for the header (asciiArt and help)
	headerFlex := tview.NewFlex().
		AddItem(asciiArt, 0, 1, false).
		AddItem(help, 0, 1, false)

	return headerFlex
}
