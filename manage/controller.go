package manage

import (
	"context"
	"sync"

	"github.com/evan-forbes/gather/db/influx"
	"github.com/evan-forbes/sink"
)

// Data than can be written locally along with influxDB
// just use sink.ErrSinker
type WritableData interface {
	sink.Sinker
	influx.Fluxer
}

type Routine interface {
	Boot(context.Context, *sync.WaitGroup) (chan WritableData, <-chan error)
}

// type Controller struct {
// 	Cancel  context.CancelFunc
// 	Ctx     context.Context
// 	LocalWG sync.WaitGroup
// }

// func NewController(parentCtx context.Context, r Routine, w Writer) {

// }
