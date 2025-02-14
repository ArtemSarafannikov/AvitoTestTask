package cstErrors

import "errors"

var (
	BadRequestDataError = errors.New("Bad request data")
	InternalError       = errors.New("Internal server error")
	NotFoundError       = errors.New("Not found")
	BadCredentialError  = errors.New("Bad credential")
)

func GenerateError(err string) error {
	return errors.New(err)
}
