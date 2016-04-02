package pulse

var registeredObjects = make(map[string]interface{})

func cgoregister(key string, obj interface{}) {
	registeredObjects[key] = obj
}

func cgoget(key string) (interface{}, bool) {
	v, ok := registeredObjects[key]
	return v, ok
}

func cgounregister(key string) {
	delete(registeredObjects, key)
}
