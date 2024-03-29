package sortedset

import (
	"math/bits"
	"math/rand"
)

const (
	maxLevel = 16
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
		header: makeNode("", 0, maxLevel),
		level:  1,
	}
}

func randomLevel() int16 {
	total := uint64(1)<<uint64(maxLevel) - 1
	u := rand.Uint64() % total
	return maxLevel - int16(bits.Len64(u+1)) + 1
}

func (skiplist *skipList) insert(ele string, score float64) *skiplistNode {
	// 1. 定位到插入节点的前一个节点
	prev := make([]*skiplistNode, maxLevel)
	spans := make([]int64, maxLevel)
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

// removeNode 移除指定节点
func (skiplist *skipList) removeNode(node *skiplistNode, prev []*skiplistNode) {
	for i := skiplist.level - 1; i >= 0; i-- {
		if prev[i].level[i].forward == node {
			prev[i].level[i].forward = node.level[i].forward
			prev[i].level[i].span += node.level[i].span - 1
		} else {
			prev[i].level[i].span--
		}
	}
	if node.level[0].forward != nil {
		node.level[0].forward.backward = node.backward
	} else {
		skiplist.tail = node.backward
	}
	for skiplist.level > 1 && skiplist.header.level[skiplist.level-1].forward == nil {
		skiplist.level--
	}
	skiplist.length--
}

// remove 根据ele和score删除节点
func (skiplist *skipList) remove(ele string, score float64) bool {
	prev := make([]*skiplistNode, skiplist.level)
	node := skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for node.level[i].forward != nil &&
			node.level[i].forward.score < score ||
			(node.level[i].forward.score == score && node.level[i].forward.ele < ele) {
			node = node.level[i].forward
		}
		prev[i] = node
	}
	node = node.level[0].forward
	if node != nil && node.score == score && node.ele == ele {
		skiplist.removeNode(node, prev)
		return true
	}
	return false
}

// getRangeKey
func (skiplist *skipList) getRangeByKey(start, stop string) []*skiplistNode {
	if start > stop {
		return nil
	}

	var startNode, stopNode *skiplistNode
	var startIndex, stopIndex int64
	startNode = skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for startNode.level[i].forward != nil &&
			startNode.level[i].forward.ele < start {
			startIndex += startNode.level[i].span
			startNode = startNode.level[i].forward
		}
	}
	startNode = startNode.level[0].forward

	stopNode = skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for stopNode.level[i].forward != nil &&
			stopNode.level[i].forward.ele < stop {
			stopIndex += stopNode.level[i].span
			stopNode = stopNode.level[i].forward
		}
	}
	stopNode = stopNode.level[0].forward

	nodes := make([]*skiplistNode, stopIndex-startIndex+1)
	for i, n := 0, startNode; n != stopNode.level[0].forward; i, n = i+1, n.level[0].forward {
		nodes[i] = n
	}
	return nodes
}

// getRangeByScore
func (skiplist *skipList) getRangeByScore(min, max float64) []*skiplistNode {
	if min > max {
		return nil
	}

	var startNode, stopNode *skiplistNode
	var startIndex, stopIndex int64
	startNode = skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for startNode.level[i].forward != nil &&
			startNode.level[i].forward.score < min {
			startIndex += startNode.level[i].span
			startNode = startNode.level[i].forward
		}
	}
	startNode = startNode.level[0].forward

	stopNode = skiplist.header
	for i := skiplist.level - 1; i >= 0; i-- {
		for stopNode.level[i].forward != nil &&
			stopNode.level[i].forward.score < max {
			stopIndex += stopNode.level[i].span
			stopNode = stopNode.level[i].forward
		}
	}
	stopNode = stopNode.level[0].forward

	nodes := make([]*skiplistNode, stopIndex-startIndex+1)
	for i, n := 0, startNode; n != stopNode.level[0].forward; i, n = i+1, n.level[0].forward {
		nodes[i] = n
	}
	return nodes
}
