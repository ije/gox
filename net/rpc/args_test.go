package rpc

import (
	"bytes"
	"encoding/gob"
	"testing"
)

func TestArgs(t *testing.T) {
	raw := map[string]interface{}{
		"b":  []byte("hello world"),
		"m":  map[string]interface{}{"hello": "world"},
		"m2": map[string]int{"abc": 123},
	}

	Register(map[string]int(nil))
	buffer := bytes.NewBuffer(nil)
	t.Log(gob.NewEncoder(buffer).Encode(raw), buffer)

	var args RPCArguments
	t.Log(gob.NewDecoder(buffer).Decode(&args), args)

	for _, at := range []string{"[]byte", "[]uint8", "[]int"} {
		v, ok := args.Get("b", at)
		t.Logf("%v(%T) %s(%v)", v, v, at, ok)
	}

	for _, at := range []string{"map[string]interface{}", "map[string]interface {}"} {
		v, ok := args.Get("m", at)
		t.Logf("%v(%T) %s(%v)", v, v, at, ok)
	}

	for _, at := range []string{"map[string]int", "map[string]interface {}"} {
		v, ok := args.Get("m2", at)
		t.Logf("%v(%T) %s(%v)", v, v, at, ok)
	}
}
