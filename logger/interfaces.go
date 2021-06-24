package logger

import (
	"github.com/milonoir/rv/common"
)

type Logger interface {
	common.Widget

	Messages() []string
}
