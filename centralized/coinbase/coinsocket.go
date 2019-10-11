package coinbase

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	client "github.com/influxdata/influxdb1-client/v2"

	"github.com/evan-forbes/gather/centralized/ws"
	"github.com/evan-forbes/gather/manage"
	"github.com/pkg/errors"
)

const (
	WSURL = "wss://ws-feed.pro.coinbase.com"
)

// Boot is the booter function for connecting to the coinbase websocket api
// start a connection, and websocket streaming subscriptions for several assets
//
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
		firstMsg := NewSub([]string{"ETH-USD", "BTC-USD", "DAI-USDC", "ZRX-USD", "ETH-DAI"})
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
			case ContainsAll(jsonPayload, "ticker", "best_bid", "price"):
				var out Ticker
				err := json.Unmarshal(payload, &out)
				if err != nil {
					errc <- errors.Wrap(err, "Coinbase parsing 'Ticker' issue:")
				}
				fluxed, err := out.Standardize().Influx()
				if err != nil {
					errc <- err
				}
				output <- fluxed
			case ContainsAll(jsonPayload, "subscriptions", "product_ids"):
				var out SubStatus
				err := json.Unmarshal(payload, &out)
				if err != nil {
					errc <- errors.Wrap(err, "Coinbase parsing 'SubStatus' issue")
				}
				fmt.Println(out)
				manage.Logger <- &out
			default:
				fmt.Println("Unidentified Payload:\n", jsonPayload)
			}
		}
	}()
	return output
}

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
