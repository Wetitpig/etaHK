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

		var j, k, d int
		fareNotes.Reset()
		for _, v := range rows {
			w := strings.Split(v, " · ")
			for c := range w {
				w[c] = strings.TrimSpace(w[c])
			}
			switch len(w) {
			case 1:
				j++
				fareNotes.WriteString(strconv.Itoa(j))
				fareNotes.WriteString(". ")
				fareNotes.WriteString(v)
				fareNotes.WriteRune('\n')
			case 2:
				for e, dir := range rt.directions {
					if strings.Contains(dir.orig[i], w[1]) {
						d, k = e, 0
						break
					}
				}
				fallthrough
			default:
				w = w[1:]
				if i == 0 {
					rt.directions[d].fareTable = append(rt.directions[d].fareTable, make([]ui.Lang, len(w)))
				}
				for l, x := range w {
					rt.directions[d].fareTable[k][l][i] = x
				}
				k++
			}
		}
		if j == 0 {
			rt.fareNotes[i] = ""
		} else {
			rt.fareNotes[i] += fareNotes.String()
		}
	}
}

func (dir *direction) renderRouteFares(fn ui.Lang, t *tview.Grid) {
	fareTableLabel := ui.Lang{"車資表", "车资表", "Faretable"}
	rc := len(dir.fareTable)

	t.Clear().AddItem(tview.NewTextView().
		SetText("[::u]"+fareTableLabel[ui.UserLang]+"[::-]").
		SetDynamicColors(true).
		SetTextAlign(tview.AlignCenter).
		SetWrap(true),
		0, 0, 1, rc, 0, 0, false)

	origHeights := make([]int, 0, rc)
	r := 2
	for _, row := range dir.fareTable {
		for c, td := range row {
			text := td[ui.UserLang]
			cell := tview.NewTextView().
				SetText(text).
				SetTextColor(tcell.ColorYellow).
				SetWrap(true)
			colSpan := 1
			if c == len(row)-1 {
				cell.SetTextStyle(tcell.StyleDefault.Bold(true))
				colSpan = rc - c
				origHeights = append(origHeights, runewidth.StringWidth(text))
			} else {
				cell.SetTextAlign(tview.AlignCenter)
			}
			t.AddItem(cell,
				r, c, 1, colSpan, 0, 0, false)
		}
		r++
	}

	if len(fn) > 0 {
		textView := tview.NewTextView().
			SetText(fn[ui.UserLang]).
			SetChangedFunc(ui.HDraw).
			SetDynamicColors(true).
			SetWrap(true)
		t.AddItem(textView,
			r, 0, 1, rc, 0, 0, false)
		origHeights = append(origHeights, tview.TaggedStringWidth(fn[ui.UserLang]), strings.Count(fn[ui.UserLang], "\n"))
	}

	delete(ui.BeforeDrawFn, "renderRouteFares")
	ui.BeforeDrawFn["renderRouteFares"] = func(tcell.Screen) bool {
		if name, pages := ui.Pages.GetFrontPage(); name == "routeGMB" {
			if name, _ = pages.(*tview.Flex).GetItem(2).(*tview.Pages).GetFrontPage(); name == "fare" {
				w, _ := consolesize.GetConsoleSize()
				rc := len(dir.fareTable)

				height := make([]int, 2, rc+2)
				height[0], height[1] = 1, 1
				for i, width := range origHeights[:rc] {
					rw := (w / rc * (rc - i))
					height = append(height, (width+rw-1)/rw)
				}
				if last := rc; len(origHeights) > last {
					height = append(height, (origHeights[last]+w-1)/w+origHeights[last+1])
				}
				t.SetRows(height...)
			}
		}
		return false
	}
}
