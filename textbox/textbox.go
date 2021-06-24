package textbox

import (
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
)

type textBox struct {
	*widgets.Paragraph
}

func NewTextBox(title string) *textBox {
	p := widgets.NewParagraph()
	p.Title = title

	return &textBox{
		Paragraph: p,
	}
}

func (t *textBox) Update() {
	ui.Render(t)
}

func (t *textBox) Resize(x1, y1, x2, y2 int) {
	t.SetRect(x1, y1, x2, y2)
}

func (t *textBox) Close() {}

func (t *textBox) SetText(s string) {
	t.Text = s
}
