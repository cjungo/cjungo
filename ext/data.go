package ext

import (
	"reflect"
)

// 只有类型严格匹配才会转化。
func MoveField[TF any, TT any](from TF, to TT) {
	rf := reflect.ValueOf(from).Elem()
	rt := reflect.ValueOf(to).Elem()
	tt := rt.Type()
	for i := 0; i < tt.NumField(); i++ {
		tName := tt.Field(i).Name
		tType := tt.Field(i).Type

		ff := rf.FieldByName(tName)
		if ff.IsValid() && ff.Type() == tType {
			tf := rt.FieldByName(tName)
			tf.Set(ff)
		}
	}
}

// 类型如果可以转化，就会转化。
func MoveFieldEx[TF any, TT any](from TF, to TT) {
	rf := reflect.ValueOf(from).Elem()
	rt := reflect.ValueOf(to).Elem()
	tt := rt.Type()
	for i := 0; i < tt.NumField(); i++ {
		tName := tt.Field(i).Name
		tType := tt.Field(i).Type

		ff := rf.FieldByName(tName)
		if ff.IsValid() {
			if ff.Type() == tType {
				tf := rt.FieldByName(tName)
				tf.Set(ff)
			} else if ff.CanConvert(tType) {
				tf := rt.FieldByName(tName)
				tf.Set(ff.Convert(tType))
			}
		}
	}
}
