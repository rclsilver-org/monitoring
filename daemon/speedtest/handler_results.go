package speedtest

import (
	"github.com/gin-gonic/gin"
)

type latestResultOut struct {
	Result *Result `json:"result"`
}

func (speedtest *SpeedtestComponent) handlerLatestResult(c *gin.Context) (*latestResultOut, error) {
	speedtest.latestResultMut.Lock()
	defer speedtest.latestResultMut.Unlock()

	return &latestResultOut{
		Result: speedtest.latestResult,
	}, nil
}
