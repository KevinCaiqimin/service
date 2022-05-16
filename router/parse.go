package router

import (
	"reflect"

	"github.com/valyala/fastjson"
)

func parse_object(tp reflect.Type, js_v *fastjson.Value) reflect.Value {
	obj := reflect.New(tp)
	for i := 0; i < tp.NumField(); i++ {
		fv := obj.Field(i)
		ftp := fv.Type()
		k := fv.String()
		switch ftp.Kind() {
		case reflect.Int:
			fv.SetInt(int64(js_v.GetInt64(k)))
		case reflect.String:
			fv.SetString((string)(js_v.GetStringBytes(k)))
		case reflect.Struct:
			fv.Set(parse_object(ftp, js_v.Get(k)))
		}
	}
	return obj
}
