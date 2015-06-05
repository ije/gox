package rpc

import (
	"encoding/gob"
	"fmt"
	"reflect"
	"sync"
)

type RPCArguments map[string]interface{}

func (args RPCArguments) Get(key, argType string) (arg interface{}, ok bool) {
	if arg, ok = args[key]; ok {
		switch argType {
		case "int":
			arg, ok = arg.(int)
		case "int8":
			arg, ok = arg.(int8)
		case "int16":
			arg, ok = arg.(int16)
		case "int32":
			arg, ok = arg.(int32)
		case "int64":
			arg, ok = arg.(int64)

		case "uint":
			arg, ok = arg.(uint)
		case "uint8":
			arg, ok = arg.(uint8)
		case "uint16":
			arg, ok = arg.(uint16)
		case "uint32":
			arg, ok = arg.(uint32)
		case "uint64":
			arg, ok = arg.(uint64)

		case "float32":
			arg, ok = arg.(float32)
		case "float64":
			arg, ok = arg.(float64)

		case "complex64":
			arg, ok = arg.(complex64)
		case "complex128":
			arg, ok = arg.(complex128)

		case "uintptr":
			arg, ok = arg.(uintptr)
		case "bool":
			arg, ok = arg.(bool)
		case "string":
			arg, ok = arg.(string)

		case "[]byte":
			arg, ok = arg.([]byte)
		case "[]int":
			arg, ok = arg.([]int)
		case "[]int8":
			arg, ok = arg.([]int8)
		case "[]int16":
			arg, ok = arg.([]int16)
		case "[]int32":
			arg, ok = arg.([]int32)
		case "[]int64":
			arg, ok = arg.([]int64)
		case "[]uint":
			arg, ok = arg.([]uint)
		case "[]uint8":
			arg, ok = arg.([]uint8)
		case "[]uint16":
			arg, ok = arg.([]uint16)
		case "[]uint32":
			arg, ok = arg.([]uint32)
		case "[]uint64":
			arg, ok = arg.([]uint64)
		case "[]float32":
			arg, ok = arg.([]float32)
		case "[]float64":
			arg, ok = arg.([]float64)
		case "[]complex64":
			arg, ok = arg.([]complex64)
		case "[]complex128":
			arg, ok = arg.([]complex128)
		case "[]uintptr":
			arg, ok = arg.([]float64)
		case "[]bool":
			arg, ok = arg.([]bool)
		case "[]string":
			arg, ok = arg.([]string)

		default:
			var rt reflect.Type
			if rt, ok = nameToConcreteType[argType]; !ok || rt != reflect.TypeOf(arg) {
				arg, ok = nil, false
			}
		}
	}
	return
}

var (
	registerLock       sync.RWMutex
	nameToConcreteType = make(map[string]reflect.Type)
	concreteTypeToName = make(map[reflect.Type]string)
)

func RegisterName(name string, v interface{}) {
	if name == "" {
		panic("attempt to register empty name")
	}

	registerLock.Lock()
	defer registerLock.Unlock()

	vt := reflect.TypeOf(v)
	if t, ok := nameToConcreteType[name]; ok && t != vt {
		panic(fmt.Sprintf("rpc: registering duplicate types for %q: %s != %s", name, t, vt))
	}
	if n, ok := concreteTypeToName[vt]; ok && n != name {
		panic(fmt.Sprintf("rpc: registering duplicate names for %s: %q != %q", vt, n, name))
	}

	nameToConcreteType[name] = vt
	concreteTypeToName[vt] = name

	gob.RegisterName(name, v)
}

func Register(v interface{}) {
	rt := reflect.TypeOf(v)
	name := rt.String()

	star := ""
	if rt.Name() == "" {
		if pt := rt; pt.Kind() == reflect.Ptr {
			star = "*"
			rt = pt
		}
	}
	if rt.Name() != "" {
		if rt.PkgPath() == "" {
			name = star + rt.Name()
		} else {
			name = star + rt.PkgPath() + "." + rt.Name()
		}
	}
	RegisterName(name, v)
}

func init() {
	RegisterName("[]interface{}", []interface{}(nil))
	RegisterName("map[string]interface{}", map[string]interface{}(nil))
}
