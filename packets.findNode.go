package sp2p

type findNodeReq struct {
	N int `json:"n,omitempty"`
}

func (t *findNodeReq) T() byte        { return findNodeReqT }
func (t *findNodeReq) String() string { return findNodeReqS }
func (t *findNodeReq) OnHandle(p ISP2P, msg *KMsg) {

	node, err := nodeFromKMsg(msg)
	if err != nil {
		getLog().Error("NodeFromKMsg error", "err", err)
		return
	}
	go p.UpdateNode(node.string())

	ns := make([]string, 0)

	// 最多不能超过16
	if t.N > 16 {
		t.N = 16
	}

	nodes, _ := p.FindMinDisNodes(node.ID.Hex(), t.N)
	for _, n := range nodes {
		ns = append(ns, n)
	}
	p.Write(&KMsg{TAddr: msg.FAddr, Data: &findNodeResp{Nodes: ns}})
}

type findNodeResp struct {
	Nodes []string `json:"nodes,omitempty"`
}

func (t *findNodeResp) T() byte        { return findNodeRespT }
func (t *findNodeResp) String() string { return findNodeRespS }
func (t *findNodeResp) OnHandle(p ISP2P, msg *KMsg) {
	for _, n := range t.Nodes {
		node, err := NodeParse(n)
		if err != nil {
			getLog().Error("parse node error", "err", err)
			continue
		}
		p.UpdateNode(node.string())
	}
}
