package item

import (
	"context"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/obitech/micro-obs/util"
	ot "github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
)

// Item defines a shop item with attributes. ID should be a HashID of the name.
// // See https://hashids.org for more info.
type Item struct {
	Name string `json:"name"`
	ID   string `json:"id"`
	Desc string `json:"desc"`
	Qty  int    `json:"qty"`
}

// NewItem creates a new item where the ID becomes the HashID of the lowercase name.
func NewItem(name, desc string, qty int) (*Item, error) {
	id, err := util.StringToHashID(strings.ToLower(name))
	if err != nil {
		return nil, err
	}

	return &Item{
		ID:   id,
		Name: name,
		Desc: desc,
		Qty:  qty,
	}, nil
}

// DataToItems takes a JSON-encoded byte array and marshals it into a list of item.Items
func DataToItems(data []byte) ([]*Item, error) {
	items := []*Item{}

	err := json.Unmarshal(data, &items)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to marshal %s", data)
	}

	if len(items) == 0 {
		return nil, errors.New("data can't be empty")
	}

	for _, i := range items {
		if err = i.SetID(context.Background()); err != nil {
			return nil, errors.Wrapf(err, "unable to set HashID of %s", i.Name)
		}
	}

	return items, nil
}

// SetID creates a HashID from an already set Name field.
func (i *Item) SetID(ctx context.Context) error {
	span, _ := ot.StartSpanFromContext(ctx, "SetID")
	defer span.Finish()

	id, err := util.StringToHashID(strings.ToLower(i.Name))
	if err != nil {
		return err
	}
	i.ID = id
	return nil
}

// MarshalRedis marshalls and Item to hand over to go-redis.
// Item.ID will be the key (as string), where the other fields will be a map[string]string.
func (i *Item) MarshalRedis() (string, map[string]string) {
	return i.ID, map[string]string{
		"name": i.Name,
		"desc": i.Desc,
		"qty":  strconv.Itoa(i.Qty),
	}
}

// UnmarshalRedis parses a passed string and map into an Item.
func UnmarshalRedis(key string, data map[string]string, i *Item) error {
	// Check for key existance
	ks := []string{"name", "desc", "qty"}
	for _, k := range ks {
		if _, prs := data[k]; !prs {
			return errors.Errorf("key %s not present", k)
		}
	}

	t, err := strconv.Atoi(data["qty"])
	if err != nil {
		return err
	}

	// Populate
	i.Name = data["name"]
	i.ID = key
	i.Desc = data["desc"]
	i.Qty = t

	return nil
}
