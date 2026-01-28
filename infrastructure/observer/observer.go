package observer

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"

	error2 "infrastructure/error"
)

type Action func(ctx context.Context)
type Receive func(ctx context.Context, sig interface{})

const chanSize = 10

// Observer ch收到任何数据时调用 recv，并按照 ttl 周期定期执行 callback
type Observer struct {
	// 当前观察者名称
	name string
	// crontab 描述
	spec string
	// 每间隔 ttl 执行一次 action
	action   Action
	ch       chan interface{}
	isClosed bool
	// 收到 ch 时，执行 recv
	recv Receive
}

func New(name, spec string, action Action, recv Receive) *Observer {
	return &Observer{
		name:     name,
		spec:     spec,
		action:   action,
		recv:     recv,
		isClosed: true,
	}
}

// Run 可以传递一个 cancelContext，主动结束任务
func (observer *Observer) Run(ctx context.Context) error {
	receiveCtx, cancelRecv := context.WithCancel(ctx)
	observer.ch = make(chan interface{}, chanSize)
	observer.isClosed = false
	defer func() {
		cancelRecv()
		observer.isClosed = true
		close(observer.ch)
	}()

	// 优先接收信号，所有信号接收完毕才会判断 receiveCtx 是否关闭
	go func() {
		for {
			select {
			case content := <-observer.ch:
				go func() {
					recvCtx, cancelFunc := context.WithCancel(ctx)
					observer.recv(recvCtx, content)
					cancelFunc()
				}()
			case <-receiveCtx.Done():
				return
			}
		}
	}()

	task := cron.New(cron.WithSeconds())
	_, err := task.AddFunc(observer.spec, func() {
		actionCtx, cancelFunc := context.WithCancel(ctx)
		observer.action(actionCtx)
		cancelFunc()
	})
	if err != nil {
		return err
	}
	task.Start()
	if <-ctx.Done(); true {
		task.Stop()
	}
	return nil
}

func (observer *Observer) Mark(content interface{}) error {
	if observer.isClosed {
		return &error2.ChanClosed{
			Deposit: fmt.Sprintf(`Observer[Name:%s].Mark()`, observer.name),
			Content: fmt.Sprintf(`%T`, content),
		}
	}
	observer.ch <- content
	return nil
}
