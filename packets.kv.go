package sp2p

import (
	"github.com/dgraph-io/badger"
	"github.com/kooksee/common"
)

type kv struct {
	K       string      `json:"k,omitempty"`
	V       interface{} `json:"v,omitempty"`
	Expired int         `json:"expired,omitempty"`
}

type KVSetReq struct {
	kv
}

func (t *KVSetReq) T() byte          { return KVGetReqT }
func (t *KVSetReq) String() string   { return KVSetReqS }
func (t *KVSetReq) Create() IMessage { return &KVSetReq{} }
func (t *KVSetReq) OnHandle(p *SP2p, msg *KMsg) {
	nodes := p.GetTable().FindNodeWithTargetBySelf(common.StringToHash(t.K))
	if len(nodes) < 2 {
		if err := cfg.Db.Update(func(txn *badger.Txn) error {
			v, err := json.Marshal(t.V)
			if err != nil {
				return err
			}
			return txn.Set([]byte(t.K), v)
		}); err != nil {
			logger.Error("kvset error", "err", err)
		}
		return
	}

	for _, node := range nodes {
		p.Write(&KMsg{
			FAddr: msg.FAddr,
			Data:  msg.Data,
			TAddr: node.Addr().String(),
		})
	}
}

type KVGetReq struct {
	kv
}

func (t *KVGetReq) T() byte          { return KVGetReqT }
func (t *KVGetReq) String() string   { return KVGetReqS }
func (t *KVGetReq) Create() IMessage { return &KVGetReq{} }
func (t *KVGetReq) OnHandle(p *SP2p, msg *KMsg) {
	nodes := p.GetTable().FindNodeWithTargetBySelf(common.StringToHash(t.K))
	if len(nodes) < 2 {
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
			resp.V = string(v)

			p.Write(&KMsg{
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
			Data:  msg.Data,
			FAddr: msg.FAddr,
			TAddr: node.Addr().String(),
		})
	}
}

type KVGetResp struct {
	kv
}

func (t *KVGetResp) T() byte          { return KVGetRespT }
func (t *KVGetResp) String() string   { return KVGetRespS }
func (t *KVGetResp) Create() IMessage { return &KVGetResp{} }
func (t *KVGetResp) OnHandle(p *SP2p, msg *KMsg) {
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
