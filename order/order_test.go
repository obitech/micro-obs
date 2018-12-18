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

func helperVerifyItem(item *Item, wantQty int, wantErr ErrReason, t *testing.T) {
	err := verifyItem(item, wantQty)
	t.Logf("(LOG) verifyItem(%#v, %#v) = err: %#v", item, wantQty, err)
	if err != nil {
		if err, ok := err.(*Err); ok {
			if err.Reason != wantErr {
				t.Errorf("wrong error type, got: %#v, want: %#v", err.Reason, wantErr)
			}
		} else {
			t.Errorf("%#v should be of type *order.Err", err)
		}
	} else {
		t.Errorf("should throw error, got: %#v, want: %#v", err, wantErr)
	}
}

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

func TestVerifyItem(t *testing.T) {
	item := items[0]
	wantQty := 0
	wantErr := OENotEnough
	t.Run("Successful verification", func(t *testing.T) {
		if err := verifyItem(item, wantQty); err != nil {
			t.Errorf("verification of %#v, %d failed: %s", item, wantQty, err)
		}

		item = items[1]
		wantQty = 1
		if err := verifyItem(item, wantQty); err != nil {
			t.Errorf("verification of %#v, %d failed: %s", item, wantQty, err)
		}
	})

	t.Run("Not enough items", func(t *testing.T) {
		item := items[0]
		wantQty = 1
		helperVerifyItem(item, wantQty, wantErr, t)

	})

	t.Run("Not found", func(t *testing.T) {
		item = nil
		wantErr = OENotFound
		helperVerifyItem(item, wantQty, wantErr, t)
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

// TODO: Test getItem

// TODO: Test BuildOrder
// TODO: Test OECantRetrieve
