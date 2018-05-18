package sp2p

import (
	"errors"
	"fmt"
	"sync"
)

var (
	hmOnce sync.Once
	hm     *HandleManager
)

func GetHManager() *HandleManager {
	hmOnce.Do(func() {
		hm = &HandleManager{hmap: make(map[byte]IMessage)}
	})
	return hm
}

type HandleManager struct {
	hmap map[byte]IMessage
}

func (h *HandleManager) Registry(name byte, handler IMessage) error {
	if h.Contain(name) {
		return errors.New(fmt.Sprintf("%s existed", name))
	}
	h.hmap[name] = handler
	return nil
}

func (h *HandleManager) Contain(name byte) bool {
	_, ok := h.hmap[name]
	return ok
}

func (h *HandleManager) GetHandler(name byte) IMessage {
	return h.hmap[name]
}
