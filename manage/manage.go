package manage

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/evan-forbes/sink"
	"github.com/pkg/errors"
)

// Controller manages a rebootable system
type Controller struct {
	Boot       Booter
	RootCtx    context.Context
	Cancel     context.CancelFunc
	WG         sync.WaitGroup
	StartTime  time.Time
	WriteLocal bool
}

func NewController(parentCtx context.Context, boot Booter) *Controller {
	return &Controller{
		Boot:      boot,
		RootCtx:   parentCtx,
		StartTime: time.Now(),
	}
}

// Booter describes a function that boots up a continous process
// that is monitored and  by the Controller
type Booter func(context.Context, *sync.WaitGroup) (<-chan error, error)

// defo going to implement this soon
// type Writer func(context.Context, *sync.WaitGroup, chan *client.Point) (<-chan error, error)
// need to have booters return two channels and writers return one
// the initial error returned by booters needs to go in the errc

// probably need a pretty hefty reorg for the managing and the exchange
// portions of this application. resetting could be better looped.

// Start begins processing described by the Booter
func (s *Controller) Start() error {
	childCtx, childCancel := context.WithCancel(s.RootCtx)
	s.Cancel = childCancel
	errc, err := s.Boot(childCtx, &s.WG)
	if err != nil {
		return err
	}
	s.HandleErrc(errc)
	return nil
}

func (s *Controller) HandleErrc(errc <-chan error) {
	go func() {
		for err := range errc {
			if val, ok := err.(ErrorSignal); ok {
				switch val.Action {
				case REBOOT:
					s.Retry(5, 5*time.Minute)
				case SHUTDOWN:
					s.Cancel()
				}
			}
			Logger <- sink.Wrap("Errors", &err)
		}
	}()
}

// Retry will attempt to reboot the system and then sleeps exponetially
// longer per attempt
func (s *Controller) Retry(attempts int, sleep time.Duration) error {
	err := s.Reboot()
	if err != nil {
		if attempts--; attempts > 0 {
			// Add some randomness
			jitter := time.Duration(rand.Int63n(int64(sleep)))
			sleep = sleep + jitter/2

			time.Sleep(sleep)
			return s.Retry(attempts, 3*sleep)
		}
		s.Cancel()
		return SignalWrap(err, SHUTDOWN)
	}
	return nil
}

// Reboot employs the Controller interface to reboot
func (s *Controller) Reboot() error {
	s.Cancel()
	s.WG.Wait()
	err := s.Start()
	if err != nil {
		return errors.Wrap(err, "Problem during Startup")
	}
	return nil
}

// WaitForInterupt spins up a goroutine to watch for the interupt
// signal produced by pressing ctrl+C to call the provided
// cancel function and alternatively watching for the cancel
// function to be called to shutdown the program
func WaitForInterupt(ctx context.Context, cancel context.CancelFunc, wg *sync.WaitGroup) {
	interupt := make(chan os.Signal, 1)
	signal.Notify(interupt, os.Interrupt, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-interupt:
				fmt.Println("sending cancels")
				cancel()
			case <-ctx.Done():
				wg.Wait()
				close(Logger)
				os.Exit(0)
			}
		}
	}()
}

// Logger, the global json logger
var Logger chan<- sink.Sinker

func StartLogging(ctx context.Context, wg *sync.WaitGroup) {
	s := sink.Start(ctx, wg, "./logs")
	Logger = s.Input
}

// in essence, what is the controller supposed to do?
// reboot a system, should the system stop responding or if the system
// reccomends to do so.

// connect the system to other systems, and reboot those systems
// should they need rebooting

// log all errors thrown by the systems
