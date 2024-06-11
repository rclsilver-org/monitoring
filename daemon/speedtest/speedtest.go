package speedtest

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/loopfz/gadgeto/tonic"
	"github.com/sirupsen/logrus"
	"github.com/wI2L/fizz"

	"github.com/rclsilver/monitoring/daemon/pkg/component"
	"github.com/rclsilver/monitoring/daemon/pkg/server"
)

type Result struct {
	Timestamp time.Time `json:"timestamp"`

	DownloadBytes int64   `json:"download-bytes,omitempty"`
	DownloadTime  int64   `json:"download-time,omitempty"`
	DownloadSpeed float64 `json:"download-speed,omitempty"`

	UploadBytes int64   `json:"upload-bytes,omitempty"`
	UploadTime  int64   `json:"upload-time,omitempty"`
	UploadSpeed float64 `json:"upload-speed,omitempty"`
}

type SpeedtestComponent struct {
	cfg *Config

	health    healthOut
	healthMut sync.Mutex

	latestResult    *Result
	latestResultMut sync.Mutex
}

func New(cfg *Config, s *server.Server) (*SpeedtestComponent, error) {
	speedtest := &SpeedtestComponent{
		cfg: cfg,

		health: healthOut{
			Status: healthOutStatusUnknown,
		},
	}

	group := s.RegisterGroup("/speedtest", "speedtest", "Speedtest monitoring API")

	group.GET("/health", []fizz.OperationOption{
		fizz.Summary("Get the state of the MQTT broker"),
		fizz.Response(fmt.Sprint(http.StatusInternalServerError), "Server Error", server.APIError{}, nil, nil),
	}, tonic.Handler(speedtest.handlerPing, http.StatusOK))

	group.GET("/result", []fizz.OperationOption{
		fizz.Summary("Get the latest speedtest result"),
		fizz.Response(fmt.Sprint(http.StatusInternalServerError), "Server Error", server.APIError{}, nil, nil),
	}, tonic.Handler(speedtest.handlerLatestResult, http.StatusOK))

	return speedtest, nil
}

func (c *SpeedtestComponent) setLatestResult(result *Result) {
	c.healthMut.Lock()
	c.latestResultMut.Lock()
	defer c.healthMut.Unlock()
	defer c.latestResultMut.Unlock()

	c.health.Status = healthOutStatusOK
	c.health.Error = nil
	c.latestResult = result
}

func (c *SpeedtestComponent) setError(err error) {
	c.healthMut.Lock()
	c.latestResultMut.Lock()
	defer c.healthMut.Unlock()
	defer c.latestResultMut.Unlock()

	c.health.Status = healthOutStatusOK
	c.health.Error = err
	c.latestResult = nil
}

func (c *SpeedtestComponent) Run(ctx context.Context) error {
	logrus.WithContext(ctx).Debug("starting the speedtest component")
	logrus.WithContext(ctx).Infof("using the %q network interface as public interface", c.cfg.Interface)

	select {
	case <-ctx.Done():
		return component.ErrInterrupted
	}
}
