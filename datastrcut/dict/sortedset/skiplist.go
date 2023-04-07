package sortedset

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
