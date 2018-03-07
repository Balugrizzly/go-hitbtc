package hitbtc

import (
	"encoding/json"
)

type OrderBookEntry struct {
	Price         float64   `json:"price,string"`
	Amount         float64   `json:"size,string"`
}

type OrderBook struct {
	Asks         []OrderBookEntry   `json:"ask,string"`
	Bids         []OrderBookEntry   `json:"bid,string"`
}

func (t *OrderBook) UnmarshalJSON(data []byte) error {
	var err error
	type Alias OrderBook
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(t),
	}
	if err = json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}
