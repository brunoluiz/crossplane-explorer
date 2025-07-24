package ds

// WalkMap function that takes a map and a callback function.
func WalkMap(m map[string]interface{}, callback func(key string, value interface{}) (interface{}, bool)) {
	walkMapRecursive(m, callback, "")
}

// Helper function to walk through the map recursively.
func walkMapRecursive(
	m map[string]interface{},
	callback func(key string, value interface{}) (interface{}, bool),
	parentKey string,
) {
	for key, value := range m {
		fullKey := key
		if parentKey != "" {
			fullKey = parentKey + "." + key
		}

		// Apply the callback function
		newValue, keep := callback(fullKey, value)
		if !keep {
			delete(m, key)
			continue
		}

		// If the value is another map, walk through it recursively
		if nestedMap, ok := newValue.(map[string]interface{}); ok {
			walkMapRecursive(nestedMap, callback, fullKey)
		} else {
			m[key] = newValue
		}
	}
}

func GetPath[V any](m map[string]any, path ...string) (V, bool) {
	var zero V
	curr := any(m)
	for i, key := range path {
		mm, ok := curr.(map[string]any)
		if !ok {
			return zero, false
		}
		v, exists := mm[key]
		if !exists {
			return zero, false
		}
		curr = v
		// If this is the last key, try to cast to V
		if i == len(path)-1 {
			val, ok := curr.(V)
			if ok {
				return val, true
			}
			return zero, false
		}
	}
	return zero, false
}
