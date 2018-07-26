package sp2p

func (s *SP2p) Write(msg *KMsg) {
	go s.writeTx(msg)
}

func (s *SP2p) GetNode() string {
	return s.tab.selfNode.string()
}

func (s *SP2p) GetNodes() []string {
	return s.tab.getRawNodes()
}

func (s *SP2p) TableSize() int {
	return s.tab.size()
}

func (s *SP2p) UpdateNode(rawUrl string) error {
	n, err := NodeParse(rawUrl)
	if err != nil {
		return err
	}
	s.tab.updateNode(n)
	return nil
}
func (s *SP2p) DeleteNode(id string) error {
	n, err := HexToHash(id)
	if err != nil {
		return err
	}
	s.tab.deleteNode(n)
	return nil
}

func (s *SP2p) AddNode(rawUrl string) error {
	n, err := NodeParse(rawUrl)
	if err != nil {
		return err
	}
	s.tab.addNode(n)
	return nil
}

func (s *SP2p) FindMinDisNodes(targetID string, n int) (nodes []string, err error) {
	h, err := HexToHash(targetID)
	if err != nil {
		return nil, err
	}

	for _, n := range s.tab.findMinDisNodes(h, n) {
		nodes = append(nodes, n.string())
	}
	return nodes, nil
}

func (s *SP2p) FindRandomNodes(n int) (nodes []string) {
	for _, n := range s.tab.findRandomNodes(n) {
		nodes = append(nodes, n.string())
	}
	return
}

func (s *SP2p) FindNodeWithTargetBySelf(d string) (nodes []string) {
	for _, n := range s.tab.findNodeWithTargetBySelf(StringToHash(d)) {
		nodes = append(nodes, n.string())
	}
	return
}

func (s *SP2p) FindNodeWithTarget(targetId string, measure string) (nodes []string) {
	for _, n := range s.tab.findNodeWithTarget(StringToHash(targetId), StringToHash(measure)) {
		nodes = append(nodes, n.string())
	}
	return
}

func (s *SP2p) PingN() {
	go s.pingN()
}

func (s *SP2p) FindN() {
	go s.findN()
}
