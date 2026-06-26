package lcl

func Filter[T any, S ~[]T](xs S, pred func(T) bool) S {
	res := make(S, 0, len(xs))
	for _, x := range xs {
		if pred(x) {
			res = append(res, x)
		}
	}
	return res
}

func Filter2[T any, S ~[]T](xs S, pred func(int, T) bool) S {
	res := make(S, 0, len(xs))
	for i, x := range xs {
		if pred(i, x) {
			res = append(res, x)
		}
	}
	return res
}

func FilterError[T any, S ~[]T](xs S, pred func(T) (bool, error)) (S, error) {
	res := make(S, 0, len(xs))
	for _, x := range xs {
		ok, err := pred(x)
		if err != nil {
			return res, err
		}
		if ok {
			res = append(res, x)
		}
	}
	return res, nil
}

func Exclude[T any, S ~[]T](xs S, pred func(T) bool) S {
	res := make(S, 0, len(xs))
	for _, x := range xs {
		if !pred(x) {
			res = append(res, x)
		}
	}
	return res
}

func Exclude2[T any, S ~[]T](xs S, pred func(int, T) bool) S {
	res := make(S, 0, len(xs))
	for i, x := range xs {
		if !pred(i, x) {
			res = append(res, x)
		}
	}
	return res
}

func ExcludeError[T any, S ~[]T](xs S, pred func(T) (bool, error)) (S, error) {
	res := make(S, 0, len(xs))
	for _, x := range xs {
		ok, err := pred(x)
		if err != nil {
			return res, err
		}
		if !ok {
			res = append(res, x)
		}
	}
	return res, nil
}

func Map[T, R any, S ~[]T](xs S, mapper func(T) R) []R {
	res := make([]R, len(xs))
	for i, x := range xs {
		res[i] = mapper(x)
	}
	return res
}

func Map2[T, R any, S ~[]T](xs S, mapper func(int, T) R) []R {
	res := make([]R, len(xs))
	for i, x := range xs {
		res[i] = mapper(i, x)
	}
	return res
}

func MapError[T, R any, S ~[]T](xs S, mapper func(T) (R, error)) ([]R, error) {
	res := make([]R, 0, len(xs))
	for _, x := range xs {
		r, err := mapper(x)
		if err != nil {
			return res, err
		}
		res = append(res, r)
	}
	return res, nil
}

func Fold[T, R any, S ~[]T](xs S, accum func(R, T) R, initial R) R {
	res := initial
	for _, x := range xs {
		res = accum(res, x)
	}
	return res
}

func Fold2[T, R any, S ~[]T](xs S, accum func(R, int, T) R, initial R) R {
	res := initial
	for i, x := range xs {
		res = accum(res, i, x)
	}
	return res
}

func FoldError[T, R any, S ~[]T](xs S, accum func(R, T) (R, error), initial R) (R, error) {
	res := initial
	for _, x := range xs {
		r, err := accum(res, x)
		if err != nil {
			return res, err
		}
		res = r
	}
	return res, nil
}

func FoldRight[T, R any, S ~[]T](xs S, accum func(R, T) R, initial R) R {
	res := initial
	for i := len(xs) - 1; i >= 0; i-- {
		res = accum(res, xs[i])
	}
	return res
}

func FoldRight2[T, R any, S ~[]T](xs S, accum func(R, int, T) R, initial R) R {
	res := initial
	for i := len(xs) - 1; i >= 0; i-- {
		res = accum(res, i, xs[i])
	}
	return res
}

func FoldRightError[T, R any, S ~[]T](xs S, accum func(R, T) (R, error), initial R) (R, error) {
	res := initial
	for i := len(xs) - 1; i >= 0; i-- {
		r, err := accum(res, xs[i])
		if err != nil {
			return res, err
		}
		res = r
	}
	return res, nil
}

func ForEach[T any, S ~[]T](xs S, f func(T)) {
	for _, x := range xs {
		f(x)
	}
}

func ForEach2[T any, S ~[]T](xs S, f func(int, T)) {
	for i, x := range xs {
		f(i, x)
	}
}

func ForEachWhile[T any, S ~[]T](xs S, pred func(T) bool) {
	for _, x := range xs {
		if !pred(x) {
			return
		}
	}
}

func GroupBy[T any, R comparable, S ~[]T](xs S, mapper func(T) R) map[R]S {
	res := map[R]S{}
	for _, x := range xs {
		key := mapper(x)
		res[key] = append(res[key], x)
	}
	return res
}

func PartitionBy[T any, R comparable, S ~[]T](xs S, mapper func(T) R) []S {
	res := []S{}
	indices := map[R]int{}
	for _, x := range xs {
		k := mapper(x)
		if i, ok := indices[k]; ok {
			res[i] = append(res[i], x)
		} else {
			indices[k] = len(res)
			res = append(res, S{x})
		}
	}
	return res
}

func FrequenciesBy[T any, R comparable, S ~[]T](xs S, mapper func(T) R) map[R]int {
	freq := make(map[R]int)
	for _, x := range xs {
		freq[mapper(x)]++
	}
	return freq
}

func Pad[T any, S ~[]T](xs S, n int) S {
	if n < 0 {
		panic("Pad: n must not be negative")
	}
	if len(xs) >= n {
		return xs
	}
	r := make(S, n)
	copy(r, xs)
	return r
}

func PadWith[T any, S ~[]T](xs S, n int, padding T) S {
	out := Pad(xs, n)
	for i := len(xs); i < n; i++ {
		out[i] = padding
	}
	return out
}
