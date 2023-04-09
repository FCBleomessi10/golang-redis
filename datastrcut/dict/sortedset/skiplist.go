package sortedset

import (
	"math/bits"
	"math/rand"
)

const (
	MaxLevel = 16
)

type skiplistLevel struct {
	forward *skiplistNode
	span    int64
}

type skiplistNode struct {
	ele      string
	score    float64
	level    []*skiplistLevel
	backward *skiplistNode
}

type skipList struct {
	header, tail *skiplistNode
	length       int64
	level        int16
}

func makeNode(ele string, score float64, level int16) *skiplistNode {
	node := &skiplistNode{
		ele:   ele,
		score: score,
		level: make([]*skiplistLevel, level),
	}
	for i := range node.level {
		node.level[i] = new(skiplistLevel)
	}
	return node
}

func makeSkipList() *skipList {
	return &skipList{
		header: makeNode("", 0, MaxLevel),
		level:  1,
	}
}

func randomLevel() int16 {
	total := uint64(1)<<uint64(MaxLevel) - 1
	u := rand.Uint64() % total
	return MaxLevel - int16(bits.Len64(u+1)) + 1
}

func (skiplist *skipList) insert(ele string, score float64) *skiplistNode {
	// 1. 定位到插入节点的前一个节点
	prev := make([]*skiplistNode, MaxLevel)
	spans := make([]int64, MaxLevel)
	cur := skiplist.header
	// 1.1 从level[]的末尾遍历
	for i := skiplist.level - 1; i >= 0; i-- {
		if i == skiplist.level-1 {
			spans[i] = 0
		} else {
			spans[i] = spans[i+1]
		}
		if cur.level[i] != nil {
			for cur.level[i].forward != nil &&
				(cur.level[i].forward.score < score ||
					cur.level[i].forward.score == score && cur.level[i].forward.ele < ele) {
				spans[i] += cur.level[i].span
				cur = cur.level[i].forward
			}
		}
		prev[i] = cur
	}

	// 2. 判断层数是否要扩容
	level := randomLevel()
	if level > skiplist.level {
		for i := skiplist.level; i < level; i++ {
			prev[i] = skiplist.header
			prev[i].level[i].span = skiplist.length
			spans[i] = 0
		}
		skiplist.level = level
	}

	// 3. 插入节点
	node := makeNode(ele, score, level)
	for i := int16(0); i < level; i++ {
		// 3.1 更新插入节点及其前驱节点的forward
		node.level[i].forward = prev[i].level[i].forward
		prev[i].level[i].forward = node
		// 3.2 更新插入节点及其前驱节点的spans
		node.level[i].span = spans[i] + prev[i].level[i].span - spans[0]
		prev[i].level[i].span = spans[0] - spans[i] + 1
	}

	// 4. 比插入节点的层级更高的前一个节点的跨度+1
	for i := level; i < skiplist.level; i++ {
		prev[i].level[i].span++
	}

	// 5. 如果插入节点的前驱节点不是头节点，那么其前驱节点不为空
	if prev[0] == skiplist.header {
		node.backward = nil
	} else {
		node.backward = prev[0]
	}

	// 6. 维护插入节点的后继节点, 将其的前驱节点指向插入节点
	if node.level[0].forward != nil {
		node.level[0].forward.backward = node
	} else {
		skiplist.tail = node
	}

	skiplist.length++
	return node
}

// getByRank 根据排名返回节点
func (skiplist *skipList) getByRank(rank int64) *skiplistNode {
	var i int64 = 0
	cur := skiplist.header
	for level := skiplist.level - 1; level >= 0; level-- {
		for cur.level[level].forward != nil && (i+cur.level[level].span) <= rank {
			i += cur.level[level].span
			cur = cur.level[level].forward
		}
		if rank == i {
			return cur
		}
	}
	return nil
}

// getRank 根据节点返回排名
func (skiplist *skipList) getRank(ele string, score float64) int64 {
	var rank int64 = 0
	node := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for node.level[i].forward != nil &&
			(node.level[i].forward.score < score ||
				node.level[i].forward.score == score && node.level[i].forward.ele <= ele) {
			rank += node.level[i].span
			node = node.level[i].forward
		}
		if node.ele == ele {
			return rank
		}
	}
	return 0
}
