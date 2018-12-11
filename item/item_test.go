package item

import (
	"testing"
)

var (
	items = []struct {
		id   int
		name string
		desc string
		qty  int
		want error
	}{
		{0, "test", "test", 0, nil},
		{12, "orange", "a juicy fruit", 100, nil},
	}
)

func TestItem(t *testing.T) {
	t.Run("Create new item", func(t *testing.T) {
		for _, tt := range items {
			if _, err := NewItem(tt.id, tt.name, tt.desc, tt.qty); err != tt.want {
				t.Errorf("Failed to create new item: %s", err)
			}
		}
	})
}
