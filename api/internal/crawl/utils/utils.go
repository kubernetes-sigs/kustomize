package utils

type SeenMap map[string]string

func (seen SeenMap) Seen(item string) bool {
	_, ok := seen[item]
	return ok
}

func (seen SeenMap) Set(k, v string) {
	seen[k] = v
}

// The caller should make sure that key is in the map.
func (seen SeenMap) Value(k string) string {
	return seen[k]
}

func NewSeenMap() SeenMap {
	return make(map[string]string)
}
