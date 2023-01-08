package main

import (
	"github.com/Wetitpig/etaHK/GMB"
	"github.com/Wetitpig/etaHK/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func homepage() (frame *tview.Frame) {
	welcomeLabel := ui.Lang{
		"歡迎使用 ETA@HK", "欢迎使用 ETA@HK", "Welcome to ETA@HK",
	}
	qLabel := ui.Lang{
		"如有任何疑問，請前往 https://github.com/Wetitpig/etaHK 新增 Issue。",
		"如有任何疑问，请前往 https://github.com/Wetitpig/etaHK 创建 Issue。",
		"Please create an issue on https://github.com/Wetitpig/etaHK for any questions.",
	}
	gmbLabel := ui.Lang{"專綫小巴", "专线小巴", "Green Minibus"}
	busLabel := ui.Lang{"巴士 ", "巴士", "Bus"}

	mode := tview.NewForm().AddButton(gmbLabel[ui.UserLang], GMB.ListGMB).
		AddButton(busLabel[ui.UserLang], nil).
		SetButtonsAlign(tview.AlignCenter)
	frame = tview.NewFrame(mode).
		AddText(welcomeLabel[ui.UserLang], true, tview.AlignCenter, tcell.ColorDefault).
		AddText(qLabel[ui.UserLang], false, tview.AlignCenter, tcell.ColorGreen)

	frame.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && (event.Rune() == 't' || event.Rune() == 's' || event.Rune() == 'e') {
			frame.Clear().
				AddText(welcomeLabel[ui.UserLang], true, tview.AlignCenter, tcell.ColorDefault).
				AddText(qLabel[ui.UserLang], false, tview.AlignCenter, tcell.ColorGreen)

			mode.GetButton(0).SetLabel(gmbLabel[ui.UserLang])
			mode.GetButton(1).SetLabel(busLabel[ui.UserLang])
		}
		return event
	})
	return
}
