package GMB

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/Wetitpig/etaHK/common"
	"github.com/Wetitpig/etaHK/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func renderRouteETA(view *tview.TextView, id, sDir int, end chan<- bool) {
	selected := ui.RetainHighlight(view)
	dir := routeList[id].directions[sDir]
	view.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			close(end)
			i, _ := strconv.Atoi(view.GetHighlights()[0])
			go stopETA(dir.stops[i-1], &stopLoc{id, sDir, i})
		}
	})

	buf := new(ui.TextViewBuffer)
	var etaBuilder strings.Builder
	for i, v := range dir.stops {
		rg := strconv.Itoa(i + 1)
		buf.Write("[\"", rg, "\"][yellow]", rg, "\t", v.name[ui.UserLang],
			"\n[red]\t", renderETA(etaBuilder, v.eta), "\n[-]",
		)
	}
	buf.Print(view)
	view.Highlight(selected)
}

func (r *route) listStops(wg *sync.WaitGroup, id, i int) {
	defer wg.Done()
	resp, err := http.Get(APIBASE + "/route-stop/" + strconv.Itoa(id) + "/" + strconv.Itoa(i+1))
	if err != nil {
		ui.Fatalln("Unable to obtain GMB stop info for route", r.code, "in region", regionSelected.Code())
	}
	defer resp.Body.Close()

	var pj getData
	if json.NewDecoder(resp.Body).Decode(&pj) != nil {
		ui.Fatalln("Unable to unmarshal GMB stop list for route", r.code, "in region", regionSelected.Code())
	}

	rs := pj.Data.(map[string]interface{})["route_stops"].([]interface{})
	r.directions[i].stops = make([]routeStop, len(rs))
	for _, v := range rs {
		s := v.(map[string]interface{})
		r.directions[i].stops[int(s["stop_seq"].(float64))-1] = routeStop{
			id:   int(s["stop_id"].(float64)),
			name: formLang(s, "name"),
		}
	}
}

func (r *route) queueRouteETA(queue <-chan [3]int, etaC chan<- common.JsonRetMsg[[]eta]) {
	for k := range queue {
		id, i, msg := k[0], k[1], k[2]
		if resp, err := http.Get(APIBASE + "/eta/route-stop/" + strconv.Itoa(id) + "/" + strconv.Itoa(msg+1) + "/" + strconv.Itoa(i+1)); err == nil {
			var pj getData
			if json.NewDecoder(resp.Body).Decode(&pj) == nil {
				re := pj.Data.(map[string]interface{})
				if re["enabled"].(bool) {
					etaC <- common.JsonRetMsg[[]eta]{i, parseETA(re)}
				} else {
					etaC <- common.JsonRetMsg[[]eta]{i, []eta{}}
				}
			}
			resp.Body.Close()
		}
	}
}
