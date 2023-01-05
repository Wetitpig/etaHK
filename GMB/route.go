package GMB

import (
	"strconv"
	"sync"
	"time"

	"github.com/Wetitpig/etaHK/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var nextUpdateLabel ui.Lang

func initRouteDetail() (newFlex *tview.Flex, view *tview.TextView) {
	if !ui.Pages.HasPage("routeGMB") {
		view = tview.NewTextView().
			SetChangedFunc(ui.HDraw).
			SetDynamicColors(true).
			SetScrollable(true).
			SetRegions(true).
			SetToggleHighlights(false)
		newFlex = tview.NewFlex().
			SetDirection(tview.FlexRow).
			AddItem(tview.NewTextView().
				SetChangedFunc(ui.HDraw), 0, 5, false).
			AddItem(tview.NewTextView().
				SetChangedFunc(ui.HDraw).
				SetTextAlign(tview.AlignCenter), 0, 20, false).
			AddItem(tview.NewPages().
				AddAndSwitchToPage("eta", view, true).
				AddPage("fare", tview.NewGrid().
					SetBorders(false).
					SetGap(0, 0), true, false),
				0, 75, true)
		ui.Pages.AddAndSwitchToPage("routeGMB", newFlex, true)
	} else {
		ui.Pages.SwitchToPage("routeGMB")
		_, nf := ui.Pages.GetFrontPage()
		newFlex = nf.(*tview.Flex)
		newFlex.GetItem(2).(*tview.Pages).SwitchToPage("eta")
		_, nv := newFlex.GetItem(2).(*tview.Pages).GetFrontPage()
		view = nv.(*tview.TextView)
	}
	return
}

func routeDetail(index, selectedDir int, seq int) {
	{
		var wg sync.WaitGroup
		if routeList[index].directions[0].fareTable == nil {
			wg.Add(1)
			go getFareTable(&wg, index)
		}
		for i, v := range routeList[index].directions {
			if len(v.stops) == 0 {
				wg.Add(1)
				go routeList[index].listStops(&wg, index, i)
			}
		}
		wg.Wait()
	}
	var sCount int

	etaChan, stopChan := make(chan int), make(chan bool, 1)

	nextUpdateLabel = ui.Lang{"下次更新：", "下回更新：", "Next update: "}
	newFlex, view := initRouteDetail()
	newFlex.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune {
			switch event.Rune() {
			case 'r':
				selectedDir = (selectedDir + 1) % len(routeList[index].directions)
				etaChan <- selectedDir
				stopChan <- true
			case 'f':
				pages := newFlex.GetItem(2).(*tview.Pages)
				name, _ := pages.GetFrontPage()
				if name == "eta" {
					pages.SwitchToPage("fare")
				} else if name == "fare" {
					pages.SwitchToPage("eta")
				}
				stopChan <- false
			case 't', 's', 'e':
				stopChan <- true
			case 'h':
				close(stopChan)
				ui.Pages.SwitchToPage("routesGMB")
				_, form := ui.Pages.GetFrontPage()
				renderRoutesLang(form.(*tview.Form))
			}
		}
		return event
	})
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		i, _ := strconv.Atoi(view.GetHighlights()[0])
		listLen := len(routeList[index].directions[selectedDir].stops)
		switch event.Key() {
		case tcell.KeyDown:
			ui.HighlightScroll(view, strconv.Itoa(i%listLen+1))
			return nil
		case tcell.KeyUp:
			ui.HighlightScroll(view, strconv.Itoa((i+listLen-2)%listLen+1))
			return nil
		}
		return event
	})
	stopChan <- true

	go func() {
		for {
			if msg, ok := <-etaChan; ok {
				var (
					etaLock sync.Mutex
					etaGp   sync.WaitGroup
				)
				etaGp.Add(len(routeList[index].directions[msg].stops))
				for i := range routeList[index].directions[msg].stops {
					go routeList[index].queueRouteETA(&etaGp, &etaLock, index, i, msg)
				}
				etaGp.Wait()
				stopChan <- false
			} else {
				return
			}
		}
	}()

	tick := time.NewTicker(time.Second)
	view.Highlight(strconv.Itoa(seq))
	for {
		select {
		case v, ok := <-stopChan:
			if ok {
				name, elem := newFlex.GetItem(2).(*tview.Pages).GetFrontPage()
				if v {
					updateTime(newFlex, sCount)
					newFlex.GetItem(1).(*tview.TextView).SetText(
						routeList[index].directions[selectedDir].orig[ui.UserLang] +
							"\n=>\n" +
							routeList[index].directions[selectedDir].dest[ui.UserLang],
					)
				}

				if name == "eta" {
					renderRouteETA(elem.(*tview.TextView), index, selectedDir, stopChan)
				} else if name == "fare" {
					routeList[index].directions[selectedDir].renderRouteFares(routeList[index].fareNotes, elem.(*tview.Grid))
				}
			} else {
				close(etaChan)
				tick.Stop()
				return
			}
		case <-tick.C:
			if sCount == 0 {
				etaChan <- selectedDir
				sCount = ui.RefreshInterval
			}
			sCount--
			updateTime(newFlex, sCount)
		}
	}
}
