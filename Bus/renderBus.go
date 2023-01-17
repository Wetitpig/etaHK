package Bus

import (
	"math/bits"
	"strconv"
	"strings"

	"github.com/Wetitpig/etaHK/ui"
	"github.com/agnivade/levenshtein"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/nathan-fiscaletti/consolesize-go"
	"github.com/rivo/tview"
	"golang.org/x/exp/slices"
)

var (
	chosenOp operator
	routeRID []uint64
)

func renderRoutesEvery(form *tview.Form) {
	view := form.GetFormItem(opCount + 1).(*tview.TextView)
	selected := ui.RetainHighlight(view)
	routeRID = routeRID[:0]
	searchStr := form.GetFormItem(0).(*tview.InputField).GetText()

	for id, route := range busList.RouteIndex {
		if chosenOp&route.Op != 0 &&
			strings.Contains(strings.ToUpper(route.Code), strings.ToUpper(searchStr)) {
			routeRID = append(routeRID, id)
		}
	}
	slices.SortFunc(routeRID, func(i, j uint64) bool {
		i_len := levenshtein.ComputeDistance(searchStr, busList.RouteIndex[i].Code)
		j_len := levenshtein.ComputeDistance(searchStr, busList.RouteIndex[j].Code)
		if i_len != j_len {
			return i_len < j_len
		} else if busList.RouteIndex[i].Code != busList.RouteIndex[j].Code {
			return busList.RouteIndex[i].Code < busList.RouteIndex[j].Code
		} else {
			i_len = bits.OnesCount8(uint8(busList.RouteIndex[i].Op))
			j_len = bits.OnesCount8(uint8(busList.RouteIndex[j].Op))
			if i_len != j_len {
				return i_len < j_len
			} else {
				return bits.TrailingZeros8(uint8(busList.RouteIndex[i].Op)) < bits.TrailingZeros8(uint8(busList.RouteIndex[j].Op))
			}
		}
	})

	buf := new(ui.TextViewBuffer)
	for _, id := range routeRID {
		v := busList.RouteIndex[id]

		rg := strconv.FormatUint(id, 10)
		buf.Write("[\"", rg, "\"]", v.Color(), runewidth.FillRight(v.Code, 6), v.Orig[ui.UserLang])
		if len(v.Dir) > 1 {
			buf.Write("<")
		}
		buf.Write("=>", v.Dest[ui.UserLang], "[-:-]\n")
		if rg == selected {
			view.Highlight(selected)
		}
	}
	buf.Print(view)
	if len(view.GetHighlights()) == 0 && len(routeRID) > 0 {
		view.Highlight(strconv.FormatUint(routeRID[0], 10))
	}
	view.ScrollToHighlight()
}

func renderRoutesLang(form *tview.Form) {
	form.GetFormItem(0).(*tview.InputField).SetLabel(ui.RouteSelectLabel[ui.UserLang])
	for i := 0; i < opCount; i++ {
		form.GetFormItem(i + 1).(*tview.Checkbox).SetLabel(operator(1 << i).String()[ui.UserLang])
	}
	renderRoutesEvery(form)
}

func renderRoutes() (form *tview.Form) {
	w, h := consolesize.GetConsoleSize()

	form = tview.NewForm().
		AddInputField(ui.RouteSelectLabel[ui.UserLang], "", 5, nil, func(string) { renderRoutesEvery(form) })

	chosenOp = opMax - 1
	for i := operator(1); i < opMax; i <<= 1 {
		j := i
		form.AddCheckbox(i.String()[ui.UserLang], true, func(checked bool) {
			if checked {
				chosenOp |= j
			} else {
				chosenOp &^= j
			}
			renderRoutesEvery(form)
		})
	}
	form.AddTextView("", "", w, h-3, true, true)

	changeLang := func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 't', 's', 'e':
				renderRoutesLang(form)
			case 'h':
				ui.Pages.SwitchToPage("home")
				ui.UpdateHomepage()
			}
		}
		return event
	}

	for i := 1; i <= opCount; i++ {
		form.GetFormItem(i).(*tview.Checkbox).SetInputCapture(changeLang)
	}

	form.GetFormItem(opCount + 1).(*tview.TextView).
		SetRegions(true).
		SetToggleHighlights(false).
		SetDoneFunc(nil).
		SetChangedFunc(ui.HDraw).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			view := form.GetFormItem(opCount + 1).(*tview.TextView)
			id, _ := strconv.ParseUint(view.GetHighlights()[0], 10, 64)
			i := slices.Index(routeRID, id)
			listLen := len(routeRID)
			switch event.Key() {
			case tcell.KeyDown:
				ui.HighlightScroll(view, strconv.FormatUint(routeRID[(i+1)%listLen], 10))
				return nil
			case tcell.KeyUp:
				ui.HighlightScroll(view, strconv.FormatUint(routeRID[(i+listLen-1)%listLen], 10))
				return nil
			default:
				event = changeLang(event)
			}
			return event
		})
	renderRoutesLang(form)
	return
}
