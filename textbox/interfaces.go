package textbox

import (
	"github.com/milonoir/rv/common"
)

type TextBox interface {
	common.Widget

	SetText(string)
}
