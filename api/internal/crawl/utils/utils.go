package utils

type SeenMap map[string]struct{}

func (seen SeenMap) Seen(item string) bool {
	_, ok := seen[item]
	return ok
}

func (seen SeenMap) Add(item string) {
	seen[item] = struct{}{}
}

func NewSeenMap() SeenMap {
	return make(map[string]struct{})
}
