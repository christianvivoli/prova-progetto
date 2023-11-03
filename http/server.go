package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"prova/app"
	"strconv"
	"time"

	log "github.com/inconshreveable/log15"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/acme/autocert"
)

// ShutdownTimeout is the time given for outstanding requests to finish before shutdown.
const ShutdownTimeout = 1 * time.Second

// ServerAPI is the main server for the API
type ServerAPI struct {
	ln net.Listener
	// server is the main server for the API
	server *http.Server

	// handler is the main handler for the API
	handler *echo.Echo

	// Addr Bind address for the server.
	Addr string
	// Domain name to use for the server.
	// If specified, server is run on TLS using acme/autocert.
	Domain string
	// BaseURL defines a base url to return as public endpoint
	BaseURL string

	UserService app.UserService

	// loggin service used by HTTP Server.
	LogService log.Logger
}

func NewServerAPI() *ServerAPI {

	s := &ServerAPI{
		server:  &http.Server{},
		handler: echo.New(),
	}

	// Set echo as the default HTTP handler.
	s.server.Handler = s.handler

	s.handler.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "ciao")
	})

	s.handler.GET("/", s.handlerIndexPage)

	s.handler.POST("/lista", s.handlerListaPage)

	return s
}

// Close closes the server with graceful shutdown.
func (s *ServerAPI) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), ShutdownTimeout)
	defer cancel()
	return s.server.Shutdown(ctx)
}

// Port returns the TCP port for the running server.
// This is useful in tests where we allocate a random port by using ":0".
func (s *ServerAPI) Port() int {
	if s.ln == nil {
		return 0
	}
	return s.ln.Addr().(*net.TCPAddr).Port
}

// Open validates the server options and start it on the bind address.
func (s *ServerAPI) Open() (err error) {

	if s.Domain != "" {
		s.ln = autocert.NewListener(s.Domain)
	} else {
		if s.ln, err = net.Listen("tcp", s.Addr); err != nil {
			return err
		}
	}

	go s.server.Serve(s.ln)

	return nil
}

// Scheme returns the scheme used by the server.
func (s *ServerAPI) Scheme() string {
	if s.Domain != "" {
		return "https"
	}
	return "http"
}

// URL returns the URL for the server.
// This is useful in tests where we allocate a random port by using ":0".
func (s *ServerAPI) URL() string {

	if s.BaseURL != "" {
		return s.BaseURL
	}

	scheme, port := s.Scheme(), s.Port()

	domain := "localhost"

	if (scheme == "http" && port == 80) || (scheme == "https" && port == 443) {
		return fmt.Sprintf("%s://%s", scheme, domain)
	}

	return fmt.Sprintf("%s://%s:%d", scheme, domain, port)
}

// UseTLS returns true if the server is using TLS.
func (s *ServerAPI) UseTLS() bool {
	return s.Domain != ""
}

// SuccessResponseJSON returns a JSON response with the given status code and data.
func SuccessResponseJSON(c echo.Context, httpCode int, data interface{}) error {
	if data == nil {
		return c.NoContent(httpCode)
	}
	return c.JSON(httpCode, data)
}

// ListenAndServeTLSRedirect runs an HTTP server on port 80 to redirect users
// to the TLS-enabled port 443 server.
func ListenAndServeTLSRedirect(domain string) error {
	return http.ListenAndServe(":80", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "https://"+domain, http.StatusFound)
	}))
}

// PaginateResponse rappresent a generic paginate response.
type PaginateResponse[T any] struct {
	Data         []T `json:"data"`
	TotalResults int `json:"total_results"`
	CurrentPage  int `json:"current_page"`
	ItemsPerPage int `json:"items_per_page"`
}

// NewPaginateResponse creates a new PaginateResponse.
func NewPaginateResponse[T any](data []T, totalResults int, page int, itemsPerPage int) PaginateResponse[T] {
	return PaginateResponse[T]{
		Data:         data,
		TotalResults: totalResults,
		CurrentPage:  page,
		ItemsPerPage: itemsPerPage,
	}
}

// DownloadFileConfig definisce i parametri per far scaricare un file.
type DownloadFileConfig struct {
	Filename   string
	MimeType   string
	StatusCode *int   // default 200
	MaxAge     *int64 // default int64((30 * 24 * time.Hour).Seconds())
}

// Validate si occupa di validare le impostazioni passate.
func (opt DownloadFileConfig) Validate() error {

	if opt.Filename == "" {
		return app.Errorf(app.EINTERNAL_INVALID, "Nome file obbligatorio")
	}

	if opt.MimeType == "" {
		return app.Errorf(app.EINTERNAL_INVALID, "MimeType obbligatorio")
	}

	return nil
}

// DownloadFile si occupa di scrivere nella response i campi per permettere di scaricar un file.
func DownloadFile(c echo.Context, b []byte, opt DownloadFileConfig) error {

	if err := opt.Validate(); err != nil {
		return err
	}

	statusCode := http.StatusOK
	maxAge := int64((30 * 24 * time.Hour).Seconds())

	if opt.StatusCode != nil {
		statusCode = *opt.StatusCode
	}
	if opt.MaxAge != nil {
		maxAge = *opt.MaxAge
	}

	c.Response().Status = statusCode
	c.Response().Header().Set(echo.HeaderContentType, opt.MimeType)
	c.Response().Header().Set(echo.HeaderContentDisposition, fmt.Sprintf("inline; filename=\"%s\"", opt.Filename))
	c.Response().Header().Set(echo.HeaderContentLength, strconv.Itoa(len(b)))
	c.Response().Header().Set(echo.HeaderCacheControl, fmt.Sprintf("private; max-age=%d", maxAge))
	c.Response().Write(b)

	return nil
}
