package assert

import "reflect"

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}

	vv, ok := v.(reflect.Value)
	if !ok {
		vv = reflect.ValueOf(v)
	}

	switch vv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.Interface, reflect.Slice:
		return vv.IsNil()
	}

	return false
}
