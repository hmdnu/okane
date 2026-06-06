package lib

import "errors"

type FormError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

var NotFoundErr = errors.New("resource not found")

func NewNotFoundErr(err error, msg string) *string {

	if errors.Is(err, NotFoundErr) {
		return &msg
	}

	return nil
}
