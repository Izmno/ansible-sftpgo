package ansible

import "errors"

type State string

const (
	StatePresent State = "present"
	StateAbsent  State = "absent"
)

var ErrInvalidState = errors.New("invalid state")

func (s State) Validate() error {
	switch s {
	case StatePresent, StateAbsent:
		return nil
	default:
		return ErrInvalidState
	}
}
