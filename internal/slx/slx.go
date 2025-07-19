package slx

func One[T any](v T) []T {
	return []T{v}
}

func Make[T any](vs ...T) []T {
	return vs
}

func First[T any](slice []T) T {
	var (
		first T
		count = len(slice)
	)
	if count == 0 {
		return first
	}
	return slice[0]
}

func Last[T any](slice []T) T {
	var (
		last  T
		count = len(slice)
	)
	if count == 0 {
		return last
	}
	return slice[count-1]
}
