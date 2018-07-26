package sp2p

import (
	"math/rand"
	"net"
	"sort"
	"sync"
	"time"

	"github.com/emirpasic/gods/sets/hashset"
)

const nBuckets = len(Hash{})*8 + 1

type table struct {
	ITable

	mutex sync.Mutex

	buckets  [nBuckets]*bucket
	selfNode *node //info of local node
}

func newTable(id Hash, addr *net.UDPAddr) *table {

	table := &table{selfNode: newNode(id, addr.IP, uint16(addr.Port))}

	for i := 0; i < nBuckets; i++ {
		table.buckets[i] = newBuckets()
	}

	return table
}

func (t *table) getNode() *node {
	return t.selfNode
}

func (t *table) getAllNodes() []*node {
	nodes := make([]*node, 0)
	for _, b := range t.buckets {
		b.peers.Each(func(index int, value interface{}) {
			nodes = append(nodes, value.(*node))
		})
	}
	return nodes
}

func (t *table) getRawNodes() []string {
	nodes := make([]string, 0)
	for _, n := range t.getAllNodes() {
		nodes = append(nodes, n.string())
	}
	return nodes
}

func (t *table) addNode(node *node) {
	t.buckets[logdist(t.selfNode.ID, node.ID)].addNodes(node)
}

func (t *table) updateNode(node *node) {
	t.buckets[logdist(t.selfNode.ID, node.ID)].updateNodes(node)
}

func (t *table) size() int {
	n := 0
	for _, b := range t.buckets {
		n += b.size()
	}
	return n
}

// ReadRandomNodes fills the given slice with random nodes from the
// table. It will not write the same node more than once. The nodes in
// the slice are copies and can be modified by the caller.
func (t *table) findRandomNodes(n int) []*node {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	nodes := make([]*node, 0)
	for _, b := range t.buckets {
		b.peers.Each(func(_ int, value interface{}) {
			nodes = append(nodes, value.(*node))
		})
	}

	n = cond(n > nBuckets, nBuckets, 5).(int)
	if len(nodes) < n+5 {
		return nodes
	}

	nodeSet := hashset.New()
	rand.Seed(time.Now().Unix())
	k := int32(len(nodes))
	for nodeSet.Size() < n {
		nodeSet.Add(nodes[rand.Int31n(k)])
	}

	rnodes := make([]*node, 0)
	for _, n := range nodeSet.Values() {
		rnodes = append(rnodes, n.(*node))
	}
	return rnodes
}

// findNodeWithTarget find nodes that distance of target is less than measure with target
func (t *table) findNodeWithTarget(target Hash, measure Hash) []*node {
	minDis := make([]*node, 0)
	for _, n := range t.findMinDisNodes(target, cfg.NodeResponseNumber) {
		if distCmp(target, n.ID, measure) < 0 {
			minDis = append(minDis, n)
		}
	}

	return minDis
}

func (t *table) findNodeWithTargetBySelf(target Hash) []*node {
	return t.findNodeWithTarget(target, t.selfNode.ID)
}

func (t *table) deleteNode(target Hash) {
	t.buckets[logdist(t.selfNode.ID, target)].deleteNodes(target)
}

func (t *table) findMinDisNodes(target Hash, number int) []*node {

	result := &nodesByDistance{
		target:   target,
		maxElems: cond(number > nBuckets, nBuckets, 5).(int),
		entries:  make([]*node, 0),
	}

	for _, b := range t.buckets {
		b.peers.Each(func(_ int, value interface{}) {
			result.push(value.(*node))
		})
	}

	return result.entries
}

// nodesByDistance is a list of nodes, ordered by
// distance to to.
type nodesByDistance struct {
	entries  []*node
	target   Hash
	maxElems int
}

// push adds the given node to the list, keeping the total size below maxElems.
func (h *nodesByDistance) push(n *node) {
	ix := sort.Search(len(h.entries), func(i int) bool {
		return distCmp(h.target, h.entries[i].ID, n.ID) > 0
	})
	if len(h.entries) < h.maxElems {
		h.entries = append(h.entries, n)
	}
	if ix == len(h.entries) {
		// farther away than all nodes we already have.
		// if there was room for it, the node is now the last element.
	} else {
		// slide existing entries down to make room
		// this will overwrite the entry we just appended.
		copy(h.entries[ix+1:], h.entries[ix:])
		h.entries[ix] = n
	}
}
