package http

import (
	"bytes"
	"net/http"

	"prova/app"
	"github.com/labstack/echo/v4"
)

func (s *ServerAPI) handlerIndexPage(c echo.Context) error {

	var buf bytes.Buffer

	if err := renderPage(&buf, IndexPageTemplate, PageTemplateData[map[string]any]{}); err != nil {
		app.LogErr(s.LogService, err)
		return errorPage(c, http.StatusInternalServerError, ErrLoadingPage)
	}
	return c.HTML(http.StatusOK, buf.String())
}
