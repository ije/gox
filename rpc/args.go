package rpc

import (
	"encoding/gob"
	"reflect"
)

type RPCArguments map[string]interface{}

func (args RPCArguments) Get(key, argType string) (arg interface{}, ok bool) {
	if arg, ok = args[key]; ok {
		switch argType {
		case "string":
			arg, ok = arg.(string)
		case "int":
			arg, ok = arg.(int)
		case "int64":
			arg, ok = arg.(int64)
		case "bytes":
			arg, ok = arg.([]byte)
		case "ints":
			arg, ok = arg.([]int)
		case "strings":
			arg, ok = arg.([]string)
		case "map":
			arg, ok = arg.(map[string]interface{})
		default:
			var rt reflect.Type
			if rt, ok = RPCArgumentTypes[argType]; !ok || !rt.Implements(reflect.TypeOf(arg)) {
				ok = false
				arg = nil
			}
		}
	}
	return
}

var RPCArgumentTypes = map[string]reflect.Type{}

func RegisterRPCArgumentType(name string, v interface{}) {
	gob.Register(v)
	RPCArgumentTypes[name] = reflect.TypeOf(v)
}
