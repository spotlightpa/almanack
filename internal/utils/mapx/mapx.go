package mapx

import (
	"maps"
	"slices"
)

// Flatten converts a map[string]string to a []string of alternating key-value
// pairs sorted by key.
func Flatten(m map[string]string) []string {
	result := make([]string, 0, len(m)*2)
	for _, k := range slices.Sorted(maps.Keys(m)) {
		result = append(result, k, m[k])
	}
	return result
}

// FlattenMulti converts a map[string][]string to a []string of alternating
// key-value pairs sorted by key, repeating the key for each value.
func FlattenMulti(m map[string][]string) []string {
	size := 0
	for _, vs := range m {
		size += len(vs)
	}
	result := make([]string, 0, size*2)
	for _, k := range slices.Sorted(maps.Keys(m)) {
		for _, v := range m[k] {
			result = append(result, k, v)
		}
	}
	return result
}
