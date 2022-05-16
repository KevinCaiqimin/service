package router

import (
	"encoding/json"
	"fmt"
	"reflect"
	"sync"

	"github.com/valyala/fastjson"
)

var r sync.Map

func Reg(mod interface{}) {
	tp := reflect.TypeOf(mod)

	for i := 0; i < tp.NumMethod(); i++ {
		mt := tp.Method(i)
		r.Store(mt.Name, mt)
	}
}

func Unref(mod interface{}) {
	tp := reflect.TypeOf(mod)
	for i := 0; i < tp.NumMethod(); i++ {
		mt := tp.Method(i)
		r.Delete(mt.Name)
	}
}

func Call(name string, json_data string) ([]byte, error) {
	v, ok := r.Load(name)
	if !ok {
		return nil, fmt.Errorf("404")
	}
	mt := v.(reflect.Method)
	js_v, err := fastjson.Parse(json_data)
	if err != nil {
		return nil, err
	}
	tp := mt.Type

	args := make([]reflect.Value, tp.NumIn())
	for i := 0; i < tp.NumIn(); i++ {
		arg_tp := tp.In(i)
		arg_name := arg_tp.Name()
		switch arg_tp.Kind() {
		case reflect.Int:
			args[i] = reflect.ValueOf(js_v.GetInt(arg_name))
		case reflect.String:
			args[i] = reflect.ValueOf((string)(js_v.GetStringBytes(arg_name)))
		case reflect.Struct:
			args[i] = parse_object(arg_tp, js_v.Get(arg_name))
		default:
			return nil, fmt.Errorf("method=%v, param=%s unsupported arg type %v", name, arg_name, arg_tp.String())
		}
	}
	ret := mt.Func.Call(args)
	js_bytes, err := json.Marshal(ret[0])
	if err != nil {
		return nil, err
	}
	return js_bytes, nil
}
