package tests

import (
	"github.com/modern-go/reflect2"
	"testing"

	"github.com/modern-go/test/must"
	"github.com/modern-go/test"
	"context"
)

func testOp(f func(api reflect2.API) interface{}) func(t *testing.T) {
	return test.Case(func(ctx context.Context) {
		unsafeResult := f(reflect2.ConfigUnsafe)
		safeResult := f(reflect2.ConfigSafe)
		must.Equal(safeResult, unsafeResult)
	})
}
