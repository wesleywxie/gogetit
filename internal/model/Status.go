package model

type Status int

const (
	INIT Status = iota
	SKIPPED
	COMPLETED
)

func (s Status) String() string {
	return [...]string{"INIT", "SKIPPED", "COMPLETED"}[s]
}
