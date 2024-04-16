package model

import (
	"errors"
	"fmt"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrAlreadyExists  = errors.New("already exists")
	ErrNotFound       = errors.New("not found")
	ErrBadCredentials = errors.New("wrong credentials")
)

type ValidationError struct {
	err    error
	msgRus string
}

func NewValidationError(field, msg, msgRus string) *ValidationError {
	return &ValidationError{
		err:    fmt.Errorf("%s: %s", field, msg),
		msgRus: msgRus,
	}
}

func (v *ValidationError) Error() string {
	if v.err != nil {
		return v.err.Error()
	}
	return ""
}

func (v *ValidationError) WithDetails(code codes.Code) error {
	statusWithDetails, _ := status.New(code, v.err.Error()).WithDetails(&errdetails.ErrorInfo{
		Metadata: map[string]string{
			"rus": v.msgRus,
		}},
	)

	return statusWithDetails.Err()
}

func WrapErrorWithMethodName(err error, msg string) error {
	return fmt.Errorf("error calling %s: %w", msg, err)
}
