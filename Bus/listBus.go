package Bus

import (
	"encoding/gob"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Wetitpig/etaHK/ui"
	"github.com/dimchansky/utfbom"
	"github.com/dustin/go-humanize"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type stopEntry struct {
	RId, SId int
	SSeq     uint16
	RSeq, UD uint8
	Name     ui.Lang
	Fare     float64
}

type subroute struct {
	Operator   operator
	Code       string
	Orig, Dest ui.Lang
	Direction  [][]*stopEntry
}
type stop struct {
	Name           []ui.Lang
	StopEntryIndex []*stopEntry
}
type busFile struct {
	RouteIndex map[int]subroute
	StopIndex  map[int]stop
	Stops      []stopEntry
}

var busList busFile

const (
	routeSrc  = "https://static.data.gov.hk/td/routes-fares-geojson/JSON_BUS.json"
	cacheFile = "etaHKBus.bin"
)

func formLang(obj map[string]interface{}, k string) (c ui.Lang) {
	if v, ok := obj[k+"C"]; ok && v != nil {
		c = ui.Lang{
			obj[k+"C"].(string), obj[k+"S"].(string), strings.Replace(obj[k+"E"].(string), "'", "’", -1),
		}
	} else {
		c = ui.Lang{}
	}
	return
}

func createBF(buttonIndex int, buttonLabel string) {
	if buttonIndex == 0 {
		os.Remove(cacheFile)

		resp, err := http.Get(routeSrc)
		if err != nil {
			log.Fatalln("Unable to obain bus route list")
		}
		defer resp.Body.Close()

		var features struct {
			Features []struct {
				Properties map[string]interface{} `json:"properties"`
			} `json:"features"`
		}
		if err := json.NewDecoder(utfbom.SkipOnly(resp.Body)).Decode(&features); err != nil {
			log.Fatalln("Unable to unmarshal bus route list")
		}

		bF := busFile{
			make(map[int]subroute),
			make(map[int]stop),
			make([]stopEntry, 0, len(features.Features)),
		}
		for _, f := range features.Features {
			bFentry := stopEntry{
				RId:  int(f.Properties["routeId"].(float64)),
				RSeq: uint8(f.Properties["routeSeq"].(float64)) - 1,
				SSeq: uint16(f.Properties["stopSeq"].(float64)) - 1,
				UD:   uint8(f.Properties["stopPickDrop"].(float64)),
				SId:  int(f.Properties["stopId"].(float64)),
				Name: formLang(f.Properties, "stopName"),
				Fare: f.Properties["fullFare"].(float64),
			}

			var bFop operator
			for i := operator(1); i < opMax; i <<= 1 {
				if strings.Contains(f.Properties["companyCode"].(string), i.String()[2]) {
					bFop += i
				}
			}
			if bFop != 0 {
				bF.Stops = append(bF.Stops, bFentry)
				bF.buildStopIndex()
				bF.buildRouteIndex(bFop, f.Properties)
			}
		}
		for k, v := range bF.RouteIndex {
			if len(v.Direction[1]) == 0 {
				v.Direction = v.Direction[:1:1]
				bF.RouteIndex[k] = v
			}
		}

		f, _ := os.Create(cacheFile)
		if gob.NewEncoder(f).Encode(bF) != nil {
			log.Fatalln("Failed to encode bus list value.")
		}
		f.Close()

		busList = bF
		ui.Pages.AddAndSwitchToPage("routesBus", renderRoutes(), true)
	}
	ui.Pages.HidePage("downloadBus")
}

func initBus() {
	if stat, err := os.Stat(cacheFile); err == nil {
		if time.Since(stat.ModTime()) <= 168*time.Hour {
			readBus()
			return
		}
	}

	header, err := http.Head(routeSrc)
	if err != nil {
		log.Fatalln("Unable to obain bus route list")
	}
	header.Body.Close()
	size, _ := strconv.ParseUint(header.Header["Content-Length"][0], 10, 64)

	sizeS := humanize.Bytes(size)
	downloadLabel := ui.Lang{
		"需要下載 " + sizeS + " 更新數據庫。",
		"需要下载 " + sizeS + " 更新数据库。",
		sizeS + " would be downloaded to update bus route data.",
	}
	buttons := ui.Lang{
		"繼續 取消", "继续 取消", "Confirm Cancel",
	}
	modal := tview.NewModal().
		SetText(downloadLabel[ui.UserLang]).
		AddButtons(strings.Fields(buttons[ui.UserLang])).SetDoneFunc(createBF)

	modal.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyRune && (event.Rune() == 't' || event.Rune() == 's' || event.Rune() == 'e') {
			modal.SetText(downloadLabel[ui.UserLang])
			modal.
				ClearButtons().
				AddButtons(strings.Fields(buttons[ui.UserLang])).SetDoneFunc(createBF)
		}
		return event
	})

	ui.Pages.AddPage("downloadBus", modal, false, true)
	ui.Pages.SendToFront("downloadBus")
}

func readBus() {
	file, err := os.Open(cacheFile)
	if err != nil {
		log.Fatalln("Unable to open cache file for reading.")
	}

	if gob.NewDecoder(file).Decode(&busList) != nil {
		log.Fatalln("Failed to decode bus list value.")
	}
	file.Close()
	ui.Pages.AddAndSwitchToPage("routesBus", renderRoutes(), true)
}

func ListBus() {
	if busList.Stops == nil {
		initBus()
	} else {
		ui.Pages.SwitchToPage("routesBus")
		_, nf := ui.Pages.GetFrontPage()
		form := nf.(*tview.Form)
		renderRoutesLang(form)
	}
}
