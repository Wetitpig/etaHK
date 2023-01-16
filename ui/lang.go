package ui

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Language int

var UserLang Language

const (
	TC Language = iota
	SC
	EN
	LangCount
)

type Lang = [LangCount]string

func (l Language) String() string {
	switch l {
	case TC:
		return "TC"
	case SC:
		return "SC"
	case EN:
		return "EN"
	default:
		return ""
	}
}

var (
	RouteSelectLabel = Lang{"路綫", "路线", "Route"}
	NextUpdateLabel  = Lang{"下次更新：", "下回更新：", "Next update: "}
)

func UpdateHomepage() {
	welcomeLabel := Lang{
		"歡迎使用 ETA@HK", "欢迎使用 ETA@HK", "Welcome to ETA@HK",
	}
	qLabel := Lang{
		"如有任何疑問，請前往 https://github.com/Wetitpig/etaHK 新增 Issue。",
		"如有任何疑问，请前往 https://github.com/Wetitpig/etaHK 创建 Issue。",
		"Please create an issue on https://github.com/Wetitpig/etaHK for any questions.",
	}
	gmbLabel := Lang{"專綫小巴", "专线小巴", "Green Minibus"}
	busLabel := Lang{"巴士 ", "巴士", "Bus"}

	_, f := Pages.GetFrontPage()
	f.(*tview.Frame).Clear().
		AddText(welcomeLabel[UserLang], true, tview.AlignCenter, tcell.ColorDefault).
		AddText(qLabel[UserLang], false, tview.AlignCenter, tcell.ColorGreen)

	m := f.(*tview.Frame).GetPrimitive().(*tview.Form)
	m.GetButton(0).SetLabel(gmbLabel[UserLang])
	m.GetButton(1).SetLabel(busLabel[UserLang])
}
