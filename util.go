package cjungo

func GetOrDefault[T any](v *T, d T) T {
	if v != nil {
		return *v
	}
	return d
}
