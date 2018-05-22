package sp2p

import "github.com/kooksee/common"

func (s *SP2p) Write(msg *KMsg) {
	go s.writeTx(msg)
}

func (s *SP2p) GetTable() *Table {
	return s.tab
}

func (s *SP2p) GetNode() string {
	return s.tab.selfNode.String()
}

func (s *SP2p) GetNodes() []string {
	return s.tab.GetRawNodes()
}

func (s *SP2p) TableSize() int {
	return s.tab.Size()
}

func (s *SP2p) UpdateNode(rawUrl string) error {
	n, err := ParseNode(rawUrl)
	if err != nil {
		return err
	}
	s.tab.UpdateNode(n)
	return nil
}
func (s *SP2p) DeleteNode(id string) {
	s.tab.DeleteNode(common.StringToHash(id))
}

func (s *SP2p) AddNode(rawUrl string) error {
	n, err := ParseNode(rawUrl)
	if err != nil {
		return err
	}
	s.tab.AddNode(n)
	return nil
}

func (s *SP2p) FindMinDisNodes(target string, n int) (nodes []string) {
	for _, n := range s.tab.FindMinDisNodes(common.StringToHash(target), n) {
		nodes = append(nodes, n.String())
	}
	return
}

func (s *SP2p) FindRandomNodes(n int) (nodes []string) {
	for _, n := range s.tab.FindRandomNodes(n) {
		nodes = append(nodes, n.String())
	}
	return
}

func (s *SP2p) FindNodeWithTargetBySelf(d string) (nodes []string) {
	for _, n := range s.tab.FindNodeWithTargetBySelf(common.StringToHash(d)) {
		nodes = append(nodes, n.String())
	}
	return
}

func (s *SP2p) FindNodeWithTarget(target string, measure string) (nodes []string) {
	for _, n := range s.tab.FindNodeWithTarget(common.StringToHash(target), common.StringToHash(measure)) {
		nodes = append(nodes, n.String())
	}
	return
}

func (s *SP2p) PingN() {
	go s.pingN()
}

func (s *SP2p) PingNode(taddr string) {
	go s.pingNode(taddr)
}

func (s *SP2p) FindN() {
	go s.findN()
}

func (s *SP2p) FindNode(taddr string, n int) {
	go s.findNode(taddr, n)
}
