package main

import (
	"log"

	"github.com/Wetitpig/etaHK/GMB"
	"github.com/Wetitpig/etaHK/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	ui.App = tview.NewApplication()
	ui.UserLang = ui.TC
	ui.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		_, ok := ui.App.GetFocus().(*tview.TextArea)
		_, ok2 := ui.App.GetFocus().(*tview.InputField)
		if event.Key() == tcell.KeyRune && !(ok || ok2) {
			switch event.Rune() {
			case 't':
				ui.UserLang = ui.TC
			case 's':
				ui.UserLang = ui.SC
			case 'e':
				ui.UserLang = ui.EN
			case 'q':
				ui.App.Stop()
			}
		}
		return event
	})
	ui.BeforeDrawFn = make(map[string]func(tcell.Screen) bool)
	ui.App.SetBeforeDrawFunc(ui.RunBeforeDraw)
	ui.Pages = tview.NewPages()

	GMB.ListGMB()

	if err := ui.App.SetRoot(ui.Pages, true).Run(); err != nil {
		log.Fatalln(err)
	}
}