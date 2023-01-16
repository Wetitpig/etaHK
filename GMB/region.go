package GMB

import "github.com/Wetitpig/etaHK/ui"

type region int

const (
	HKI region = iota
	KLN
	NT
	regionCount
)

func (r region) String() ui.Lang {
	switch r {
	case HKI:
		return ui.Lang{"香港島", "香港岛", "HK Island"}
	case KLN:
		return ui.Lang{"九龍", "九龙", "Kowloon"}
	case NT:
		return ui.Lang{"新界", "新界", "N.T."}
	default:
		return ui.Lang{}
	}
}

func (r region) Code() string {
	switch r {
	case HKI:
		return "HKI"
	case KLN:
		return "KLN"
	case NT:
		return "NT"
	default:
		return ""
	}
}

func regionLabelLang() (r []string) {
	r = make([]string, 0, regionCount)
	for i := region(0); i < regionCount; i++ {
		r = append(r, region(i).String()[ui.UserLang])
	}
	return
}
