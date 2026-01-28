package observer

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestObserver_Mark(t *testing.T) {
	i := 10
	mutex := sync.RWMutex{}
	spec := `*/2 * * * * *`
	action := func(ctx context.Context) {
		mutex.RLock()
		if i != 0 {
			t.Errorf(`i should be 0, actual %d`, i)
		} else {
			t.Log(`i = 0`)
		}
		mutex.RUnlock()
	}
	recv := func(ctx context.Context, sig interface{}) {
		mutex.Lock()
		i += sig.(int)
		mutex.Unlock()
	}
	observer := New(`test`, spec, action, recv)
	// 未初始化，调用 Observer.Mark() 将报错
	err := observer.Mark(-1)
	if err == nil {
		t.Fatalf(`Observer.Mark() should return error on not running Observer.Run()`)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		time.Sleep(time.Second)
		err := observer.Run(ctx)
		if err != nil {
			t.Error(err)
		}
	}()
	time.Sleep(2 * time.Second)
	for cnt := i; cnt > 0; cnt -= 1 {
		err = observer.Mark(-1)
		if err != nil {
			t.Error(err)
		}
	}
	time.Sleep(10 * time.Second)
	cancel()
}
