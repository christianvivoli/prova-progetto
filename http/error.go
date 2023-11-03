package http

import (
	"net/http"

	"prova/app"

	"github.com/labstack/echo/v4"
)

const (
	GenericErrorMessage = "An error occurred"
)

// codes represents an HTTP status code.
var codes = map[string]int{
	app.ECONFLICT:         http.StatusConflict,
	app.EFORBIDDEN:        http.StatusForbidden,
	app.EINVALID:          http.StatusBadRequest,
	app.ENOTFOUND:         http.StatusNotFound,
	app.ENOTIMPLEMENTED:   http.StatusNotImplemented,
	app.EUNAUTHORIZED:     http.StatusUnauthorized,
	app.ENOTAUTHENTICATED: http.StatusUnauthorized,
	app.ESHOULDLOGOUT:     http.StatusUnauthorized,
	app.EINTERNAL:         http.StatusInternalServerError,
	app.EUNAVAILABLE:      http.StatusServiceUnavailable,
}

// MessageFromErr returns the message for the given app error.
// EINTERNAL & EUNKNOWN message is obscured by HTTP response.
func MessageFromErr(err error) string {

	appErrMessage := app.ErrorMessage(err)
	appErrCode := app.ErrorCode(err)

	if appErrMessage == "" || appErrCode == app.EINTERNAL || appErrCode == app.EUNKNOWN {
		return GenericErrorMessage
	}

	return appErrMessage
}

// StatusCodeFromErr returns the HTTP status code for the given app error.
func StatusCodeFromErr(err error) int {

	appErrCode := app.ErrorCode(err)

	code, ok := codes[appErrCode]
	if !ok {
		code = http.StatusInternalServerError
	}
	return code
}

// ErrorAPI represents an error returned by the API.
type ErrorAPI struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details"`
}

// NewErrorAPI returns an ErrorAPI instance.
func NewErrorAPI(err error, details interface{}) *ErrorAPI {
	e := &ErrorAPI{
		Code:    app.ErrorCode(err),
		Message: MessageFromErr(err),
		Details: details,
	}
	return e
}

// ErrorResponseJSON returns an HTTP error response with JSON content.
func ErrorResponseJSON(c echo.Context, err error, details interface{}) error {
	return c.JSON(StatusCodeFromErr(err), NewErrorAPI(err, details))
}

// InvalidRequestErrorJSON restituisce un errore JSON standard per una richiesta non valida
func InvalidRequestErrorJSON(c echo.Context) error {
	return ErrorResponseJSON(c, app.Errorf(app.EINVALID, "Richiesta non valida"), nil)
}
