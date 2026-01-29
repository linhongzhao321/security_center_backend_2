package monitor

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"go.uber.org/zap"

	"infrastructure/logger"
	"infrastructure/observer"
	"infrastructure/pagerduty"
	"infrastructure/slack"
	"internal/core"
)

func SlackNotifier(webhook string, recipientIDs ...string) core.MonitorNotifier {
	return func(ctx context.Context, result core.MonitorResult) error {
		if !result.IsFail() {
			return nil
		}

		extensions := ``
		for k, v := range result.Extensions() {
			extensions += fmt.Sprintf("\t%s: %s\n", k, v)
		}
		content := fmt.Sprintf("[%s] %s.%s.%s\nerror: %s\n%s",
			result.MonitorName(), result.ProjectName(), result.ComponentName(), result.ReleaseChannelName(),
			result.Error(), extensions)
		return slack.Notify(ctx, webhook, content, recipientIDs...)
	}
}

func PagerDutyNotifier(routingKey string) core.MonitorNotifier {
	return func(ctx context.Context, result core.MonitorResult) error {
		if !result.IsFail() {
			return nil
		}

		summary := result.String()
		source := fmt.Sprintf(`%s-%s`, result.ProjectName(), result.ReleaseChannelName())
		group := fmt.Sprintf(`%s-%s`, result.ProjectName(), result.ComponentName())
		params := []pagerduty.EventParamSetters{
			pagerduty.Component(result.ComponentName()),
			pagerduty.Group(group),
			pagerduty.Class(result.MonitorName()),
		}
		for k, v := range result.Extensions() {
			params = append(params, pagerduty.AppendCustomDetails(k, v))
		}
		severity := pagerduty.Warning
		return pagerduty.SendEvent(ctx, pagerduty.Trigger, routingKey, summary, source, severity, params...)
	}
}

const observerName = `PeriodicReportSlack`

func PeriodicReportSlack(webhook string, spec string) core.MonitorNotifier {
	counters := &sync.Map{}
	action := func(ctx context.Context) {
		var content []string
		counters.Range(func(key, value any) bool {
			content = append(content, fmt.Sprintf("%s\t%d", key.(string), *value.(*int64)))
			return true
		})
		counters.Range(func(key, _ any) bool {
			counters.Delete(key)
			return true
		})
		sort.Strings(content)
		sContent := fmt.Sprintf("距上一次报告后，新增执行检查次数：\n%s", strings.Join(content, "\n"))
		_ = slack.Notify(ctx, webhook, sContent)
		logger.Write(context.Background(), zap.InfoLevel, `PeriodicReportSlack`,
			zap.String(`content`, sContent))
	}
	recv := func(ctx context.Context, sig interface{}) {
		result := sig.(core.MonitorResult)
		key := fmt.Sprintf(`%s-%s-%s`,
			result.ProjectName(), result.ComponentName(), result.ReleaseChannelName())
		updateCounter(counters, key)
	}
	ob := observer.New(observerName, spec, action, recv)
	go func() {
		err := ob.Run(context.Background())
		if err != nil {
			logger.Write(context.Background(), zap.ErrorLevel, `observer.Run() return error`,
				zap.Error(err))
		}
	}()
	return func(ctx context.Context, result core.MonitorResult) error {
		err := ob.Mark(result)
		return err
	}
}

func updateCounter(counters *sync.Map, key string) {
	val, _ := counters.LoadOrStore(key, new(int64))
	ptr := val.(*int64)
	atomic.AddInt64(ptr, 1)
}
