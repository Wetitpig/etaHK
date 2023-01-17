package Bus

import (
	"regexp"
	"strings"

	"github.com/Wetitpig/etaHK/ui"
	"github.com/mattn/go-runewidth"
)

type operator uint8

const opCount = 5

const (
	KMB operator = 1 << iota
	LWB
	CTB
	NWFB
	NLB
	opMax
)

func (i operator) String() ui.Lang {
	switch i {
	case KMB:
		return ui.Lang{"九巴", "九巴", "KMB"}
	case LWB:
		return ui.Lang{"龍運", "龙运", "LWB"}
	case CTB:
		return ui.Lang{"城巴", "城巴", "CTB"}
	case NWFB:
		return ui.Lang{"新巴", "新巴", "NWFB"}
	case NLB:
		return ui.Lang{"嶼巴", "屿巴", "NLB"}
	default:
		return ui.Lang{}
	}
}

func (r subroute) Color() string {
	var formatStr strings.Builder
	switch r.Op {
	case KMB, KMB + CTB, KMB + NWFB:
		formatStr.WriteString("[#E32222]")
	case LWB:
		formatStr.WriteString("[#FF5000]")
	case CTB, CTB + LWB:
		formatStr.WriteString("[white:#0080FF]")
	case NWFB:
		formatStr.WriteString("[white:#6C3F99]")
	case NLB:
		formatStr.WriteString("[#53D0C2]")
	}
	opList := ""
	for op := operator(1); op < opMax; op <<= 1 {
		if o := r.Op & op; o != 0 {
			opList += o.String()[ui.UserLang] + "/"
		}
	}
	formatStr.WriteString(runewidth.FillRight(opList[:len(opList)-1], 10))

	formatStr.WriteString("[-:-]")
	switch r.Code[0] {
	case 'N':
		formatStr.WriteString("[yellow]")
	case 'A':
		formatStr.WriteString("[white:#B11116]")
	default:
		if regexp.MustCompile(`[136]\d{2}`).FindString(r.Code) != "" {
			formatStr.WriteString("[white:red]")
		} else if regexp.MustCompile(`9\d{2}`).FindString(r.Code) != "" {
			formatStr.WriteString("[white:#009140]")
		}
	}
	return formatStr.String()
}
