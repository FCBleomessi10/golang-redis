package sortedset

import (
	"strconv"
	"testing"
)

func TestMakeSkipList(t *testing.T) {
	s := *makeSkipList()
	t.Logf("header: %v, tail: %v, length: %v, level: %v",
		s.header, s.tail, s.length, s.level)
}

func TestInsert(t *testing.T) {
	s := *makeSkipList()
	for i := 0; i < 20; i++ {
		s.insert(strconv.Itoa(i), float64(i))
	}
	for i, node := -1, s.header; node != nil; i, node = i+1, node.level[0].forward {
		if e, _ := strconv.Atoi(node.ele); node != s.header && e != i {
			t.Errorf("Wrong! insert error!")
		}
		for _, l := range node.level {
			if node != s.header && l.forward != nil {
				if l.span != int64(l.forward.score-node.score) {
					t.Errorf("Wrong! span error!")
				}
			}
		}
		//t.Logf("ele: %v, score: %v, level: %v", node.ele, node.score, len(node.level))
	}
	t.Logf("level: %v, length: %v", s.level, s.length)
}

func TestGetByRank(t *testing.T) {
	length := 1000
	s := *makeSkipList()
	for i := 0; i < length; i++ {
		s.insert(strconv.Itoa(i), float64(i))
	}

	for i := 1; i <= length; i++ {
		node := s.getByRank(int64(i))
		if e, _ := strconv.Atoi(node.ele); node != s.header && e != i-1 {
			t.Errorf("Wrong")
		}
		//t.Logf("rank: %v, ele: %v, score: %v", i, node.ele, node.score)
	}
}

func TestGetRank(t *testing.T) {
	length := 1000
	s := *makeSkipList()
	for i := 0; i < length; i++ {
		s.insert(strconv.Itoa(i), float64(i))
	}
	for i := 0; i < length; i++ {
		rank := s.getRank(strconv.Itoa(i), float64(i))
		if rank != int64(i+1) {
			t.Errorf("Wrong")
		}
		//t.Logf("%v", rank)
	}
}
