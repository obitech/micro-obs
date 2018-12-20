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
	ot "github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
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

func (o *Order) String() string {
	return fmt.Sprintf("ID=%d Items=%+v,", o.ID, o.Items)
}

func (i *Item) String() string {
	return fmt.Sprintf("ID=%s Qty=%d,", i.ID, i.Qty)
}

// NewItem creates a new Item from an existing Item.
func NewItem(item *item.Item) (*Item, error) {
	return &Item{
		ID:  item.ID,
		Qty: item.Qty,
	}, nil
}

// Sort will sort the order items according to ID
func (o *Order) Sort() error {
	if o.Items == nil || len(o.Items) == 0 {
		return errors.New("order needs items")
	}

	// Sort the list of items according to ID
	sort.Slice(o.Items, func(i, j int) bool {
		return strings.Compare(o.Items[i].ID, o.Items[j].ID) == -1
	})

	return nil
}

// NewOrder creates a new order according to arguments. This will sort the passed items.
func NewOrder(id int64, items ...*Item) (*Order, error) {
	oi := make([]*Item, len(items))
	for i, v := range items {
		if v == nil {
			return nil, errors.New("order needs items")
		}
		oi[i] = v
	}

	order := &Order{
		ID:    id,
		Items: oi,
	}

	order.Sort()
	return order, nil
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
	span, ctx := ot.StartSpanFromContext(ctx, "getItem")
	defer span.Finish()

	// Create item service request
	url := fmt.Sprintf("%s/items/%s", s.itemService, itemID)
	req, err := http.NewRequest("GET", url, nil)

	// Inect tracer
	ext.SpanKindRPCClient.Set(span)
	ext.HTTPMethod.Set(span, "GET")
	span.Tracer().Inject(
		span.Context(),
		ot.HTTPHeaders,
		ot.HTTPHeadersCarrier(req.Header),
	)

	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to connect to item service")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("invalid status code from item service: %d", resp.StatusCode)
	}

	// Read response
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to read response from item service")
	}
	defer resp.Body.Close()

	// Parse respone
	var r item.Response
	err = json.Unmarshal(b, &r)
	if err != nil {
		return nil, errors.Wrapf(err, "unalbe to parse respone from item service")
	}

	if r.Count == 0 || r.Data == nil {
		return nil, errors.New("no items returned from item service")
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
	return nil, errors.New("no items left to yield")
}
