package yaml

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsYaml1_1NonString(t *testing.T) {
	type testCase struct {
		val      string
		expected bool
	}

	testCases := []testCase{
		{val: "hello world", expected: false},
		{val: "2.0", expected: true},
	}

	for k := range valueToTagMap {
		testCases = append(testCases, testCase{val: k, expected: true})
	}

	for _, test := range testCases {
		assert.Equal(t, test.expected,
			IsYaml1_1NonString(&Node{Kind: ScalarNode, Value: test.val}), test.val)
	}
}

// valueToTagMap is a map of values interpreted as non-strings in yaml 1.1 when left
// unquoted.
// To keep compatibility with the yaml parser used by Kubernetes (yaml 1.1) make sure the values
// which are treated as non-string values are kept as non-string values.
// https://github.com/go-yaml/yaml/blob/v2/resolve.go
var valueToTagMap = func() map[string]string {
	val := map[string]string{}

	// https://yaml.org/type/null.html
	values := []string{"", "~", "null", "Null", "NULL"}
	for i := range values {
		val[values[i]] = "!!null"
	}

	// https://yaml.org/type/bool.html
	values = []string{
		"y", "Y", "yes", "Yes", "YES", "true", "True", "TRUE", "on", "On", "ON", "n", "N", "no",
		"No", "NO", "false", "False", "FALSE", "off", "Off", "OFF"}
	for i := range values {
		val[values[i]] = "!!bool"
	}

	// https://yaml.org/type/float.html
	values = []string{
		".nan", ".NaN", ".NAN", ".inf", ".Inf", ".INF",
		"+.inf", "+.Inf", "+.INF", "-.inf", "-.Inf", "-.INF"}
	for i := range values {
		val[values[i]] = "!!float"
	}

	return val
}()
