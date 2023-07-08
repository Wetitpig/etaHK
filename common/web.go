package common

import (
	"strings"

	"github.com/Wetitpig/etaHK/ui"
)

const MAX_CONN = 16

type JsonRetMsg[T any] struct {
	UID int
	Ret T
}

type GetData struct {
	Data interface{} `json:"data"`
}

func FormLang(obj map[string]interface{}, k string) (c ui.Lang) {
	if v, ok := obj[k+"_tc"]; ok && v != nil {
		c = ui.Lang{
			obj[k+"_tc"].(string), obj[k+"_sc"].(string), obj[k+"_en"].(string),
		}
	} else if v, ok := obj[k+"_c"]; ok && v != nil {
		c = ui.Lang{
			obj[k+"_c"].(string), obj[k+"_s"].(string), obj[k+"_e"].(string),
		}
	} else if v, ok := obj[k+"C"]; ok && v != nil {
		c = ui.Lang{
			obj[k+"C"].(string), obj[k+"S"].(string), obj[k+"E"].(string),
		}
	} else {
		c = ui.Lang{}
	}
	c[2] = strings.Replace(c[2], "'", "’", -1)
	return
}
