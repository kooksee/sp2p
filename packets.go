package sp2p

import (
	"github.com/dgraph-io/badger"
	"github.com/kooksee/srelay/protocol"
	"github.com/kooksee/uspnet/common"
)

func findNode(p *SP2p, msg *KMsg) {
	switch msg.Data.(type) {
	case protocol.FindNodeReq:
		s := msg.Data.(protocol.FindNodeReq)
		nodes := p.tab.FindMinDisNodes(common.StringToHash(s.NID), s.N)
		ns := make([]string, 0)
		for _, n := range nodes {
			ns = append(ns, n.String())
		}
		p.Write(&KMsg{
			Event: msg.Event,
			TAddr: msg.FAddr,
			Data:  protocol.FindNodeResp{Nodes: ns},
		})

		if node, err := NodeFromKMsg(msg); err != nil {
			logger.Error("NodeFromKMsg error", "err", err)
		} else {
			p.tab.UpdateNode(node)
		}
	case protocol.FindNodeResp:
		s := msg.Data.(protocol.FindNodeResp)
		for _, n := range s.Nodes {
			node, err := ParseNode(n)
			if err != nil {
				logger.Error("parse node error", "err", err)
				continue
			}
			p.tab.AddNode(node)
		}
	}
}

func ping(p *SP2p, msg *KMsg) {
	if node, err := NodeFromKMsg(msg); err != nil {
		logger.Error("NodeFromKMsg error", "err", err)
	} else {
		p.tab.UpdateNode(node)
	}
}

func kvGet(p *SP2p, msg *KMsg) {
	switch msg.Data.(type) {
	case protocol.KVGetReq:
		req := msg.Data.(protocol.KVGetReq)
		nodes := p.GetTable().FindNodeWithTargetBySelf(common.StringToHash(req.K))
		if len(nodes) < 1 {
			if err := cfg.Db.View(func(txn *badger.Txn) error {
				item, err := txn.Get([]byte(req.K))
				if err != nil {
					return err
				}
				v, err := item.Value()
				if err != nil {
					return err
				}

				resp := protocol.KVGetResp{}
				resp.K = req.K
				resp.V = v

				p.Write(&KMsg{
					Event: msg.Event,
					Data:  resp,
					TAddr: msg.FAddr,
				})

				return nil

			}); err != nil {
				logger.Error(err.Error())
			}
			return
		}

		for _, node := range nodes {
			p.Write(&KMsg{
				Event: msg.Event,
				Data:  msg.Data,
				FAddr: msg.FAddr,
				FID:   msg.FID,
				TAddr: node.Addr().String(),
			})
		}
	case protocol.KVGetResp:
		kv := msg.Data.(protocol.KVGetResp)
		if err := cfg.Db.Update(func(txn *badger.Txn) error {
			v, err := json.Marshal(kv.V)
			if err != nil {
				return err
			}
			return txn.Set([]byte(kv.K), v)
		}); err != nil {
			logger.Error(err.Error())
		}
	}
}

func kvSet(p *SP2p, msg *KMsg) {
	req, ok := msg.Data.(protocol.KV)
	if !ok {
		return
	}
	nodes := p.GetTable().FindNodeWithTargetBySelf(common.StringToHash(req.K))
	if len(nodes) < 2 {
		if err := cfg.Db.Update(func(txn *badger.Txn) error {
			v, err := json.Marshal(req.V)
			if err != nil {
				return err
			}
			return txn.Set([]byte(req.K), v)
		}); err != nil {
			logger.Error("kvset error", "err", err)
		}
		return
	}

	for _, node := range nodes {
		p.Write(&KMsg{
			Event: msg.Event,
			Data:  msg.Data,
			TAddr: node.Addr().String(),
		})
	}
}

func init() {
	hm.Registry("findNode", findNode)
	hm.Registry("ping", ping)
	hm.Registry("kvSet", kvSet)
	hm.Registry("kvGet", kvGet)
}