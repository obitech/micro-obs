package item

// Item defines a shop item with attributes.
type Item struct {
	ID   int    `redis:"name"`
	Name string `redis:"string"`
	Desc string `redis:"desc"`
	Qty  int    `redis:"qty"`
}

// NewItem creates a new item.
func NewItem(id int, name, desc string, qty int) (*Item, error) {
	return &Item{
		ID:   id,
		Name: name,
		Desc: desc,
		Qty:  qty,
	}, nil
}

// marshalRedis takes an Item struct and marshalls it to hand over to go-redis.
// Item.ID will be the key (as string), where the other fields will be a map[string]string.
// func (i *Item) marshalRedis() (string, map[string]string) {

// }
