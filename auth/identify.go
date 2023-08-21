package auth

type IdentifyState uint

const (
	Vaild IdentifyState = iota
	Invaild
	Expired
	Error
)

type Identify interface {
	Check() (uint64, IdentifyState, error)
	Refresh(string) error
}
