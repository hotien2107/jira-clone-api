package error

var (
	ErrResourceNotFound  = "Resource not found"
	ErrResourceDuplicate = "Resource is duplicated"

	ErrTokenRequired    = "Token is required"
	ErrTokenWrongFormat = "Token is wrong format"
	ErrTokenWrong       = "Token is wrong"
	ErrTokenRevoked     = "Token is revoked"

	ErrUrlNotFound            = "URL not found"
	ErrQueryMethodNotAllowed  = "Query method not allowed"
	ErrQueryByFieldNotAllowed = "Query by field not allowed"
	ErrValueIsNotAccepted     = "Value is not accepted"
	ErrFieldWrongType         = "Field wrong type"

	ErrTransactionExpired = "Transaction is expired"
)
