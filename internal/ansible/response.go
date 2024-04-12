package ansible

import (
	"encoding/json"
	"io"
	"os"
)

type Response struct {
	Message string `json:"message"`
	Changed bool   `json:"changed"`
	Failed  bool   `json:"failed"`
}

func Changed(message string) Response {
	return Response{
		Message: message,
		Changed: true,
	}
}

func Unchanged(message string) Response {
	return Response{
		Message: message,
		Changed: false,
	}
}

func Failed(err error) Response {
	return Response{
		Message: err.Error(),
		Changed: false,
		Failed:  true,
	}
}

func Fail(err error) error {
	return Return(Failed(err))
}

func Change(message string, err error) error {
	if err != nil {
		return Fail(err)
	}

	return Return(Changed(message))
}

func NoChange(message string) error {
	return Return(Unchanged(message))
}

func Return(r Response) error {
	return ReturnWriter(r, os.Stdout)
}

func ReturnWriter(r Response, w io.Writer) error {
	defer func() {
		if r.Failed {
			os.Exit(1)
		}

		os.Exit(0)
	}()

	return json.NewEncoder(w).Encode(r)
}
