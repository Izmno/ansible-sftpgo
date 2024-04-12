package ansible

import (
	"encoding/json"
	"errors"
	"os"
)

var (
	ErrInvalidArgs = errors.New("no argument file provided")
)

func Read[T any]() *T {
	if len(os.Args) != 2 {
		Fail(ErrInvalidArgs)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		Fail(err)
	}

	defer f.Close()

	a := new(T)

	if err := json.NewDecoder(f).Decode(a); err != nil {
		Fail(err)
	}

	return a
}
