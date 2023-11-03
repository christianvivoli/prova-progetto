package http

import (
	_ "embed"
	"io"

	"html/template"

	"prova/app"

	"github.com/labstack/echo/v4"
)

var (

	//go:embed views/layout/base.html
	BaseTemplateHtml string
	//go:embed views/layout/head.html
	HeadTemplateHtml string

	// errors template

	// components template

	// pages templates

	//go:embed views/index.html
	IndexPageTemplateHtml string
	IndexPageTemplate     = template.Must(template.New("index").Parse(BaseTemplateHtml + HeadTemplateHtml + IndexPageTemplateHtml))

	//go:embed views/lista.html
	ListaPageTemplateHTML string
	ListaPageTemplate     = template.Must(template.New("lista").Parse(BaseTemplateHtml + HeadTemplateHtml + ListaPageTemplateHTML))
)

const (
	ErrLoadingPage = "Errore durante il caricamento della pagina"
)

// HeadData definisce i dati dell'head della pagina.
type HeadData struct {
	Title   string
	AppName string
	NoIndex bool
}

// PageTemplateData definisce i dati dell'intera pagina.
type PageTemplateData[T any] struct {
	HeadData    HeadData
	ContentData T
}

// renderPage si occupa di effettuare il render della pagina passata.
func renderPage[T any](wr io.Writer, t *template.Template, data PageTemplateData[T]) error {

	data.HeadData.AppName = app.AppName

	return t.Execute(wr, data)
}

// errorPage restituisce una pagina d'errore.
func errorPage(c echo.Context, httpCode int, msg string) error {
	return c.HTML(httpCode, msg)
}
