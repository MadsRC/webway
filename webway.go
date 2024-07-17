package webway

type Config interface {
	String(key string) string
	Bool(key string) bool
}

type MissingOptionError struct {
	Option string
}

func (e *MissingOptionError) Error() string {
	return "missing option: " + e.Option
}
