package http

import (
	"bytes"
	"net/http"
	"strconv"

	"prova/app"

	"github.com/labstack/echo/v4"
)

func (s *ServerAPI) handlerListaPage(c echo.Context) error {

	phone, err := strconv.ParseInt(c.FormValue("phone"), 10, 64)
	if err != nil {
		return err
	}

	crt := app.UserCreate{
		Name:     c.FormValue("name"),
		Surname:  c.FormValue("surname"),
		Email:    c.FormValue("email"),
		Password: c.FormValue("password"),
		Phone:    phone,
	}

	if _, err := s.UserService.CreateUser(c.Request().Context(), crt); err != nil {
		app.LogErr(s.LogService, err)
		return errorPage(c, http.StatusInternalServerError, ErrLoadingPage)
	}

	users, _, err := s.UserService.FindUsers(c.Request().Context(), app.UserFilter{})
	if err != nil {
		app.LogErr(s.LogService, err)
		return errorPage(c, http.StatusInternalServerError, ErrLoadingPage)
	}

	var buf bytes.Buffer

	if err := renderPage(&buf, ListaPageTemplate, PageTemplateData[map[string]any]{
		ContentData: map[string]any{
			"Users": users,
		},
	}); err != nil {
		app.LogErr(s.LogService, err)
		return errorPage(c, http.StatusInternalServerError, ErrLoadingPage)
	}
	return c.HTML(http.StatusOK, buf.String())
}
