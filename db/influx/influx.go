package influx

import (
	"os"
	"sync"
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/pkg/errors"
)

// ---Measurements---
// Orders
// - bid or ask bool or string tagged
// - price float64
// - amount // try to go numeric but unfortunately, string might be the best option
// - owner string

// Trades

// Stats
// - spread
// - amount mkr/dai/weth locked in contract / orders
// -

// todo
// make sure to handle the logs
// make a fluxer method for all of the logs that I deem important

var (
	LocalHTTPConfig = client.HTTPConfig{
		Addr:      "http://localhost:8086",
		UserAgent: os.Getenv("username"),
		Password:  os.Getenv("darthPass"),
	}
	TradingConfig = client.BatchPointsConfig{
		Precision:       "u",
		Database:        "trade",
		RetentionPolicy: "autogen",
	}
)

type BatchManager struct {
	DefaultConf  client.BatchPointsConfig
	Client       client.Client
	ErrorC       chan error
	CountTrigger int
	TimeTrigger  int64
	WG           *sync.WaitGroup
}

func NewBatchManager(cConf client.HTTPConfig, bconf client.BatchPointsConfig) (*BatchManager, error) {
	clnt, err := client.NewHTTPClient(cConf)
	if err != nil {
		return nil, errors.Wrapf(err, "Connection error to %s", cConf.Addr)
	}
	errc := make(chan error, 1)
	return &BatchManager{
		DefaultConf:  bconf,
		Client:       clnt,
		ErrorC:       errc,
		CountTrigger: 100,
		TimeTrigger:  40,
	}, nil
}

// Handler will auto batch and write points based on the number of points
// or the time.
func (bm *BatchManager) Handler(points <-chan *client.Point) {
	bm.WG.Add(1)
	go func() {
		cache, err := client.NewBatchPoints(bm.DefaultConf)
		if err != nil {
			bm.ErrorC <- errors.Wrap(err, "Writing")
		}
		count := 0
		timer := time.Now().Unix()
		for point := range points {
			cache.AddPoint(point)
			curr := time.Now().Unix()
			if (curr-timer) >= bm.TimeTrigger || count >= bm.CountTrigger {
				err := bm.Client.Write(cache)
				if err != nil {
					bm.ErrorC <- errors.Wrap(err, "Writing")
				}
				cache, _ = client.NewBatchPoints(bm.DefaultConf)
			}
		}
		bm.Client.Write(cache)
		bm.WG.Done()
	}()
}

type Fluxer interface {
	Influx() (*client.Point, error)
}
