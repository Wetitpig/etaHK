package GMB

import (
	"strings"
	"time"

	"github.com/Wetitpig/etaHK/ui"
)

type eta struct {
	time    time.Time
	remarks ui.Lang
}

func parseETA(obj map[string]interface{}) (etaSlice []eta) {
	etaSeq := obj["eta"].([]interface{})
	etaSlice = make([]eta, len(etaSeq))
	for _, v := range etaSeq {
		etaD := v.(map[string]interface{})
		if t, e := time.Parse(time.RFC3339, etaD["timestamp"].(string)); e == nil {
			etaSlice[int(etaD["eta_seq"].(float64))-1] = eta{t, formLang(etaD, "remarks")}
		} else {
			return
		}
	}
	return
}

func renderETA(builder strings.Builder, etaSlice []eta) string {
	var etaString string
	if len(etaSlice) > 0 {
		for _, e := range etaSlice {
			builder.WriteString(", " + e.time.Format("15:04"))
			if e.remarks[ui.UserLang] != "" {
				builder.WriteString(" (" + e.remarks[ui.UserLang] + ")")
			}
		}
		etaString = builder.String()[2:]
		builder.Reset()
	}
	return etaString
}
