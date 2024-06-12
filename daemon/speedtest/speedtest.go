package speedtest

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/loopfz/gadgeto/tonic"
	"github.com/shirou/gopsutil/v4/net"
	"github.com/showwin/speedtest-go/speedtest"
	"github.com/sirupsen/logrus"
	"github.com/wI2L/fizz"

	"github.com/rclsilver-org/monitoring/daemon/pkg/component"
	"github.com/rclsilver-org/monitoring/daemon/pkg/server"
)

type Result struct {
	Timestamp time.Time `json:"timestamp"`

	Ping time.Duration `json:"ping,omitempty"`

	Download       speedtest.ByteRate `json:"download,omitempty"`
	DownloadString string             `json:"download-string,omitempty"`

	Upload       speedtest.ByteRate `json:"upload,omitempty"`
	UploadString string             `json:"upload-string,omitempty"`
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

	if _, _, err := c.fetchInterfaceInfo(ctx); err != nil {
		return err
	}
	logrus.WithContext(ctx).Debugf("using the %q network interface as public interface", c.cfg.Interface)

	logrus.WithContext(ctx).Debugf("executing a test every %s (retry interval is %s)", c.cfg.Interval, c.cfg.RetryInterval)

	for {
		interval := c.cfg.Interval

		result, err := c.execute(ctx)
		if err != nil {
			c.setError(err)
			interval = c.cfg.RetryInterval
			logrus.WithContext(ctx).WithError(err).Warningf("test has failed, retrying in %s", interval)
		} else {
			c.setLatestResult(result)
			logrus.WithContext(ctx).Infof("test finished: Ping: %s - DL: %s - UL: %s", result.Ping, result.Download.String(), result.Upload.String())
			logrus.WithContext(ctx).Debugf("next test scheduled in %s", interval)
		}

		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return component.ErrInterrupted
		}
	}
}

func (c *SpeedtestComponent) execute(ctx context.Context) (*Result, error) {
	client := speedtest.New()

	if _, _, err := c.fetchInterfaceInfo(ctx); err != nil {
		return nil, err
	}

	logrus.WithContext(ctx).Debug("fetching the user info")
	user, err := client.FetchUserInfoContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch user info: %w", err)
	}
	logrus.WithContext(ctx).Debugf("found user info: %s (%s) [%s, %s]", user.IP, user.Isp, user.Lat, user.Lon)

	logrus.WithContext(ctx).Debug("fetching servers list")
	servers, err := client.FetchServerListContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to fetch servers list: %w", err)
	}

	logrus.WithContext(ctx).Debug("searching better server")
	targets, err := servers.FindServer([]int{})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch targets: %w", err)
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no server available to execute a test")
	}

	logrus.WithContext(ctx).Debug("executing the ping test")
	if err := targets[0].PingTestContext(ctx, nil); err != nil {
		return nil, fmt.Errorf("error while executing the ping test: %w", err)
	}

	logrus.WithContext(ctx).Debug("executing the download test")
	download, err := c.executeDownload(ctx, func() error {
		return targets[0].DownloadTestContext(ctx)
	})
	if err != nil {
		return nil, fmt.Errorf("error while executing the download test: %w", err)
	}

	logrus.WithContext(ctx).Debug("executing the upload test")
	upload, err := c.executeUpload(ctx, func() error {
		return targets[0].UploadTestContext(ctx)
	})
	if err != nil {
		return nil, fmt.Errorf("error while executing the upload test: %w", err)
	}

	return &Result{
		Timestamp: time.Now(),
		Ping:      targets[0].Latency,

		Download:       download,
		DownloadString: download.String(),

		Upload:       upload,
		UploadString: upload.String(),
	}, nil
}

// fetchInterfaceInfo returns the byte sent and received for the configured interface
func (c *SpeedtestComponent) fetchInterfaceInfo(ctx context.Context) (uint64, uint64, error) {
	stats, err := net.IOCountersWithContext(ctx, true)
	if err != nil {
		return 0, 0, fmt.Errorf("unable to get the I/O counters: %w", err)
	}

	for _, i := range stats {
		if i.Name == c.cfg.Interface {
			return i.BytesSent, i.BytesRecv, nil
		}
	}

	return 0, 0, fmt.Errorf("interface %q not found", c.cfg.Interface)
}

// executeDownload fetch the interface data to compute the real byte rate
func (c *SpeedtestComponent) executeDownload(ctx context.Context, f func() error) (speedtest.ByteRate, error) {
	_, rxBefore, err := c.fetchInterfaceInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("unable to get network interface info: %w", err)
	}

	timeBefore := time.Now()

	if err := f(); err != nil {
		return 0, err
	}

	timeAfter := time.Now()

	_, rxAfter, err := c.fetchInterfaceInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("unable to get network interface info: %w", err)
	}

	return speedtest.ByteRate(float64(rxAfter-rxBefore) / (timeAfter.Sub(timeBefore).Seconds())), nil
}

// executeUpload fetch the interface data to compute the real byte rate
func (c *SpeedtestComponent) executeUpload(ctx context.Context, f func() error) (speedtest.ByteRate, error) {
	txBefore, _, err := c.fetchInterfaceInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("unable to get network interface info: %w", err)
	}

	timeBefore := time.Now()

	if err := f(); err != nil {
		return 0, err
	}

	timeAfter := time.Now()

	txAfter, _, err := c.fetchInterfaceInfo(ctx)
	if err != nil {
		return 0, fmt.Errorf("unable to get network interface info: %w", err)
	}

	return speedtest.ByteRate(float64(txAfter-txBefore) / (timeAfter.Sub(timeBefore).Seconds())), nil
}

type interfaceInfo struct {
	DownloadedBytes int64
	UploadedBytes   int64
}

type contextRoundTripper struct {
	rt      http.RoundTripper
	ctx     context.Context
	timeout time.Duration
}

// RoundTrip executes a single HTTP transaction and injects the context into the request
func (crt *contextRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx, cancel := context.WithTimeout(crt.ctx, crt.timeout)
	defer cancel()

	req = req.WithContext(ctx)

	return crt.rt.RoundTrip(req)
}
