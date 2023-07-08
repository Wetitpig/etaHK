package GMB

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Wetitpig/etaHK/common"
	"github.com/Wetitpig/etaHK/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"golang.org/x/exp/slices"
)

type stop struct {
	stops map[stopLoc]*routeStop
	keys  []stopLoc
}

func (stopT *stop) renderStopName(view *tview.TextView) {
	var stopName string
	var stopBuilder strings.Builder
	for _, route := range stopT.stops {
		if partName := route.name[ui.UserLang]; len(stopName) == 0 || !strings.Contains(stopName, partName) {
			stopBuilder.WriteString("\n" + partName)
		}
		stopName = stopBuilder.String()[1:]
	}
	view.SetText(stopName)
}

func (data *stop) renderStopETA(view *tview.TextView, end chan<- bool) {
	selected := ui.RetainHighlight(view)
	view.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			k := new(stopLoc)
			k.unmarshal(view.GetHighlights()[0])
			close(end)
			go routeDetail(k.route_id, k.route_seq, k.stop_seq)
		}
	})

	buf := new(ui.TextViewBuffer)
	var etaBuilder strings.Builder
	for _, id := range data.keys {
		v := data.stops[id]
		rCode := routeList[id.route_id]
		direction := rCode.directions[id.route_seq]

		buf.Write("[\"", id.marshal(), "\"][yellow]", rCode.code, "\t", direction.orig[ui.UserLang], " => ", direction.dest[ui.UserLang],
			"\n[green]\t", rCode.description[ui.UserLang],
			"\n[red]\t", renderETA(etaBuilder, v.eta),
			"\n[-]",
		)
	}
	buf.Print(view)
	view.Highlight(selected).ScrollToHighlight()
}

func (stopT stop) queueStopETA(msg int) {
	if resp, err := http.Get(APIBASE + "/eta/stop/" + strconv.Itoa(msg)); err == nil {
		defer resp.Body.Close()
		var pj common.GetData
		if json.NewDecoder(resp.Body).Decode(&pj) == nil {
			for _, s := range pj.Data.([]interface{}) {
				sj := s.(map[string]interface{})
				k := formStopLoc(sj)
				stopT.stops[k].eta = stopT.stops[k].eta[:0]
				if sj["enabled"].(bool) {
					stopT.stops[k].eta = parseETA(sj)
				}
			}

			slices.SortFunc(stopT.keys, func(iid, jid stopLoc) bool {
				if len(stopT.stops[iid].eta) > 0 && len(stopT.stops[jid].eta) == 0 {
					return true
				} else if len(stopT.stops[jid].eta) > 0 && len(stopT.stops[iid].eta) == 0 {
					return false
				} else if routeList[iid.route_id].code != routeList[jid.route_id].code {
					return routeList[iid.route_id].code < routeList[jid.route_id].code
				} else if iid.route_seq != jid.route_seq {
					return iid.route_seq < jid.route_seq
				} else {
					return iid.stop_seq < jid.stop_seq
				}
			})
		}
	}
}

func (s *stop) listStopRoutes(id int) {
	resp, err := http.Get(APIBASE + "/stop-route/" + strconv.Itoa(id))
	if err != nil {
		ui.Fatalln("Unable to obtain GMB stop info for stop", id)
	}
	defer resp.Body.Close()
	var pj common.GetData
	if json.NewDecoder(resp.Body).Decode(&pj) != nil {
		ui.Fatalln("Unable to unmarshal GMB stop info for stop", id)
	}
	for _, st := range pj.Data.([]interface{}) {
		sr := st.(map[string]interface{})
		k := formStopLoc(sr)
		s.stops[k] = &routeStop{
			0, common.FormLang(sr, "name"), []eta{},
		}
		s.keys = append(s.keys, k)
	}
}

func initStopETA() (newFlex *tview.Flex) {
	if !ui.Pages.HasPage("stopGMB") {
		newFlex = tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(tview.NewTextView().
				SetChangedFunc(ui.HDraw).
				SetTextAlign(tview.AlignCenter), 0, 10, false).
			AddItem(tview.NewTextView().
				SetChangedFunc(ui.HDraw).
				SetDynamicColors(true).
				SetScrollable(true).
				SetRegions(true).
				SetToggleHighlights(false), 0, 90, true)
		ui.Pages.AddAndSwitchToPage("stopGMB", tview.NewFrame(newFlex).SetBorders(0, 0, 0, 0, 0, 0), true)
	} else {
		ui.Pages.SwitchToPage("stopGMB")
		_, nf := ui.Pages.GetFrontPage()
		newFlex = nf.(*tview.Frame).GetPrimitive().(*tview.Flex)
	}
	return
}

func stopETA(s routeStop, id *stopLoc) {
	etaChan := make(chan int)
	printChan := make(chan bool, 1)

	var sCount int
	stopT := &stop{make(map[stopLoc]*routeStop), []stopLoc{}}
	stopT.listStopRoutes(s.id)

	newFlex := initStopETA()
	newFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		view := newFlex.GetItem(1).(*tview.TextView)
		k := new(stopLoc)
		k.unmarshal(view.GetHighlights()[0])
		i := slices.Index(stopT.keys, *k)
		listLen := len(stopT.keys)
		switch event.Key() {
		case tcell.KeyDown:
			ui.HighlightScroll(view, stopT.keys[(i+1)%listLen].marshal())
			return nil
		case tcell.KeyUp:
			ui.HighlightScroll(view, stopT.keys[(i+listLen-1)%listLen].marshal())
			return nil
		case tcell.KeyRune:
			switch event.Rune() {
			case 't', 's', 'e':
				printChan <- true
			case 'b':
				close(printChan)
				ui.Pages.SwitchToPage("routesGMB")
				_, form := ui.Pages.GetFrontPage()
				renderRoutesLang(form.(*tview.Form))
			case 'h':
				close(printChan)
				ui.Pages.SwitchToPage("home")
			}
		}
		return event
	})
	printChan <- true

	go func() {
		for msg := range etaChan {
			stopT.queueStopETA(msg)
			printChan <- false
		}
	}()

	tick := time.NewTicker(time.Second)
	newFlex.GetItem(1).(*tview.TextView).Highlight(id.marshal())
	for {
		select {
		case v, ok := <-printChan:
			if ok {
				if v {
					updateTime(sCount)
					stopT.renderStopName(newFlex.GetItem(0).(*tview.TextView))
				}
				stopT.renderStopETA(newFlex.GetItem(1).(*tview.TextView), printChan)
			} else {
				close(etaChan)
				tick.Stop()
				return
			}
		case <-tick.C:
			if sCount == 0 {
				etaChan <- s.id
				sCount = ui.RefreshInterval
			}
			sCount--
			updateTime(sCount)
		}
	}
}
