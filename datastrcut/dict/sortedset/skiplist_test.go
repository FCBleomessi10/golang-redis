package sortedset

import (
	"testing"
)

func TestMakeSkipList(t *testing.T) {
	s := *makeSkipList()
	t.Logf("header: %v, tail: %v, length: %v, level: %v",
		s.header, s.tail, s.length, s.level)
}
