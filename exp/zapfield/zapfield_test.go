package zapfield

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type (
	MyKey    string
	MyValue  string
	MyValues []MyValue
)

func TestFieldConstructors(t *testing.T) {
	var (
		key    = MyKey("test key")
		value  = MyValue("test value")
		values = []MyValue{
			MyValue("test value 1"),
			MyValue("test value 2"),
		}
	)

	tests := []struct {
		name   string
		field  zap.Field
		expect zap.Field
	}{
		{"Str", zap.Field{Type: zapcore.StringType, Key: "test key", String: "test value"}, Str(key, value)},
		{"Strs", zap.Array("test key", stringArray[MyValue]{"test value 1", "test value 2"}), Strs(key, values)},
	}

	for _, tt := range tests {
		if !assert.Equal(t, tt.expect, tt.field, "Unexpected output from convenience field constructor %s.", tt.name) {
			t.Logf("type expected: %T\nGot: %T", tt.expect.Interface, tt.field.Interface)
		}
		assertCanBeReused(t, tt.field)
	}

}

func assertCanBeReused(t testing.TB, field zap.Field) {
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		enc := zapcore.NewMapObjectEncoder()

		// Ensure using the field in multiple encoders in separate goroutines
		// does not cause any races or panics.
		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NotPanics(t, func() {
				field.AddTo(enc)
			}, "Reusing a field should not cause issues")
		}()
	}

	wg.Wait()
}
