package sp2p

import (
	"time"

	"github.com/dgraph-io/badger"
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/kooksee/uspnet/common"
	"github.com/kooksee/uspnet/common/hexutil"
	"github.com/kooksee/uspnet/log"
)

type bucket struct {
	peers *arraylist.List
}

func newBuckets() *bucket {
	return &bucket{
		peers: arraylist.New(),
	}
}

func (b *bucket) updateNodes(nodes ... *Node) {
	for _, n := range nodes {
		n.updateAt = time.Now()
		b.addNodes(n)
	}
}

// addNode add node to bucket, if bucket is full, will remove an old one
func (b *bucket) addNodes(nodes ... *Node) {
	for _, node := range nodes {
		log.Info("add node: %s", hexutil.BytesToHex(node.ID.Bytes()))
		b.peers.Add(nodes)
	}

	// 把最活跃的放到最前面,然后移除最不活跃的
	b.peers.Sort(func(a, b interface{}) int { return int(b.(*Node).updateAt.Sub(a.(*Node).updateAt)) })
	size := b.peers.Size()
	if size > cfg.BucketSize {
		cfg.Db.Update(func(txn *badger.Txn) error {
			for i := cfg.BucketSize; i < size; i++ {
				val, e := b.peers.Get(i)
				if !e {
					continue
				}
				b.peers.Remove(i)
				if err := txn.Delete(append([]byte(cfg.NodesBackupKey), val.(*Node).ID.Bytes()...)); err != nil {
					return err
				}
			}
			return nil
		})
	}
}

// findNode check if the bucket already have this node, if so, return its index, otherwise, return -1
func (b *bucket) findNode(node *Node) int {
	return b.peers.IndexOf(node)
}

func (b *bucket) Random() *Node {
	if b.size() == 0 {
		return nil
	}
	a := int(randUint(uint32(b.size())))
	val, _ := b.peers.Get(a)
	return val.(*Node)
}

func (b *bucket) getLast(n int) []*Node {
	nodes := make([]*Node, n)
	for i, j := b.peers.Size()-1, 0; ; {
		if i >= 0 && j <= n {
			v, _ := b.peers.Get(i)
			nodes = append(nodes, v.(*Node))
			j++
			i--
		}
	}
	return nodes
}

func (b *bucket) deleteNodes(targets ... common.Hash) {
	for _, node := range targets {
		if a := b.peers.IndexOf(node); a != -1 {
			log.Info("delete node: %s", hexutil.BytesToHex(node.Bytes()))
			b.peers.Remove(a)
		}
	}
}

func (b *bucket) size() int {
	return b.peers.Size()
}

// printNodeList only use for debug test
func (b *bucket) printNodeList() {
	log.Debug("bucket size %d", b.peers.Size())

	b.peers.Each(func(index int, value interface{}) {
		log.Debug("%s", hexutil.BytesToHex(value.(common.Hash).Bytes()))
	})
}
