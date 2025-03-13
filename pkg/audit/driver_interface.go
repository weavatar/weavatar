package audit

type Driver interface {
	Check(url string) (bool, string, error)
}
