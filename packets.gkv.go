package sp2p

/*
通过广播的方式进行数据存储,同时随机抽样检测数据的一致性
 */

import (
	"github.com/dgraph-io/badger"
)

type GKVSetReq struct{ kv }

func (t *GKVSetReq) T() byte          { return GKVGetReqT }
func (t *GKVSetReq) String() string   { return GKVSetReqS }
func (t *GKVSetReq) Create() IMessage { return &GKVSetReq{} }
func (t *GKVSetReq) OnHandle(p *SP2p, msg *KMsg) {
	if err := cfg.Db.Update(func(txn *badger.Txn) error {
		// 检查是否存在
		item, _ := txn.Get(t.K)
		if item != nil {
			return nil
		}

		// 不存在就存储
		if err := txn.Set([]byte(t.K), t.V); err != nil {
			return err
		}

		// 随机广播
		for _, node := range p.tab.FindRandomNodes(cfg.NodeBroadcastNumber) {
			p.writeTx(&KMsg{
				FAddr: msg.FAddr,
				Data:  msg.Data,
				TAddr: node.Addr().String(),
			})
		}
		return nil
	}); err != nil {
		logger.Error(err.Error())
	}
}

type GKVGetReq struct{ kv }

func (t *GKVGetReq) T() byte          { return GKVGetReqT }
func (t *GKVGetReq) String() string   { return GKVGetReqS }
func (t *GKVGetReq) Create() IMessage { return &GKVGetReq{} }
func (t *GKVGetReq) OnHandle(p *SP2p, msg *KMsg) {
	if err := cfg.Db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(t.K))
		if err != nil {
			return err
		}
		v, err := item.Value()
		if err != nil {
			return err
		}

		resp := &KVGetResp{}
		resp.K = t.K
		resp.V = v

		p.Write(&KMsg{
			Data:  resp,
			TAddr: msg.FAddr,
		})

		return nil

	}); err != nil {
		logger.Error(err.Error())

		for _, node := range p.tab.FindRandomNodes(8) {
			p.writeTx(&KMsg{
				Data:  msg.Data,
				FAddr: msg.FAddr,
				TAddr: node.Addr().String(),
			})
		}
	}
}

type GKVGetResp struct{ kv }

func (t *GKVGetResp) T() byte          { return KVGetRespT }
func (t *GKVGetResp) String() string   { return KVGetRespS }
func (t *GKVGetResp) Create() IMessage { return &KVGetResp{} }
func (t *GKVGetResp) OnHandle(p *SP2p, msg *KMsg) {
	if err := cfg.Db.Update(func(txn *badger.Txn) error {
		v, err := json.Marshal(t.V)
		if err != nil {
			return err
		}
		return txn.Set([]byte(t.K), v)
	}); err != nil {
		logger.Error(err.Error())
	}
}
