package sp2p

type findNodeReq struct {
	IMessage
	N int `json:"n,omitempty"`
}

func (t *findNodeReq) T() byte        { return findNodeReqT }
func (t *findNodeReq) String() string { return findNodeReqS }
func (t *findNodeReq) OnHandle(p ISP2P, msg *KMsg) error {
	if err := p.NodeUpdate(msg.FN); err != nil {
		return err
	}

	// 最多不能超过16
	if t.N > 16 {
		t.N = 16
	}

	ns := make([]string, 0)
	n, err := NodeParse(msg.TN)
	if err != nil {
		return err
	}

	nodes, err := p.FindMinDisNodes(n.ID.Hex(), t.N)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		ns = append(ns, n)
	}

	p.Write(&KMsg{TN: msg.FN, Data: &findNodeResp{Nodes: ns}})
	return nil
}

type findNodeResp struct {
	IMessage
	Nodes []string `json:"nodes,omitempty"`
}

func (t *findNodeResp) T() byte        { return findNodeRespT }
func (t *findNodeResp) String() string { return findNodeRespS }
func (t *findNodeResp) OnHandle(p ISP2P, msg *KMsg) error {
	for _, n := range t.Nodes {
		node, err := NodeParse(n)
		if err != nil {
			getLog().Error("parse node error", "err", err)
			continue
		}
		p.NodeUpdate(node.string())
	}
	return nil
}
