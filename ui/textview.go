package ui

import (
	"bytes"

	"github.com/rivo/tview"
)

var (
	App   *tview.Application
	Pages *tview.Pages
)

const RefreshInterval = 30

func HighlightScroll(view *tview.TextView, rid string) {
	view.Highlight(rid)
	view.ScrollToHighlight()
}

func HDraw() {
	App.Draw()
}

func RetainHighlight(view *tview.TextView) (selected string) {
	s := view.GetHighlights()
	selected = ""
	if len(s) > 0 {
		selected = s[0]
	}
	view.Clear().Highlight()
	return
}

type TextViewBuffer struct {
	buf bytes.Buffer
}

func (t *TextViewBuffer) Write(input ...string) {
	for _, s := range input {
		t.buf.WriteString(s)
	}
}

func (t *TextViewBuffer) Print(view *tview.TextView) {
	view.Write(t.buf.Bytes())
}
