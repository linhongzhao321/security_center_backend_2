package pagerduty

import (
	"context"
	"fmt"
	"time"

	"github.com/PagerDuty/go-pagerduty"
)

type Severity string

var (
	Info     Severity = `info`
	Warning  Severity = `warning`
	Error    Severity = `error`
	Critical Severity = `critical`
)

type Action string

var (
	Acknowledge Action = `acknowledge`
	Resolve     Action = `resolve`
	Trigger     Action = `trigger`
)

var eventClient *pagerduty.Client

func init() {
	// event api 不需要 token，使用 request param 中的 routing key 验证身份
	eventClient = pagerduty.NewClient(``)
}

func SendEvent(ctx context.Context, action Action, routingKey, summary, source string, severity Severity, setters ...EventParamSetters) error {
	event := &pagerduty.V2Event{
		RoutingKey: routingKey,
		Action:     string(action),
		Payload: &pagerduty.V2Payload{
			Summary:   summary,
			Source:    source,
			Timestamp: time.Now().Format(time.RFC3339),
			Severity:  string(severity),
			Details:   map[string]interface{}{},
		},
		Links:  []interface{}{},
		Images: []interface{}{},
	}
	for _, setter := range setters {
		setter(event)
	}
	resp, err := eventClient.ManageEventWithContext(ctx, event)
	if err != nil {
		return err
	}
	if resp.Status != `success` {
		return fmt.Errorf(`%s, errors: %+v`, resp.Message, resp.Errors)
	}
	return nil
}

func SendChangeEvent(ctx context.Context, routingKey, summary, source string, setters ...ChangeEventParamSetters) error {
	event := pagerduty.ChangeEvent{
		RoutingKey: routingKey,
		Payload: pagerduty.ChangeEventPayload{
			Summary:   summary,
			Source:    source,
			Timestamp: time.Now().Format(time.RFC3339),
		},
	}
	for _, setter := range setters {
		setter(event)
	}
	resp, err := eventClient.CreateChangeEventWithContext(ctx, event)
	if err != nil {
		return err
	}
	if resp.Status != `success` {
		return fmt.Errorf(`%s, errors: %+v`, resp.Message, resp.Errors)
	}
	return nil
}

type EventParamSetters func(event *pagerduty.V2Event)
type ChangeEventParamSetters func(event pagerduty.ChangeEvent)

func Component(component string) EventParamSetters {
	return func(event *pagerduty.V2Event) {
		event.Payload.Component = component
	}
}

func Group(group string) EventParamSetters {
	return func(event *pagerduty.V2Event) {
		event.Payload.Group = group
	}
}

func Class(class string) EventParamSetters {
	return func(event *pagerduty.V2Event) {
		event.Payload.Class = class
	}
}

func AppendCustomDetails(key string, val interface{}) EventParamSetters {
	return func(event *pagerduty.V2Event) {
		event.Payload.Details.(map[string]interface{})[key] = val
	}
}

type Link struct {
	Href string `json:"href"`
	Text string `json:"text"`
}

func Links(links ...Link) EventParamSetters {
	return func(event *pagerduty.V2Event) {
		for _, link := range links {
			event.Links = append(event.Links, link)
		}
	}
}

type Image struct {
	Src  string `json:"src"`
	Href string `json:"href"`
	Alt  string `json:"alt"`
}

func Images(images ...Image) EventParamSetters {
	return func(event *pagerduty.V2Event) {
		for _, image := range images {
			event.Images = append(event.Images, image)
		}
	}
}
