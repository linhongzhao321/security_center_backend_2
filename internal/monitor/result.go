package monitor

import (
	"errors"
	"fmt"
)

type Result struct {
	err                error
	monitorName        string
	projectName        string
	componentName      string
	releaseChannelName string
	extensions         map[string]string
}

func (r *Result) MonitorName() string {
	return r.monitorName
}

func (r *Result) ProjectName() string {
	return r.projectName
}

func (r *Result) ComponentName() string {
	return r.componentName
}

func (r *Result) ReleaseChannelName() string {
	return r.releaseChannelName
}

func (r *Result) Extensions() map[string]string {
	return r.extensions
}

func (r *Result) String() string {
	errMsg := `error is nil`
	if r.err != nil {
		errMsg = fmt.Sprintf(`error:%+v`, r.err)
	}
	return fmt.Sprintf(`[%s.%s.%s.%s] %s extensions:%+v`,
		r.monitorName, r.projectName, r.componentName, r.releaseChannelName, errMsg, r.extensions)
}

func (r *Result) IsFail() bool {
	return r.err != nil
}

func (r *Result) Error() error {
	return r.err
}

func NewResult(monitor, project, component, releaseChannel string, options ...ResultOption) *Result {
	result := &Result{
		monitorName:        monitor,
		projectName:        project,
		componentName:      component,
		releaseChannelName: releaseChannel,
		extensions:         map[string]string{},
	}
	for _, option := range options {
		option(result)
	}
	return result
}

type ResultOption func(result *Result)

func WithError(err error) ResultOption {
	return func(result *Result) {
		result.err = err
	}
}

func Error(msg string) ResultOption {
	return func(result *Result) {
		result.err = errors.New(msg)
	}
}

func Extension(key, val string) ResultOption {
	return func(result *Result) {
		result.extensions[key] = val
	}
}
