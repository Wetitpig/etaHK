package main

import (
	"github.com/Wetitpig/etaHK/Bus"
	"github.com/Wetitpig/etaHK/GMB"
	"github.com/Wetitpig/etaHK/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func homepage() (frame *tview.Frame) {
	mode := tview.NewForm().AddButton("", GMB.ListGMB).
		AddButton("", Bus.ListBus).
		SetButtonsAlign(tview.AlignCenter)
	frame = tview.NewFrame(mode).
		AddText("", true, tview.AlignCenter, tcell.ColorDefault).
		AddText("", false, tview.AlignCenter, tcell.ColorGreen)

	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && (event.Rune() == 't' || event.Rune() == 's' || event.Rune() == 'e') {
			ui.UpdateHomepage()
		}
		return event
	})
	return
}
