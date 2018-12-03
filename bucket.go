package sp2p

import (
	"time"

	"errors"
	"github.com/emirpasic/gods/lists/arraylist"
	"github.com/kooksee/kdb"
)

var bucketPrefix = []byte("bkt")

type bucket struct {
	peers *arraylist.List
	h     kdb.IKHash
}

func newBuckets() *bucket {
	return &bucket{
		peers: arraylist.New(),
		h:     getDb().KHash(bucketPrefix),
	}
}

// updateNodes update node info with time
func (b *bucket) updateNodes(nodes ... *node) error {
	for _, n := range nodes {
		n.UpdateAt = time.Now()
		if err := b.addNodes(n); err != nil {
			return err
		}
	}

	return nil
}

// addNode add node to bucket, if bucket is full, will remove an old one
func (b *bucket) addNodes(nodes ... *node) error {

	// 把最活跃的放到最前面,然后移除最不活跃的
	return errPipe("bucket addNodes error", b.h.WithBatch(func(k kdb.IKHBatch) error {
		for _, node := range nodes {
			b.peers.Add(node)
			nm, err := node.marshal()
			if err := errPipe("add node error", err, k.Set(nodesBackupKey(node.ID.Bytes()), nm)); err != nil {
				return err
			}
		}

		b.peers.Sort(func(a, b interface{}) int { return int(b.(*node).UpdateAt.Sub(a.(*node).UpdateAt)) })
		size := b.peers.Size()
		if size < cfg.BucketSize {
			return errors.New("")
		}

		for i := cfg.BucketSize; i < size; i++ {
			val, e := b.peers.Get(i)
			if !e {
				continue
			}
			b.peers.Remove(i)
			if err := errPipe(
				"delete peer error",
				k.MDel(nodesBackupKey(val.(*node).ID.Bytes()))); err != nil {
				return err
			}
		}
		return nil
	}))
}

// findNode check if the bucket already have this node, if so, return its index, otherwise, return -1
func (b *bucket) findNode(node *node) int {
	return b.peers.IndexOf(node)
}

func (b *bucket) random() *node {
	if b.size() == 0 {
		return nil
	}

	val, _ := b.peers.Get(int(rand32(uint32(b.size()))))
	return val.(*node)
}

func (b *bucket) deleteNodes(targets ... Hash) error {
	return errPipe(
		"bucket deleteNodes error",
		b.h.WithBatch(func(k kdb.IKHBatch) error {
			for _, n := range targets {
				if a := b.peers.IndexOf(n); a != -1 {
					val, e := b.peers.Get(a)
					if !e {
						continue
					}

					if err := k.MDel(nodesBackupKey(val.(*node).ID.Bytes())); err != nil {
						return err
					}

					getLog().Info("delete node: %s ok", n.Hex())
					b.peers.Remove(a)
				}
			}
			return nil
		}),
	)
}

func (b *bucket) size() int {
	return b.peers.Size()
}
