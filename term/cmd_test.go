package term

import (
	"bytes"
	"testing"
)

func TestCMD(t *testing.T) {
	buf := bytes.NewBufferString("(c) 2016 gox\n")
	err := CMD("lessc", "-").Input("@base: #f938ab;\nh1{color:@base;}").Output(buf)
	if err != nil {
		t.Error(err)
	}

	t.Log(buf.String())
}
