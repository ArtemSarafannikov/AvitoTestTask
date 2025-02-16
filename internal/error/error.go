package cstErrors

import (
	"errors"
	"github.com/labstack/echo/v4"
	"net/http"
)

type KnownError interface {
	Error() string
	IsKnown() bool
	Code() int
}

type CustomError struct {
	msg  string
	code int
}

func (e *CustomError) Error() string {
	return e.msg
}

func (e *CustomError) Code() int {
	return e.code
}

func (e *CustomError) IsKnown() bool {
	return true
}

var (
	BadRequestDataError       = GenerateError(http.StatusBadRequest, "Bad request data")
	InternalError             = GenerateError(http.StatusInternalServerError, "Internal server error")
	NotFoundError             = GenerateError(http.StatusNotFound, "Not found")
	BadCredentialError        = GenerateError(http.StatusUnauthorized, "Bad credential")
	UnauthorizedError         = GenerateError(http.StatusUnauthorized, "Authorize to this operation")
	NoCoinError               = GenerateError(http.StatusBadRequest, "There are not enough coins in the balance for this operation")
	NoSellingMerchError       = GenerateError(http.StatusBadRequest, "No selling merchant")
	CantSendCoinYourselfError = GenerateError(http.StatusBadRequest, "Cant send coin to yourself")
)

func GenerateError(code int, err string) error {
	return &CustomError{
		msg:  err,
		code: code,
	}
}

func IsCustomError(err error) bool {
	var knownError KnownError
	return errors.As(err, &knownError)
}

func GetAndLogCustomError(err error, logger echo.Logger) KnownError {
	if err == nil {
		return nil
	}
	knErr, ok := err.(KnownError)
	if !ok {
		logger.Error(err)
		return InternalError.(KnownError)
	}
	return knErr
}
