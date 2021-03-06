// Package HitBTC is an implementation of the HitBTC API in Golang.
package hitbtc

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const (
	API_BASE = "https://api.hitbtc.com/api/2" // HitBtc API endpoint
)

// New returns an instantiated HitBTC struct
func New(apiKey, apiSecret string) *HitBtc {
	client := NewClient(apiKey, apiSecret)
	return &HitBtc{client}
}

// NewWithCustomHttpClient returns an instantiated HitBTC struct with custom http client
func NewWithCustomHttpClient(apiKey, apiSecret string, httpClient *http.Client) *HitBtc {
	client := NewClientWithCustomHttpConfig(apiKey, apiSecret, httpClient)
	return &HitBtc{client}
}

// NewWithCustomTimeout returns an instantiated HitBTC struct with custom timeout
func NewWithCustomTimeout(apiKey, apiSecret string, timeout time.Duration) *HitBtc {
	client := NewClientWithCustomTimeout(apiKey, apiSecret, timeout)
	return &HitBtc{client}
}

// handleErr gets JSON response from livecoin API en deal with error
func handleErr(r interface{}) error {
	switch v := r.(type) {
	case map[string]interface{}:
		error := r.(map[string]interface{})["error"]
		if error != nil {
			switch v := error.(type) {
			case map[string]interface{}:
				errorMessage := error.(map[string]interface{})["message"]
				return errors.New(errorMessage.(string))
			default:
				return fmt.Errorf("I don't know about type %T!\n", v)
			}
		}
	case []interface{}:
		return nil
	default:
		return fmt.Errorf("I don't know about type %T!\n", v)
	}

	return nil
}

// HitBtc represent a HitBTC client
type HitBtc struct {
	client *client
}

// set enable/disable http request/response dump
func (c *HitBtc) SetDebug(enable bool) {
	c.client.debug = enable
}

// GetCurrencies is used to get all supported currencies at HitBtc along with other meta data.
func (b *HitBtc) GetCurrencies() (currencies []Currency, err error) {
	r, err := b.client.do("GET", "public/currency", nil, false)
	if err != nil {
		return
	}
	var response interface{}
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}
	if err = handleErr(response); err != nil {
		return
	}
	err = json.Unmarshal(r, &currencies)
	return
}

// GetSymbols is used to get the open and available trading markets at HitBtc along with other meta data.
func (b *HitBtc) GetSymbols() (symbols []Symbol, err error) {
	r, err := b.client.do("GET", "public/symbol", nil, false)
	if err != nil {
		return
	}
	var response interface{}
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}
	if err = handleErr(response); err != nil {
		return
	}
	err = json.Unmarshal(r, &symbols)
	return
}

// GetTicker is used to get the current ticker values for a market.
func (b *HitBtc) GetTicker(market string) (ticker Ticker, err error) {
	r, err := b.client.do("GET", "public/ticker/"+strings.ToUpper(market), nil, false)
	if err != nil {
		return
	}
	var response interface{}
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}
	if err = handleErr(response); err != nil {
		return
	}
	err = json.Unmarshal(r, &ticker)
	return
}

// Market
// GetOrderBook is used to get the current market order book values for a market.
func (b *HitBtc) GetOrderBook(market string, depth int) (orderBook OrderBook, err error) {
	r, err := b.client.do("GET", "public/orderbook/"+strings.ToUpper(market), nil, false)
	if err != nil {
		return
	}
	var response interface{}
	println("Response: ", string(r))
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}
	if err = handleErr(response); err != nil {
		return
	}
	err = json.Unmarshal(r, &orderBook)
	orderBook.Asks = orderBook.Asks[:Min(len(orderBook.Asks), depth)]
	orderBook.Bids = orderBook.Bids[:Min(len(orderBook.Bids), depth)]
	return
}

// Account

// GetBalances is used to retrieve all balances from your account
func (b *HitBtc) GetBalances() (balances []Balance, err error) {
	r, err := b.client.do("GET", "trading/balance", nil, true)
	if err != nil {
		return
	}
	var response interface{}
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}
	if err = handleErr(response); err != nil {
		return
	}
	err = json.Unmarshal(r, &balances)
	return
}

// Getbalance is used to retrieve the balance from your account for a specific currency.
// currency: a string literal for the currency (ex: LTC)
func (b *HitBtc) GetBalance(currency string) (balance Balance, err error) {
	r, err := b.client.do("GET", "payment/balance", nil, true)
	if err != nil {
		return
	}
	var response interface{}
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}
	if err = handleErr(response); err != nil {
		return
	}
	var balances []Balance
	currency = strings.ToUpper(currency)
	err = json.Unmarshal(r, &balances)
	if err != nil {
		return
	}
	for _, balance = range balances {
		if balance.Currency == currency {
			return
		}
	}
	return Balance{}, errors.New("Currency not found")
}

// GetTrades used to retrieve your trade history.
// market string literal for the market (ie. BTC/LTC). If set to "all", will return for all market
func (b *HitBtc) GetTrades(currencyPair string) (trades []Trade, err error) {
	payload := make(map[string]string)
	if currencyPair != "all" {
		payload["symbol"] = currencyPair
	}
	r, err := b.client.do("GET", "history/trades", payload, true)
	if err != nil {
		return
	}
	var response interface{}
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}
	if err = handleErr(response); err != nil {
		return
	}
	err = json.Unmarshal(r, &trades)
	return
}

// GetTransactions is used to retrieve your withdrawal and deposit history
// "Start" and "end" are given in UNIX timestamp format in miliseconds and used to specify the date range for the data returned.
func (b *HitBtc) GetTransactions(start uint64, end uint64, limit uint32) (transactions []Transaction, err error) {
	payload := make(map[string]string)
	if start > 0 {
		payload["from"] = strconv.FormatUint(uint64(start), 10)
	}
	if end == 0 {
		end = uint64(time.Now().Unix()) * 1000
	}
	if end > 0 {
		payload["till"] = strconv.FormatUint(uint64(end), 10)
	}
	if limit > 1000 {
		limit = 1000
	}
	if limit > 0 {
		payload["limit"] = strconv.FormatUint(uint64(limit), 10)
	}
	r, err := b.client.do("GET", "account/transactions", payload, true)
	if err != nil {
		return
	}
	var response interface{}
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}
	if err = handleErr(response); err != nil {
		return
	}
	err = json.Unmarshal(r, &transactions)
	return
}

// CreateOrder is used to retrieve all balances from your account
func (b *HitBtc) CreateOrder(pair string, amount float64, price float64, side string, orderType string) (order Order, err error) {
	payload := make(map[string]string)
	payload["symbol"] = pair
	payload["side"] = side
	payload["quantity"] = strconv.FormatFloat(amount, 'f', 6, 64)
	if orderType != "market" {
		payload["price"] = strconv.FormatFloat(price, 'f', 6, 64)
	}
	payload["type"] = orderType

	payload["timeInForce"] = "GTC"

	r, err := b.client.do("POST", "order", payload, true)
	if err != nil {
		return
	}
	var response interface{}
	if err = json.Unmarshal(r, &response); err != nil {
		return
	}
	if err = handleErr(response); err != nil {
		return
	}
	err = json.Unmarshal(r, &order)
	return order, nil
}

func Min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
