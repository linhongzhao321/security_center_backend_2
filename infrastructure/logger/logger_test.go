package logger

import (
	"context"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"testing"

	"go.uber.org/zap"
)

func ExampleWrite() {
	Write(context.Background(), zap.InfoLevel,
		`this is example`,
		zap.String(`string-field`, `string-field-value`),
		zap.Int(`int-field`, 1),
		zap.Any(`any-field`, `any-field-value`),
		zap.Ints(`ints-field`,
			[]int{3, 4, 1}),
	)
	// Output:
}

func TestWrite(t *testing.T) {
	var balance uint64 = math.MaxUint64
	fmt.Printf("当前余额: %d\n", balance)
	newBalance := balance + 1
	fmt.Printf("进账后余额: %d (资产清零事故)\n", newBalance)
}

type CustomStruct struct {
	A uint64
	_ [56]byte // Padding: 填充 56 字节
	B uint64
}

func BenchmarkPadding(b *testing.B) {
	custom := &CustomStruct{}
	b.Run("padding", func(b *testing.B) {
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			for i := 0; i < b.N; i++ {
				atomic.AddUint64(&custom.A, 1)
			}
			wg.Done()
		}()
		go func() {
			for i := 0; i < b.N; i++ {
				atomic.AddUint64(&custom.B, 1)
			}
			wg.Done()
		}()
		wg.Wait()
	})
}
