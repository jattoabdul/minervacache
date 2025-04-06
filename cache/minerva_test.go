package cache

import (
	"testing"
)

func TestNewMinervaCache(t *testing.T) {
	c := NewMinervaCache(10, 0)
	defer c.Stop()

	if c == nil {
		t.Fatal("expected cache to be non-nil")
	}
}
