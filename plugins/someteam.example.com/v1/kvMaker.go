// +build plugin

package main
var database = map[string]string{
  "TREE":      "oak",
  "ROCKET":    "Saturn V",
  "FRUIT":     "apple",
  "VEGETABLE": "carrot",
  "SIMPSON":   "homer",
}

type plugin struct{}
var KVSource plugin
func (p plugin) Get(
    root string, args []string) (map[string]string, error) {
  r := make(map[string]string)
  for _, k := range args {
    v, ok := database[k]
    if ok {
      r[k] = v
    }
  }
  return r, nil
}
