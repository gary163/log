package log

import (
	"testing"
)

func TestConsole_WriteMsg(t *testing.T) {
	SetLogger(AdapterConsole,`{"color":true}`)
	Info("info")
	Warning("warning")
	Error("error")
}
