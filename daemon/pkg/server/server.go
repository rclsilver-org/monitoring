package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/loopfz/gadgeto/tonic"
	"github.com/sirupsen/logrus"
	"github.com/wI2L/fizz"
	"github.com/wI2L/fizz/openapi"
)

type ServerOptions func(opts *config)

func WithVerbose(verbose bool) func(opts *config) {
	return func(opts *config) {
		opts.Verbose = verbose
	}
}

func WithTitle(title string) func(opts *config) {
	return func(opts *config) {
		opts.Title = title
	}
}

func WithVersion(version string) func(opts *config) {
	return func(opts *config) {
		opts.Version = version
	}
}

type Server struct {
	cfg    *config
	router *fizz.Fizz
}

func NewServer(ctx context.Context, opts ...ServerOptions) (*Server, error) {
	// load the configuration
	cfg, err := loadConfig()
	if err != nil {
		return nil, fmt.Errorf("unable to load the configuration: %w", err)
	}
	logrus.WithContext(ctx).Debug("loaded the HTTP server configuration")

	// apply the configuration modifiers
	for _, modifier := range opts {
		modifier(cfg)
	}

	// set the gin release mode when verbose mode is disabled
	if !cfg.Verbose {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.UseRawPath = true

	infos := &openapi.Info{
		Title:   cfg.Title,
		Version: cfg.Version,
	}

	router := fizz.NewFromEngine(engine)
	router.GET("/spec.json", nil, router.OpenAPI(infos, "json"))

	mon := router.Group("/mon", "monitoring", "monitoring of the API")
	{
		mon.GET("/ping", []fizz.OperationOption{
			fizz.Summary("Checks if the API is healthy"),
			fizz.Response(fmt.Sprint(http.StatusInternalServerError), "Server Error", APIError{}, nil, nil),
		}, tonic.Handler(monPing, http.StatusOK))
	}

	tonic.SetErrorHook(errorHook)

	return &Server{
		cfg:    cfg,
		router: router,
	}, nil
}

func (s *Server) RegisterGroup(path, name, description string) *fizz.RouterGroup {
	return s.router.Group(path, name, description)
}

func (s *Server) Serve(ctx context.Context) error {
	endpoint := fmt.Sprintf("%s:%d", s.cfg.ListenHost, s.cfg.ListenPort)
	srv := &http.Server{Addr: endpoint, Handler: s.checkAllowedSources(withLogging(s.router))}

	go func() {
		logrus.WithContext(ctx).Debugf("starting the HTTP server on %s", endpoint)

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logrus.WithContext(ctx).WithError(err).Fatal("unable to start the HTTP server")
		}

		logrus.WithContext(ctx).Debug("stopped the HTTP server")
	}()

	select {
	case <-ctx.Done():
		logrus.WithContext(ctx).Debug("stopping the HTTP server")
		break
	}

	return srv.Shutdown(context.Background())
}

func (s *Server) checkAllowedSources(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		remote_address := getRemoteAddress(r)

		if len(s.cfg.AllowedSources) > 0 {
			allowed := false

			for _, ip := range s.cfg.AllowedSources {
				if ip == remote_address {
					allowed = true
					break
				}
			}

			if !allowed {
				logrus.WithContext(r.Context()).Debugf("request rejected from %q", remote_address)
				http.Error(w, "not allowed", http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
