// Package exitcode defines the canonical exit code constants and the
// ExitError type used to communicate exit codes through the cobra command
// chain back to main().
package exitcode

// Exit code constants matching contracts/exit-codes.md.
const (
	Success          = 0
	GeneralError     = 1
	InvalidArgument  = 2
	NotFound         = 3
	PermissionDenied = 4
	Timeout          = 5
)

// ExitError is a sentinel error that carries a process exit code.
// Commands return ExitError (wrapping the display error) when they need a
// specific non-1 exit code. Execute() in cmd/root.go unwraps it.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return "exit code " + itoa(e.Code)
}

func (e *ExitError) Unwrap() error { return e.Err }

// New returns an *ExitError with the given code and message.
func New(code int, err error) *ExitError {
	return &ExitError{Code: code, Err: err}
}

// itoa is a minimal int-to-string without importing strconv.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := make([]byte, 0, 3)
	for n > 0 {
		buf = append([]byte{byte('0' + n%10)}, buf...)
		n /= 10
	}
	return string(buf)
}
