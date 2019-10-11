package kraken

import (
	json "encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/evan-forbes/gather/standard"
)

///////////////////////////////////
//  	Sending
//////////////////////////////

var tickerSub = struct {
	Name string `json:"name"`
}{"ticker"}

type Sub struct {
	Event        string   `json:"event"`
	Pair         []string `json:"pair"`
	Subscription struct {
		Name string `json:"name"`
	} `json:"subscription"`
}

func NewSub(pair []string) Sub {
	message := Sub{
		Event:        "subscribe",
		Pair:         pair,
		Subscription: tickerSub,
	}
	return message
}

///////////////////////////////////
//  	Recieving
//////////////////////////////

type ConnectionStatus struct {
	ConnectionID *big.Int `json:"connectionID"`
	Event        string   `json:"event"`
	Status       string   `json:"status"`
	Version      string   `json:"version"`
}

// Label is a method to abide by the
// sink interface for sorting data
func (c *ConnectionStatus) Label() string {
	return "ConnectionStatusKraken"
}

// Channels maps the int that represent kraken
// ticker channels to ticker name
var Channels = make(map[int]string)

type SubStatus struct {
	ChannelID    int    `json:"channelID"`
	Event        string `json:"event"`
	Status       string `json:"status"`
	Pair         string `json:"pair"`
	Subscription struct {
		Name string `json:"name"`
	} `json:"subscription"`
}

// AddChannel
func (sub *SubStatus) AddChannel() {
	if _, contains := Channels[sub.ChannelID]; !contains {
		Channels[sub.ChannelID] = strings.Replace(sub.Pair, "/", "_", 1)
	}
}

// Label is a method to abide by the
// sink interface for sorting data
func (c *SubStatus) Label() string {
	return "SubsciptionStatusKraken"
}

//////// Tickers //////////

type Ticker struct {
	Channel  int           `json:",omitempty"`
	Pair     string        `json:",omitempty"`
	Ask      []interface{} `json:"a"`
	Bid      []interface{} `json:"b"`
	Close    []string      `json:"c"`
	Volume   []string      `json:"v"`
	VolPrice []string      `json:"p"`
	Trades   []int         `json:"t"`
	Low      []string      `json:"l"`
	High     []string      `json:"h"`
	Open     []string      `json:"o"`
	Time     time.Time     `json:,omitempty"`
}

func (rt Ticker) Label() string {
	return rt.Pair + "_kraken"
}

func parseOffer(data []interface{}) string {
	out, ok := data[0].(string)
	if ok {
		return out
	}
	return ""
}

// right now I'm using tickers, but I'm going to have to monitor
// the entire orderbook eventually/now to make any useuful bots.
func (t *Ticker) Standardize() *standard.PriceUpdate {
	ba, _ := t.Ask[0].(string)
	bb, _ := t.Bid[0].(string)
	return &standard.PriceUpdate{
		Pair:    t.Pair,
		Source:  "Kraken",
		Price:   t.Close[0],
		BestAsk: ba,
		BestBid: bb,
		Time:    t.Time,
	}
}

// parsing a ticker
// split by "{" and "}"

// ConvertToTicker is an fugly function to parse the json array of mixed
// objects that kraken uses to hold ticker information
func ConvertToTicker(payload string) (Ticker, error) {
	var out Ticker
	load := strings.SplitN(payload, ",", 2)
	rawData := strings.Split(load[1], "ticker")
	channel, err := strconv.Atoi(load[0][1:])
	if err != nil {
		fmt.Println("Could not convert string to int while converting ticker: ", err)
		return out, err
	}
	fmt.Println(rawData[0][:len(rawData[0])-2])
	err = json.Unmarshal([]byte(rawData[0][:len(rawData[0])-2]), &out)
	if err != nil {
		fmt.Println("Could not unmarshal ticker while converting: ", err)
		return out, err
	}
	out.Channel = channel
	out.Pair = Channels[channel]
	out.Time = time.Now()
	return out, nil
}

///////	Orderbook ///////

type BookSnapshot struct {
	Bids [][]string `json:"bs"`
	Asks [][]string `json:"as"`
}

// func ConvertToSnapshot(payload string) *BookSnapshot {

// }
