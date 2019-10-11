package coinbase

import (
	"strings"
	"time"

	"github.com/evan-forbes/gather/standard"
)

////////////////////////////////////
// 			Sending
//////////////////////////////////

type Channel struct {
	Name       string   `json:"name"`
	ProductIds []string `json:"product_ids"`
}

type Subscribe struct {
	Type       string   `json:"type"`
	ProductIds []string `json:"product_ids"`
	Channels   []string `json:"channels"`
}

func NewSub(tickers []string) Subscribe {
	return Subscribe{
		Type:       "subscribe",
		ProductIds: tickers,
		Channels:   []string{"ticker"},
	}
}

////////////////////////////////////
// 			Recieving
//////////////////////////////////

type SubStatus struct {
	Type     string    `json:"type"`
	Channels []Channel `json:"channels"`
}

func (s *SubStatus) Label() string {
	return "Coinbase"
}

type Ticker struct {
	Type      string    `json:"type"`
	TradeID   int       `json:"trade_id"`
	Sequence  int64     `json:"sequence"`
	Time      time.Time `json:"time"`
	ProductID string    `json:"product_id"`
	Price     string    `json:"price"`
	Side      string    `json:"side"`
	LastSize  string    `json:"last_size"`
	BestBid   string    `json:"best_bid"`
	BestAsk   string    `json:"best_ask"`
}

func (t *Ticker) Label() string {
	return strings.Replace(t.ProductID, "-", "_", 1) + "_coinbase"
}

func (t *Ticker) Standardize() *standard.PriceUpdate {
	return &standard.PriceUpdate{
		Pair:    strings.Replace(t.ProductID, "-", "_", 1),
		Source:  "CoinbaseWS",
		Price:   t.Price,
		Time:    t.Time,
		BestAsk: t.BestAsk,
		BestBid: t.BestBid,
	}
}

type Heartbeat struct {
	Type        string    `json:"type"`
	Sequence    int       `json:"sequence"`
	LastTradeID int       `json:"last_trade_id"`
	ProductID   string    `json:"product_id"`
	Time        time.Time `json:"time"`
}

type UnknownMessage struct {
	Message string
}

func (ufo UnknownMessage) Label() string {
	return "UnknownMessageKraken"
}

type UnknownError struct {
	Err     error
	Message string
}

func (ufo UnknownError) Label() string {
	return "UnknownErrorKraken"
}
