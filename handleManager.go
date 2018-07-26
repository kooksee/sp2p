package sp2p

import (
	"sync"
	"reflect"
)

var (
	hmOnce sync.Once
	hm     *handleManager
)

func RegistryHandlers(handlers ... interface{}) {
	getHManager().registry(handlers...)
}

func getHManager() *handleManager {
	hmOnce.Do(func() {
		hm = &handleManager{hmap: make(map[byte]reflect.Type)}
	})
	return hm
}

type handleManager struct {
	hmap map[byte]reflect.Type
}

func (h *handleManager) registry(handlers ... interface{}) {
	for _, handler := range handlers {

		h1 := reflect.TypeOf(handler)
		h3 := reflect.New(h1).Interface().(IMessage)

		name := h3.T()
		if h.contain(name) {
			getLog().Error("handle exist", "type", name, "desc", h3.String())
			panic("")
		}
		h.hmap[name] = h1
	}
}

func (h *handleManager) contain(name byte) bool {
	_, ok := h.hmap[name]
	return ok
}

func (h *handleManager) getHandler(name byte) IMessage {
	h1 := h.hmap[name]
	h2 := reflect.New(h1)
	return h2.Interface().(IMessage)
}
