package ext

import "reflect"

func MoveField[TF any, TT any](from TF, to TT) {
	rf := reflect.ValueOf(from).Elem()
	rt := reflect.ValueOf(to).Elem()
	tt := rt.Type()
	for i := 0; i < tt.NumField(); i++ {
		tName := tt.Field(i).Name
		tType := tt.Field(i).Type

		ff := rf.FieldByName(tName)
		if ff.IsValid() {
			fft := ff.Type()
			if fft == tType {
				tf := rt.FieldByName(tName)
				tf.Set(ff)
			}
		}
	}
}
