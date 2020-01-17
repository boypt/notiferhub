package common

// Must ...
func Must(err error) {
	if err != nil {
		panic(err)
	}
}

// Must2 ...
func Must2(v interface{}, err error) interface{} {
	Must(err)
	return v
}

// Error2 ...
func Error2(v interface{}, err error) error {
	return err
}
