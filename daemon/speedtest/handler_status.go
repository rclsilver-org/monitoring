package speedtest

import (
	"github.com/gin-gonic/gin"
)

type healthOutStatus string

const (
	healthOutStatusOK      healthOutStatus = "OK"
	healthOutStatusUnknown healthOutStatus = "UNKNOWN"
)

type healthOut struct {
	Status healthOutStatus `json:"status"`
	Error  error           `json:"error"`
}

func (speedtest *SpeedtestComponent) handlerPing(c *gin.Context) (*healthOut, error) {
	speedtest.healthMut.Lock()
	defer speedtest.healthMut.Unlock()

	return &healthOut{
		Status: speedtest.health.Status,
		Error:  speedtest.health.Error,
	}, nil
}
