package ui

type Lang = [3]string

type Language int

var UserLang Language

const (
	TC Language = iota
	SC
	EN
)

func (l Language) String() string {
	switch l {
	case TC:
		return "TC"
	case SC:
		return "SC"
	case EN:
		return "EN"
	default:
		return ""
	}
}
