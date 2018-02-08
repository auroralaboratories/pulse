package pulse

import (
	"sync"
	"unsafe"
)

var registeredObjects sync.Map

type Registerable interface {
	Userdata() unsafe.Pointer
}

func cgoregister(key string, obj interface{}) {
	registeredObjects.Store(key, obj)
}

func cgoget(key string) interface{} {
	value, _ := registeredObjects.Load(key)
	return value
}

func cgounregister(key string) {
	registeredObjects.Delete(key)
}
