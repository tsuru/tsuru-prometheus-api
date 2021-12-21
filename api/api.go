package api

import (
	"fmt"
	"io"
	"net/http"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/prometheus/prometheus/pkg/rulefmt"
	"github.com/tsuru/tsuru-prometheus-api/service"
)

type Server struct {
	echoInstance *echo.Echo
}

type ServerOpts struct {
	Service      service.Service
	AuthUser     string
	AuthPassword string
}

func NewServer(opts ServerOpts) *Server {

	echoInstance := echo.New()
	echoInstance.Debug = true
	echoInstance.Use(
		// Logger
		middleware.Logger(),

		// Auth
		middleware.BasicAuthWithConfig(middleware.BasicAuthConfig{
			Validator: func(username, password string, c echo.Context) (bool, error) {
				return true, nil
			},
			Skipper: func(c echo.Context) bool {
				return c.Path() == "/"
			},
		}),
	)
	echoInstance.GET("/", func(c echo.Context) error {
		return c.String(200, "Tsuru Prometheus API is running")
	})

	echoInstance.GET("/v1/pools/:pool/rules/:rule", func(c echo.Context) error {
		return c.String(200, "Not implemented yet")
	})

	echoInstance.PUT("/v1/pools/:pool/rules/:rule", func(c echo.Context) error {
		contentType := c.Request().Header.Get("Content-Type")
		if contentType != "application/x-yaml" {
			return c.String(400, "Content-Type must be application/x-yaml")
		}

		data, err := io.ReadAll(c.Request().Body)
		if err != nil {
			return err
		}

		ruleGroups, errs := rulefmt.Parse(data)
		if len(errs) > 0 {
			return c.String(400, fmt.Sprintf("Errors: %v", errs))
		}

		err = opts.Service.EnsurePrometheusRule(c.Param("pool"), c.Param("rule"), *ruleGroups)
		if err != nil {
			return err
		}
		return c.NoContent(http.StatusNoContent)
	})

	return &Server{
		echoInstance: echoInstance,
	}
}

// Run starts the server
func (s *Server) Run() error {
	return s.echoInstance.Start(":8888")
}
