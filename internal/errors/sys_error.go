package err

type SystemError string

const (
	SystemErrorShortCodeAlreadyInUse SystemError = "SystemErrorShortCodeAlreadyInUse"
	SystemErrorShortCodeExpired      SystemError = "SystemErrorShortCodeExpired"
)

// Returns true if the provided error is a specific system error.
func IsSystemError(err error, systemError SystemError) bool {
	return err != nil && err.Error() == string(systemError)
}
