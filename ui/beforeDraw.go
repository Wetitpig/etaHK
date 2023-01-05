package ui

import "github.com/gdamore/tcell/v2"

var BeforeDrawFn map[string]func(tcell.Screen) bool

func RunBeforeDraw(screen tcell.Screen) bool {
	res := false
	for _, f := range BeforeDrawFn {
		res = res || f(screen)
	}
	return res
}
