package sp2p

/*
采用分片的方式进行kv存储,同时定时抽样的方式检测自己的数据是否有合适的节点可以存储
 */

import (
	"github.com/dgraph-io/badger"
)

type kv struct {
	K       []byte `json:"k,omitempty"`
	V       []byte `json:"v,omitempty"`
	Expired int    `json:"expired,omitempty"`
	Time    int    `json:"time,omitempty"`
}

type KVSetReq struct{ kv }

func (t *KVSetReq) T() byte          { return KVGetReqT }
func (t *KVSetReq) String() string   { return KVSetReqS }
func (t *KVSetReq) Create() IMessage { return &KVSetReq{} }
func (t *KVSetReq) OnHandle(p *SP2p, msg *KMsg) {
	nodes := p.GetTable().FindNodeWithTargetBySelf(BytesToHash(t.K))
	if len(nodes) < cfg.NodePartitionNumber {
		if err := GetDb().Update(func(txn *badger.Txn) error {
			v, err := json.Marshal(t.V)
			if err != nil {
				return err
			}
			return txn.Set([]byte(t.K), v)
		}); err != nil {
			GetLog().Error("kvset error", "err", err)
		}
		return
	}

	for _, node := range nodes {
		p.writeTx(&KMsg{FAddr: msg.FAddr, Data: msg.Data, TAddr: node.Addr().String()})
	}
}

type KVGetReq struct{ kv }

func (t *KVGetReq) T() byte          { return KVGetReqT }
func (t *KVGetReq) String() string   { return KVGetReqS }
func (t *KVGetReq) Create() IMessage { return &KVGetReq{} }
func (t *KVGetReq) OnHandle(p *SP2p, msg *KMsg) {
	nodes := p.GetTable().FindNodeWithTargetBySelf(BytesToHash(t.K))
	if len(nodes) < cfg.NodePartitionNumber {
		if err := GetDb().View(func(txn *badger.Txn) error {
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

			p.writeTx(&KMsg{Data: resp, TAddr: msg.FAddr})

			return nil

		}); err != nil {
			GetLog().Error(err.Error())
		}
		return
	}

	for _, node := range nodes {
		p.writeTx(&KMsg{Data: msg.Data, FAddr: msg.FAddr, TAddr: node.Addr().String()})
	}
}

type KVGetResp struct{ kv }

func (t *KVGetResp) T() byte          { return KVGetRespT }
func (t *KVGetResp) String() string   { return KVGetRespS }
func (t *KVGetResp) Create() IMessage { return &KVGetResp{} }
func (t *KVGetResp) OnHandle(p *SP2p, msg *KMsg) {
	if err := GetDb().Update(func(txn *badger.Txn) error {
		v, err := json.Marshal(t.V)
		if err != nil {
			return err
		}
		return txn.Set([]byte(t.K), v)
	}); err != nil {
		GetLog().Error(err.Error())
	}
}
