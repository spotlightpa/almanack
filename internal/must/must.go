package must

func Do(err error) {
	if err != nil {
		panic(err)
	}
}

func Get[T any](v T, err error) T {
	if err != nil {
		panic(err)
	}
	return v
}
