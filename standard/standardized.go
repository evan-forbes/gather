package standard

import (
	"time"

	client "github.com/influxdata/influxdb1-client/v2"
	"github.com/pkg/errors"
)

/////////////////////////////////
// 		   PriceUpdate
//////////////////////////////

type PriceUpdater interface {
	Standardize() *PriceUpdate
}

// PriceUpdate is a standardized price update for a given market
type PriceUpdate struct {
	// Pair is the assets that define the price
	// format: ABC_DEF
	Pair string
	// Source is some string that represents where the
	// update came from
	Source string
	// Price might need to use something with better precision
	Price, BestBid, BestAsk string
	// Time should be the unix epoch as close as possible
	// to the time recieved or time included with the order
	Time time.Time
}

func (p *PriceUpdate) Influx() (*client.Point, error) {
	tags := map[string]string{
		"Pair":   p.Pair,
		"Source": p.Source,
	}
	fields := map[string]interface{}{
		"Price":   p.Price,
		"BestBid": p.BestBid,
		"BestAsk": p.BestAsk,
	}

	return client.NewPoint("PriceUpdates", tags, fields, p.Time)
}

/////////////////////////////////
// 		    Offers
//////////////////////////////

// OfferAction describes a category of OfferEvent
type OfferAction int

const (
	TAKE OfferAction = 1
	MAKE OfferAction = 2
	BUMP OfferAction = 3
	KILL OfferAction = 4
)

// String fulfills the Stringer interface for
// OfferAction
func (o OfferAction) String() string {
	names := [...]string{
		"na",
		"TAKE", "MAKE", "BUMP", "KILL",
	}
	if o > KILL || o < TAKE {
		return "Unregistered OfferAction"
	}

	return names[o]
}

type OfferEventer interface {
	Standardize() *OfferEvent
}

// OfferEvent is the standardized object to represent an offer
// it uses strings to keep precision
type OfferEvent struct {
	Id              string
	Selling, Buying string
	Amount          string
	Price           string
	Timestamp       time.Time
	Source          string
	Event           OfferAction
	Actor           string
	SecondaryActor  string
}

// Influx fulfills the Fluxer interface by determining
// what is tagged (indexed) by influx db
func (oe *OfferEvent) Influx() (*client.Point, error) {
	tags := map[string]string{
		"Source":  oe.Source,
		"Buying":  oe.Buying,
		"Selling": oe.Selling,
		"Event":   oe.Event.String(),
	}
	fields := map[string]interface{}{
		"Id":     oe.Id,
		"Amount": oe.Amount,
		"Price":  oe.Price,
		"Actor":  oe.Actor,
	}
	if oe.Event == TAKE {
		fields["SecondaryActor"] = oe.SecondaryActor
	}
	pt, err := client.NewPoint("Offers", tags, fields, oe.Timestamp)
	if err != nil {
		return nil, errors.Wrapf(err, "Influxing %s", oe.Source)
	}
	return pt, nil
}

////////////////////////////////
// idk if I'm goin to use the below quite yet

// Order is the Standardized version of an order
type Order struct {
	Pair         string
	Price        float64
	Action       interface{}
	Amount       float64
	TimeCreated  int64
	TimeExecuted int64
	Filled       bool
}

// Shotgun will split a single order into smaller spread out
// Orders in the dir (1 or -1) into count number of orders,
// and incremented by percentage of the the original price
// by increment % (max .2)
func (o Order) Shotgun(dir, count int, increment float64) []Order {
	var out []Order
	orderSize := o.Amount / float64(count)
	stepSize := o.Price * increment
	// set a max increment of 20% of the price
	if increment > .2 {
		stepSize = o.Price * .2
	}
	// make copies of the order, with apro changes to the amounts and price
	for i := 1; i <= count; i++ {
		newOrder := o
		newOrder.Price = o.Price + (stepSize * float64(dir) * float64(i))
		newOrder.Amount = orderSize
		out = append(out, newOrder)
	}
	return out
}
