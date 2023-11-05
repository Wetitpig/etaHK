package GMB

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/Wetitpig/etaHK/common"
	"github.com/Wetitpig/etaHK/ui"
	"github.com/agnivade/levenshtein"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/nathan-fiscaletti/consolesize-go"
	"github.com/rivo/tview"
	"golang.org/x/exp/slices"
)

type routeStop struct {
	id   int
	name ui.Lang
	eta  []eta
}

var (
	regionSelected region
	routeRID       []int
)

const APIBASE = "https://data.etagmb.gov.hk"

type getData struct {
	Data interface{} `json:"data"`
}

func updateTime(sCount int) {
	_, frame := ui.Pages.GetFrontPage()
	ui.App.QueueUpdateDraw(func() {
		frame.(*tview.Frame).Clear().AddText(ui.NextUpdateLabel[ui.UserLang]+strconv.Itoa(sCount), true, tview.AlignLeft, tcell.ColorDefault)
	})
}

func renderRoutesEvery(form *tview.Form) {
	view := form.GetFormItem(2).(*tview.TextView)
	selected := ui.RetainHighlight(view)
	routeRID = routeRID[:0]
	searchStr := form.GetFormItem(1).(*tview.InputField).GetText()
	for i, v := range routeList {
		if strings.Contains(strings.ToUpper(v.code), strings.ToUpper(searchStr)) && v.region == regionSelected.Code() {
			routeRID = append(routeRID, i)
		}
	}
	slices.SortFunc(routeRID, func(i, j int) int {
		i_len := levenshtein.ComputeDistance(searchStr, routeList[i].code)
		j_len := levenshtein.ComputeDistance(searchStr, routeList[j].code)
		if i_len == j_len {
			return strings.Compare(routeList[i].code, routeList[j].code)
		} else {
			return i_len - j_len
		}
	})

	buf := new(ui.TextViewBuffer)
	for _, id := range routeRID {
		v := routeList[id]

		rg := strconv.Itoa(id)
		buf.Write("[\"", rg, "\"]", "[yellow]", runewidth.FillRight(v.code, 5), v.directions[0].orig[ui.UserLang])
		if len(v.directions) > 1 {
			buf.Write("<")
		}
		buf.Write("=>", v.directions[0].dest[ui.UserLang],
			"\n[green]", "\t", v.description[ui.UserLang],
			"\n[-]",
		)
		if rg == selected {
			view.Highlight(selected)
		}
	}
	buf.Print(view)
	if len(view.GetHighlights()) == 0 && len(routeRID) > 0 {
		view.Highlight(strconv.Itoa(routeRID[0]))
	}
}

func renderRoutesLang(form *tview.Form) {
	form.GetFormItem(0).(*tview.DropDown).
		SetLabel(regionSelectLabel[ui.UserLang]).
		SetOptions(regionLabelLang(), func(text string, index int) {
			regionSelected = region(index)
			renderRoutesEvery(form)
		})
	form.GetFormItem(1).(*tview.InputField).SetLabel(ui.RouteSelectLabel[ui.UserLang])

	renderRoutesEvery(form)
}

func renderRoutes() (form *tview.Form) {
	_, h := consolesize.GetConsoleSize()

	form = tview.NewForm().
		AddDropDown(regionSelectLabel[ui.UserLang], regionLabelLang(), 0, nil).
		AddInputField(ui.RouteSelectLabel[ui.UserLang], "", 5, nil, func(string) { renderRoutesEvery(form) }).
		AddTextView("", "", 0, h-7, true, true)

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
	form.GetFormItem(0).(*tview.DropDown).
		SetSelectedFunc(func(text string, index int) {
			regionSelected = region(index)
			renderRoutesEvery(form)
		}).
		SetInputCapture(changeLang)
	form.GetFormItem(2).(*tview.TextView).
		SetRegions(true).
		SetToggleHighlights(false).
		SetDoneFunc(func(key tcell.Key) {
			if key == tcell.KeyEnter {
				i, _ := strconv.Atoi(form.GetFormItem(2).(*tview.TextView).GetHighlights()[0])
				go routeDetail(i, 0, 1)
			}
		}).
		SetChangedFunc(ui.HDraw).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			view := form.GetFormItem(2).(*tview.TextView)
			i, _ := strconv.Atoi(view.GetHighlights()[0])
			i = slices.Index(routeRID, i)
			listLen := len(routeRID)
			switch event.Key() {
			case tcell.KeyDown:
				ui.HighlightScroll(view, strconv.Itoa(routeRID[(i+1)%listLen]))
				return nil
			case tcell.KeyUp:
				ui.HighlightScroll(view, strconv.Itoa(routeRID[(i+listLen-1)%listLen]))
				return nil
			default:
				event = changeLang(event)
			}
			return event
		})
	renderRoutesLang(form)

	return
}

func searchRoutes() (routeMap [][]string) {
	resp, err := http.Get(APIBASE + "/route")
	if err != nil {
		ui.Fatalln("Unable to obtain GMB route list.")
	}
	defer resp.Body.Close()

	var pj getData
	if json.NewDecoder(resp.Body).Decode(&pj) != nil {
		ui.Fatalln("Unable to unmarshal GMB route list.")
	}
	routes := pj.Data.(map[string]interface{})["routes"].(map[string]interface{})

	routeMap = make([][]string, regionCount)
	for i := region(0); i < regionCount; i++ {
		for _, r := range routes[region(i).Code()].([]interface{}) {
			routeMap[i] = append(routeMap[i], r.(string))
		}
	}
	return
}

var (
	regionSelectLabel ui.Lang
	routeList         map[int]*route
)

func getRoute(input <-chan [2]string, res chan<- common.JsonRetMsg[*route]) {
	for r := range input {
		resp, err := http.Get(APIBASE + "/route/" + r[0] + "/" + r[1])
		if err != nil {
			ui.Fatalln("Unable to obtain GMB route info for route", r[1], "in region", r[0])
		}

		var pj getData
		if json.NewDecoder(resp.Body).Decode(&pj) != nil {
			ui.Fatalln("Unable to unmarshal GMB route list.")
		}
		for _, ri := range pj.Data.([]interface{}) {
			routeInfo := ri.(map[string]interface{})
			dirInfo := routeInfo["directions"].([]interface{})

			routeData := route{
				r[1], r[0], formLang(routeInfo, "description"), make([]direction, len(dirInfo)), ui.Lang{
					"[yellow::u]備註[::-]\n", "[yellow::u]备注[::-]\n", "[yellow::u]Notes[::-]\n",
				},
			}

			for _, d := range dirInfo {
				dir := d.(map[string]interface{})
				routeData.directions[int(dir["route_seq"].(float64))-1] = direction{
					formLang(dir, "orig"), formLang(dir, "dest"), formLang(routeInfo, "remarks"), []routeStop{}, fareTable{},
				}
			}
			res <- common.JsonRetMsg[*route]{int(routeInfo["route_id"].(float64)), &routeData}
		}
		resp.Body.Close()
	}
}

func ListGMB() {
	if routeList == nil {
		regionSelected = HKI

		regionSelectLabel = ui.Lang{"地區：", "地区：", "Region:"}
		routes := searchRoutes()

		routeList = make(map[int]*route)

		totalLen := 0
		for i := region(0); i < regionCount; i++ {
			totalLen += len(routes[i])
		}
		result, routeIn := make(chan common.JsonRetMsg[*route], totalLen), make(chan [2]string, totalLen)

		for c := 0; c < common.MAX_CONN; c++ {
			go getRoute(routeIn, result)
		}

		for i, routeReg := range routes {
			for _, routeCode := range routeReg {
				routeIn <- [2]string{region(i).Code(), routeCode}
			}
		}
		close(routeIn)
		for j := 0; j < totalLen; j++ {
			res := <-result
			routeList[res.UID] = res.Ret
		}

		ui.Pages.AddAndSwitchToPage("routesGMB", renderRoutes(), true)
	} else {
		ui.Pages.SwitchToPage("routesGMB")
		_, nf := ui.Pages.GetFrontPage()
		form := nf.(*tview.Form)
		renderRoutesLang(form)
	}
}
