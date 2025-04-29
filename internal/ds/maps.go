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
