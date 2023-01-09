package GMB

import (
	"net/http"
	"strconv"
	"strings"
	"sync"

	htmlmd "github.com/JohannesKaufmann/html-to-markdown"
	"github.com/JohannesKaufmann/html-to-markdown/plugin"
	"github.com/PuerkitoBio/goquery"
	"github.com/Wetitpig/etaHK/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/mattn/go-runewidth"
	"github.com/nathan-fiscaletti/consolesize-go"
	"github.com/rivo/tview"
)

type fareTable struct {
	breaks []ui.Lang
	fare   [][]string
}

func getFareTable(wg *sync.WaitGroup, id int) {
	defer wg.Done()
	rt := routeList[id]
	var fareChan [3]chan *goquery.Selection
	for i := range fareChan {
		fareChan[i] = make(chan *goquery.Selection)
		go func(lang int) {
			resp, err := http.Get("https://h2-app-rr.hkemobility.gov.hk/ris_page/get_gmb_detail.php?route_id=" + strconv.Itoa(id) + "&lang=" + ui.Language(lang).String())
			if err != nil {
				ui.Fatalln("Cannot obtain fare table for GMB route", id)
			}
			defer resp.Body.Close()
			html, err := goquery.NewDocumentFromReader(resp.Body)
			if err != nil {
				ui.Fatalln("Cannot parse fare table for GMB route", id)
			}

			fareTableHtml := html.Find(".table_hd1")
			fareTableHtml.EachWithBreak(func(i int, s *goquery.Selection) bool {
				if t := s.Text(); t == "Faretable" || t == "车资表" || t == "車資表" {
					fareTableHtml = s.
						Parent(). // tr
						Parent(). // tbody
						Parent(). // table
						Parent(). // td
						Parent()  // tr
					return false
				}
				return true
			})
			fareTableHtml = fareTableHtml.Next(). // tr
								Children().First(). // td
								Children().First(). // table
								Children().First(). // tbody
								Children().First(). // tr
								Children().First()  // td
			fareChan[lang] <- fareTableHtml
			close(fareChan[lang])
		}(i)
	}

	converter := htmlmd.NewConverter("", true, nil)
	converter.Use(plugin.TableCompat())
	var fareNotes strings.Builder
	for i := 0; i < 3; i++ {
		html := <-fareChan[i]
		md := converter.Convert(html)
		md = strings.Replace(md, "\n\n", "\n", -1)
		rows := strings.Split(md, "\n")

		var (
			j    bool
			k, d int
		)
		fareNotes.Reset()
		for _, v := range rows {
			w := strings.Split(v, " · ")
			for c := range w {
				w[c] = strings.TrimSpace(w[c])
			}
			switch len(w) {
			case 1:
				if !strings.Contains(fareNotes.String(), v) {
					j = true
					fareNotes.WriteString(v)
					fareNotes.WriteString("\n\n")
				}
			case 2:
				d = -1
				for e, dir := range rt.directions {
					if strings.Contains(strings.ToLower(dir.orig[i]), strings.ToLower(w[1])) {
						d, k = e, 0
						break
					}
				}
				if d < 0 {
					continue
				}
				fallthrough
			default:
				if d >= 0 {
					if i == 0 {
						rt.directions[d].fareTable.fare = append(rt.directions[d].fareTable.fare, w[1:len(w)-1])
						rt.directions[d].fareTable.breaks = append(rt.directions[d].fareTable.breaks, ui.Lang{w[len(w)-1]})
					} else {
						rt.directions[d].fareTable.breaks[k][i] = w[len(w)-1]
						k++
					}
				}
			}
		}
		if !j {
			rt.fareNotes[i] = ""
		} else {
			rt.fareNotes[i] += fareNotes.String()
		}
	}
}

func (dir *direction) renderRouteFares(fn ui.Lang, t *tview.Grid) {
	fareTableLabel := ui.Lang{"車資表", "车资表", "Faretable"}
	rc := len(dir.fareTable.breaks)

	t.Clear().AddItem(tview.NewTextView().
		SetText("[::u]"+fareTableLabel[ui.UserLang]+"[::-]").
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetWrap(true),
		0, 0, 1, rc, 0, 0, false)

	origHeights := make([]int, 0, rc)
	r := 2
	for i, row := range dir.fareTable.fare {
		for c, td := range row {
			t.AddItem(
				tview.NewTextView().
					SetText(td).
					SetTextColor(tcell.ColorYellow).
					SetWrap(true).
					SetTextAlign(tview.AlignCenter),
				r, c, 1, 1, 0, 0, false)
		}
		b := dir.fareTable.breaks[i][ui.UserLang]
		t.AddItem(
			tview.NewTextView().
				SetText(b).
				SetTextColor(tcell.ColorYellow).
				SetWrap(true).
				SetTextStyle(tcell.StyleDefault.Bold(true)),
			r, i, 1, rc-i, 0, 0, false)
		origHeights = append(origHeights, runewidth.StringWidth(b))
		r++
	}

	notesView := tview.NewTextView().
		SetChangedFunc(ui.HDraw).
		SetDynamicColors(true).
		SetWrap(true)
	if len(fn) > 0 {
		notesView.SetText(fn[ui.UserLang])
		t.AddItem(notesView, r, 0, 1, rc, 0, 0, false)
		origHeights = append(origHeights, tview.TaggedStringWidth(fn[ui.UserLang]), strings.Count(fn[ui.UserLang], "\n"))
	}

	height := make([]int, rc+3)
	t.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyUp:
			r, c := notesView.GetScrollOffset()
			if r == 0 {
				r, c = t.GetOffset()
				t.SetOffset(r-1, c)
			} else {
				notesView.ScrollTo(r-1, c)
			}
			return nil
		case tcell.KeyDown:
			r, c := t.GetOffset()
			if r == len(dir.fareTable.breaks)+2 {
				r, c = notesView.GetScrollOffset()
				notesView.ScrollTo(r+1, c)
			} else {
				t.SetOffset(r+1, c)
			}
			return nil
		}
		return event
	})
	delete(ui.BeforeDrawFn, "renderRouteFares")
	ui.BeforeDrawFn["renderRouteFares"] = func(tcell.Screen) bool {
		if name, pages := ui.Pages.GetFrontPage(); name == "routeGMB" {
			if name, _ = pages.(*tview.Frame).GetPrimitive().(*tview.Flex).GetItem(1).(*tview.Pages).GetFrontPage(); name == "fare" {
				w, _ := consolesize.GetConsoleSize()
				rc := len(dir.fareTable.breaks)

				height[0], height[1] = 1, 1
				for i, width := range origHeights[:rc] {
					rw := (w / rc * (rc - i))
					height[i+2] = (width + rw - 1) / rw
				}
				if last := rc; len(origHeights) > last {
					height[rc+2] = (origHeights[last]+w-1)/w + origHeights[last+1]
					t.SetRows(height...)
				} else {
					t.SetRows(height[:rc+2]...)
				}
			}
		}
		return false
	}
}
