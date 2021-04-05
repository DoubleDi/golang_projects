package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/shopspring/decimal"
)

type OrderType int

const (
	OrderTypeAsk OrderType = iota + 1
	OrderTypeBid
)

type Order struct {
	t      OrderType
	amount decimal.Decimal
	price  decimal.Decimal
}

type ExecutionResult struct {
	amount string
	price  string
}

type OrderStorage struct {
	asks []*Order // sorted from high to low
	bids []*Order // sorted from low to high
}

func NewOrderStorage() *OrderStorage {
	return &OrderStorage{}
}

func (s *OrderStorage) Execute(o *Order) []ExecutionResult {
	if o.t == OrderTypeAsk {
		return s.executeAsk(o)
	} else if o.t == OrderTypeBid {
		return s.executeBid(o)
	}
	return []ExecutionResult{}
}

func (s *OrderStorage) executeBid(o *Order) []ExecutionResult {
	result := []ExecutionResult{}
	for o.amount.Cmp(decimal.Zero) > 0 {
		if len(s.asks) == 0 {
			s.insertBid(o)
			return result
		}
		leastAsk := s.asks[len(s.asks)-1]
		if o.price.Cmp(leastAsk.price) >= 0 {
			amount := leastAsk.amount
			if o.amount.Cmp(leastAsk.amount) >= 0 {
				result = append(result, ExecutionResult{
					price:  leastAsk.price.String(),
					amount: leastAsk.amount.String(),
				})
				s.asks = s.asks[:len(s.asks)-1]
			} else {
				result = append(result, ExecutionResult{
					price:  leastAsk.price.String(),
					amount: o.amount.String(),
				})
				leastAsk.amount = leastAsk.amount.Sub(o.amount)
			}
			o.amount = o.amount.Sub(amount)
		} else {
			s.insertBid(o)
			return result
		}
	}
	return result
}

func (s *OrderStorage) executeAsk(o *Order) []ExecutionResult {
	result := []ExecutionResult{}
	for o.amount.Cmp(decimal.Zero) > 0 {
		if len(s.bids) == 0 {
			s.insertAsk(o)
			return result
		}
		mostBid := s.bids[len(s.bids)-1]
		if o.price.Cmp(mostBid.price) <= 0 {
			amount := mostBid.amount
			if o.amount.Cmp(mostBid.amount) >= 0 {
				result = append(result, ExecutionResult{
					price:  mostBid.price.String(),
					amount: mostBid.amount.String(),
				})
				s.bids = s.bids[:len(s.bids)-1]
			} else {
				result = append(result, ExecutionResult{
					price:  mostBid.price.String(),
					amount: o.amount.String(),
				})
				mostBid.amount = mostBid.amount.Sub(o.amount)
			}
			o.amount = o.amount.Sub(amount)
		} else {
			s.insertAsk(o)
			return result
		}
	}
	return result
}

func (s *OrderStorage) insertBid(o *Order) {
	s.bids = append(s.bids, o)
	sort.Slice(s.bids, func(i, j int) bool { return s.bids[i].price.Cmp(s.bids[j].price) < 0 })
}

func (s *OrderStorage) insertAsk(o *Order) {
	s.asks = append(s.asks, o)
	sort.Slice(s.asks, func(i, j int) bool { return s.asks[i].price.Cmp(s.asks[j].price) > 0 })
}

func main() {
	data := `buy, 1.00000000, 100.00, 2018-08-06T19:26:26+00:00
	sell, 1.32350000, 104.00, 2018-08-06T19:27:31+00:00
	sell, 4.20000000, 103.50, 2018-08-06T19:30:34+00:00
	buy, 2.25000000, 101.00, 2018-08-06T19:33:37+00:00
	sell, 5.00000000, 100.75, 2018-08-06T19:35:12+00:00
	buy, 3.45000000, 100.55, 2018-08-06T19:40:07+00:00
	sell, 2.73400000, 100.45, 2018-08-06T19:45:55+00:00
	sell, 2.20000000, 103.50, 2018-08-06T19:48:55+00:00
	buy, 0.50000000, 100.75, 2018-08-06T19:48:55+00:00`

	s := NewOrderStorage()

	rows := strings.Split(data, "\n")
	fmt.Printf("%#v\n", rows)
	for _, row := range rows {
		elems := strings.Split(row, ", ")
		o := &Order{}
		t := strings.TrimSpace(elems[0])
		if t == "buy" {
			o.t = OrderTypeBid
		} else if t == "sell" {
			o.t = OrderTypeAsk
		} else {
			continue
		}
		amount, err := decimal.NewFromString(strings.TrimSpace(elems[1]))
		if err != nil {
			continue
		}
		o.amount = amount
		price, err := decimal.NewFromString(strings.TrimSpace(elems[2]))
		if err != nil {
			continue
		}
		o.price = price

		fmt.Printf("Executing %v\n", elems)
		r := s.Execute(o)
		fmt.Printf("Result %#v\n\n", r)
	}
}
