package bulk

// MapInvert swaps keys and values in a map. If duplicate values exist, the resulting key is nondeterministic.
func MapInvert[K comparable, V comparable](m map[K]V) map[V]K {
	result := make(map[V]K, len(m))
	MapInvertInto(result, m)
	return result
}

// MapInvertInto swaps keys and values from the source map into the destination map.
// If duplicate values exist, the resulting key is nondeterministic.
func MapInvertInto[K comparable, V comparable](dest map[V]K, m map[K]V) {
	for key, value := range m {
		dest[value] = key
	}
}

// MapKeysSlice returns a slice containing all keys from the map.
// The result is pre-allocated to the exact size needed, avoiding reallocations and minimizing memory usage.
// Order of keys is nondeterministic.
func MapKeysSlice[K comparable, V any](m map[K]V) []K {
	if len(m) == 0 {
		return nil
	}
	result := make([]K, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}

// MapValuesSlice returns a slice containing all values from the map.
// The result is pre-allocated to the exact size needed, avoiding reallocations and minimizing memory usage.
// Order of values is nondeterministic.
func MapValuesSlice[K comparable, V any](maps ...map[K]V) []V {
	size := mapTotalSize(maps)
	if size == 0 {
		return nil
	}
	result := make([]V, 0, size)
	for _, m := range maps {
		for _, v := range m {
			result = append(result, v)
		}
	}
	return result
}

func mapTotalSize[K comparable, V any](maps []map[K]V) int {
	var size int
	for _, m := range maps {
		size += len(m)
	}
	return size
}
