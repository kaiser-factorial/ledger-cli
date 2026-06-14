package exit

import (
	"errors"
	"fmt"
	"os"
)

const (
	Success    = 0
	AuthError  = 10
	NotFound   = 20
	Network    = 30
	Validation = 40
	Other      = 50
)

// ExitError carries both a message and an exit code.
type ExitError struct {
	Code    int
	Message string
}

func (e *ExitError) Error() string {
	return e.Message
}

// New creates a new ExitError with the given code and message.
func New(code int, msg string) *ExitError {
	return &ExitError{Code: code, Message: msg}
}

// Exit prints the error to stderr and exits with the specified code.
func Exit(err error) {
	if err == nil {
		os.Exit(Success)
	}

	var exitErr *ExitError
	if errors.As(err, &exitErr) {
		fmt.Fprintln(os.Stderr, exitErr.Message)
		os.Exit(exitErr.Code)
	}

	// Default to "Other" error
	fmt.Fprintln(os.Stderr, err)
	os.Exit(Other)
}