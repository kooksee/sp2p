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
		hm = &HandleManager{hmap: make(map[string]IHandler)}
	})
	return hm
}

type HandleManager struct {
	hmap map[string]IHandler
}

func (h *HandleManager) Registry(name string, handler IHandler) error {
	if h.Contain(name) {
		return errors.New(fmt.Sprintf("%s existed", name))
	}
	h.hmap[name] = handler
	return nil
}

func (h *HandleManager) Contain(name string) bool {
	_, ok := h.hmap[name]
	return ok
}

func (h *HandleManager) GetHandler(name string) IHandler {
	return h.hmap[name]
}
