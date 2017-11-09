package exec

import (
	"bytes"
	"testing"
)

func TestCMD(t *testing.T) {
	buf := bytes.NewBufferString("(c) 2016 gox\n")
	err := CMD("lessc", "-").Input("@base: #ffffff;\nh1{color:@base;}").Input([]byte("h2{color:@base;width: 100+2px;}")).Output(buf)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(buf.String())
}
