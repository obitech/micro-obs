package order

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/obitech/micro-obs/item"
	_ "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

// Order defines a placed order with identifier and items.
// An Order ID of -1 means that the item can be
type Order struct {
	ID    int64        `json:"id"`
	Items []*item.Item `json:"items"`
}

type orderItem struct {
	id  string
	qty int
}

// ErrReason encodes different error reasons for an Order to fail.
type ErrReason int

// This block defines error reasons
const (
	OENotEnough ErrReason = iota
	OENotFound
	OECantRetrieve
)

// Err provides Errors with additional information.
type Err struct {
	Err    string
	Reason ErrReason
}

func (e *Err) Error() string {
	return e.Err
}

// NewOrder creates a new order according to arguments.
func NewOrder(id int64, items []*item.Item) (*Order, error) {
	return &Order{
		ID:    id,
		Items: items,
	}, nil
}

// BuildOrder queries the item service to create a new order.
func (s *Server) BuildOrder(items ...*orderItem) (*Order, error) {
	// Get OrderID from Redis
	id, err := s.RedisGetNextOrderID()
	if err != nil {
		return nil, err
	}

	// Get requested items for order
	var oi []*item.Item
	for _, v := range items {
		// Get item from Item service
		item, err := s.getItem(v.id)
		if err != nil {
			return nil, &Err{"", OECantRetrieve}
		}

		// Check if item exists and qty is enough
		err = verifyItem(item, v.qty)
		if err != nil {
			return nil, err
		}
		oi = append(oi, item)
	}

	return &Order{
		ID:    id,
		Items: oi,
	}, nil
}

// getItem will query the item service to retrieve a item for a specific quantity
func (s *Server) getItem(itemID string) (*item.Item, error) {
	// Contact Item service
	resp, err := http.Get(s.itemService)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to Item Service")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("invalid status code from Item Service: %d", resp.StatusCode)
	}

	// Get response
	b, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read response from Item Service")
	}

	// Parse respone
	var r item.Response
	err = json.Unmarshal(b, &r)
	if err != nil {
		return nil, errors.Wrapf(err, "unalbe to parse respone from Item Service")
	}

	if r.Count == 0 || r.Data == nil {
		return nil, nil
	}

	// Retrieve items from response
	for _, item := range r.Data {
		if itemID == item.ID {
			return item, nil
		}
	}
	return nil, nil
}

func verifyItem(item *item.Item, wantQty int) error {
	if item == nil {
		return &Err{"item doesn't exist", OENotFound}
	}
	if item.Qty < wantQty {
		return &Err{fmt.Sprintf("not enough items, want %d, in stock: %d", wantQty, item.Qty), OENotEnough}
	}
	return nil
}
