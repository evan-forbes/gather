package kraken

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/evan-forbes/gather/centralized/ws"
	"github.com/evan-forbes/gather/manage"
	"github.com/evan-forbes/trend-tracker/process/parse"
	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/pkg/errors"
)

const (
	WSURL = "wss://ws.kraken.com"
)

func Boot(ctx context.Context, wg *sync.WaitGroup) (<-chan error, error) {
	////// Setup //////
	// errc channel for all errors in the this 'session'
	errc := make(chan error, 5)
	// channel to write to the socket
	writeInput := make(chan interface{}, 1)
	// Connect to Websocket
	socket, err := ws.NewSocket(WSURL, errc)
	if err != nil {
		return nil, errors.Wrap(err, "Could not boot coinbase")
	}
	go func() {
		defer close(errc)
		defer close(writeInput)
		// recieving channel of raw data
		payloads := socket.Listen(ctx, writeInput)
		processed := sortPayload(payloads, errc)
		// Send first message
		firstMsg := NewSub([]string{"XTZ/USD", "ATOM/USD"})
		writeInput <- firstMsg
		// Process the output of
		for point := range processed {
			fmt.Println(point)
		}
	}()
	return errc, nil
}

// sortPayload  unmarshals the incoming payloads from coinbase
// into the prespecified types in coinbase/types.go
func sortPayload(payloads <-chan []byte, errc chan<- error) <-chan *client.Point {
	output := make(chan *client.Point, 10)
	go func() {
		defer close(output)
		for payload := range payloads {
			jsonPayload := string(payload)
			switch {
			case ContainsAll(jsonPayload, "[", "a", "b", "c", "v", "p", "t"):
				out, err := ConvertToTicker(jsonPayload)
				if err != nil {
					fmt.Println("Could not sort kraken ticker payload", err)
					errc <- err
				}
				fluxed, err := out.Standardize().Influx()
				if err != nil {
					errc <- err
				}
				output <- fluxed
			case parse.ContainsAll(jsonPayload, "heartbeat"):
			case ContainsAll(jsonPayload, "channelID", "subscription"):
				var out SubStatus
				err := json.Unmarshal(payload, &out)
				if err != nil {
					fmt.Println("Could not Unmarshal while sorting kraken payload", err)
					errc <- err
				}
				manage.Logger <- &out
			case ContainsAll(jsonPayload, "connectionID", "status"):
				var out ConnectionStatus
				err := json.Unmarshal(payload, &out)
				if err != nil {
					errc <- errors.Wrapf(err, "Processing error %s %s", "kraken", time.Now())
				}

			default:
				fmt.Println("Unidentified Payload:\n", jsonPayload)
			}
		}
	}()
	return output
}

// func sortPayload(payload []byte) (sink.Sinker, error) {
// 	jsonPayload := string(payload)
// 	switch {
// 	// check if the payload is a ticker
// 	case parse.ContainsAll(jsonPayload, "[", "a", "b", "c", "v", "p", "t"):
// 		out, err := ConvertToTicker(jsonPayload)
// 		if err != nil {
// 			fmt.Println("Could not sort kraken ticker payload", err)
// 			return nil, err
// 		}
// 		return out, nil

// 	case parse.ContainsAll(jsonPayload, "heartbeat"):
// 		fmt.Println("---")

// 	case parse.ContainsAll(jsonPayload, "connectionID"):
// 		var out ConnectionStatus
// 		err := json.Unmarshal(payload, &out)
// 		if err != nil {
// 			fmt.Println("Could not Unmarshal while sorting kraken payload", err)
// 			return nil, err
// 		}
// 		return out, nil

// 	case parse.ContainsAll(jsonPayload, "channelID", "subscription"):
// 		var out SubStatus
// 		err := json.Unmarshal(payload, &out)
// 		if err != nil {
// 			fmt.Println("Could not Unmarshal while sorting kraken payload", err)
// 			return out, err
// 		}
// 		out.AddChannel()
// 		return out, nil
// 	default:
// 		fmt.Println("Unidentified Payload:\n", jsonPayload)
// 		return sink.Unregistered{Item: jsonPayload}, nil
// 	}
// 	return nil, nil
// }

// ContainsAll only return true in all chars are found
// at least once in main string
func ContainsAll(main string, chars ...string) bool {
	for _, char := range chars {
		if !strings.Contains(main, char) {
			return false
		}
	}
	return true
}
