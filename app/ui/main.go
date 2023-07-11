package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func SetTheme() {
	// Define custom theme
	tview.Styles = tview.Theme{
		PrimitiveBackgroundColor:    tcell.ColorBlack,
		ContrastBackgroundColor:     tcell.ColorBlue,
		MoreContrastBackgroundColor: tcell.ColorGreen,
		BorderColor:                 tcell.ColorWhite,
		TitleColor:                  tcell.ColorWhite,
		GraphicsColor:               tcell.ColorWhite,
		PrimaryTextColor:            tcell.ColorCadetBlue,
		SecondaryTextColor:          tcell.ColorYellow,
		TertiaryTextColor:           tcell.ColorGreen,
		InverseTextColor:            tcell.ColorBlue,
		ContrastSecondaryTextColor:  tcell.ColorDarkBlue,
	}
}

func InitInfoBox() *tview.Flex {
	infoBox := tview.NewFlex()
	infoBox.SetBorder(true)
	infoBox.SetTitle("Infos")
	return infoBox
}

func HandleTabSwitchView(app *tview.Application, focused **tview.Flex, infoBox, tableBox **tview.Flex) func(event *tcell.EventKey) *tcell.EventKey {
	return func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyTab:
			if *focused == *tableBox {
				*focused = *infoBox
			} else {
				*focused = *tableBox
			}
			app.SetFocus(tview.Primitive(*focused))
			return nil
		}
		return event
	}
}
