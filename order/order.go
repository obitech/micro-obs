package order

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/obitech/micro-obs/item"
	_ "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

// Order defines a placed order with identifier and items.
// An Order ID of -1 means that the item can be
type Order struct {
	ID    int64   `json:"id"`
	Items []*Item `json:"items"`
}

// Item holds stripped down information of a regular item, to be used in an Order.
type Item struct {
	ID  string `json:"id"`
	Qty int    `json:"qty"`
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

func (o *Order) String() string {
	return fmt.Sprintf("ID=%d Items=%+v,", o.ID, o.Items)
}

func (i *Item) String() string {
	return fmt.Sprintf("ID=%s Qty=%d,", i.ID, i.Qty)
}

func (e *Err) Error() string {
	return e.Err
}

// NewItem creates a new Item from an existing Item.
func NewItem(item *item.Item) (*Item, error) {
	return &Item{
		ID:  item.ID,
		Qty: item.Qty,
	}, nil
}

// NewOrder creates a new order according to arguments. This will sort the passed items.
func NewOrder(id int64, items ...*Item) (*Order, error) {
	oi := make([]*Item, len(items))
	for i, v := range items {
		if v == nil {
			return nil, errors.New("item can't be nil")
		}
		oi[i] = v
	}

	// Sort the list of items according to ID
	sort.Slice(oi, func(i, j int) bool {
		return strings.Compare(oi[i].ID, oi[j].ID) == -1
	})

	return &Order{
		ID:    id,
		Items: oi,
	}, nil
}

// BuildOrder queries the item service to create a new order.
func (s *Server) BuildOrder(ctx context.Context, items ...*Item) (*Order, error) {
	// Get OrderID from Redis
	id, err := s.RedisGetNextOrderID(ctx)
	if err != nil {
		return nil, err
	}

	// Get requested items for order
	var oi []*Item
	for _, v := range items {
		// Get item from Item service
		item, err := s.getItem(ctx, v.ID)
		if err != nil {
			return nil, &Err{"", OECantRetrieve}
		}

		// Check if item exists and qty is enough
		err = verifyItem(item, v.Qty)
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

// MarshalRedis marshals an Order to hand over to go-redis.
func (o *Order) MarshalRedis() (string, map[string]int) {
	id := strconv.FormatInt(o.ID, 10)
	if o.Items == nil {
		return id, nil
	}

	items := make(map[string]int)
	for _, v := range o.Items {
		items[v.ID] = v.Qty
	}

	return id, items
}

// UnmarshalRedis parses a string and map into an Order. Order.Items will be sorted according to the ID.
func UnmarshalRedis(id string, items map[string]int, order *Order) error {
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	// Sort map according to keys
	var keys []string
	for k := range items {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	oi := []*Item{}
	for _, k := range keys {
		oi = append(oi, &Item{
			ID:  k,
			Qty: items[k],
		})
	}

	order.ID = i
	order.Items = oi

	return nil
}

// getItem will query the item service to retrieve a item for a specific quantity
func (s *Server) getItem(ctx context.Context, itemID string) (*Item, error) {
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
			return &Item{
				ID:  itemID,
				Qty: item.Qty,
			}, nil
		}
	}
	return nil, nil
}

func verifyItem(item *Item, wantQty int) error {
	if item == nil {
		return &Err{"item doesn't exist", OENotFound}
	}
	if item.Qty < wantQty {
		return &Err{fmt.Sprintf("not enough items, want %d, in stock: %d", wantQty, item.Qty), OENotEnough}
	}
	return nil
}
