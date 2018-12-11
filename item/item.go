package item

import (
	"strconv"
	"strings"

	"github.com/obitech/micro-obs/util"
)

// Item defines a shop item with attributes. ID should be a HashID of the name.
// // See https://hashids.org for more info.
type Item struct {
	Name string
	ID   string
	Desc string
	Qty  int
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

// marshalRedis takes an Item struct and marshalls it to hand over to go-redis.
// Item.ID will be the key (as string), where the other fields will be a map[string]string.
func (i *Item) marshalRedis() (string, map[string]string) {
	return i.ID, map[string]string{
		"name": i.Name,
		"desc": i.Desc,
		"qty":  strconv.Itoa(i.Qty),
	}
}
