package order

import (
	_ "net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/alicebob/miniredis"
	"github.com/obitech/micro-obs/item"
)

var (
	sampleItems = []struct {
		name string
		desc string
		qty  int
	}{
		{"test", "test", 0},
		{"orange", "a juicy fruit", 100},
		{"😍", "lovely smily", 999},
		{"     ", "﷽", 249093419},
		{" 123asd🙆   🙋 asdlloqwe", "test", 0},
	}

	sampleOrderIDs = []int64{-1, 0, 12, 42, 1242352235}
	items          = []*Item{}
	orders         = []*Order{}
	uniqueOrders   = []*Order{}
)

func helperInitItemServer(t *testing.T) (*miniredis.Miniredis, *item.Server) {
	// Create mr for item service
	_, mr := helperPrepareMiniredis(t)

	// Setup server
	s, err := item.NewServer(
		item.SetRedisAddress(strings.Join([]string{"redis://", mr.Addr()}, "")),
	)
	if err != nil {
		mr.Close()
		t.Errorf("unable to create item server: %s", err)
	}

	return mr, s
}

func TestNewOrderItem(t *testing.T) {
	for _, v := range sampleItems {
		i, err := item.NewItem(v.name, v.desc, v.qty)
		if err != nil {
			t.Errorf("unable to create item: %s", err)
		}

		oi, err := NewItem(i)
		if err != nil {
			t.Errorf("unable to create order item: %s", err)
		}
		items = append(items, oi)
	}
}

func TestNewOrder(t *testing.T) {
	for _, v := range items {
		t.Run("With single item", func(t *testing.T) {
			for _, id := range sampleOrderIDs {
				o, err := NewOrder(id, v)
				if err != nil {
					t.Errorf("unable to create order: %s", err)
				}
				orders = append(orders, o)
			}
		})
	}
	t.Run("With multiple items", func(t *testing.T) {
		for _, id := range sampleOrderIDs {
			o, err := NewOrder(id, items...)
			if err != nil {
				t.Errorf("unable to create order: %s", err)
			}
			orders = append(orders, o)
		}
	})

	t.Run("Order with nil Items", func(t *testing.T) {
		_, err := NewOrder(42, nil)
		if err == nil {
			t.Error("should throw error with nil item")
		}
	})

}

func TestMarshalRedis(t *testing.T) {
	var idMarshalled string
	var itemsMarshalled map[string]int

	for _, v := range orders {
		idMarshalled, itemsMarshalled = v.MarshalRedis()

		verify := &Order{}
		err := UnmarshalRedis(idMarshalled, itemsMarshalled, verify)
		if err != nil {
			t.Errorf("unmarshaling failed: %s", err)
		}
		if verify.ID != v.ID {
			t.Errorf("ID: %#v != %#v", verify.ID, v.ID)
		}

		if !reflect.DeepEqual(v.Items, verify.Items) {
			t.Errorf("%+v != %+v", v.Items, verify.Items)
		}
	}
}
