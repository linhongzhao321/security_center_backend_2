package project

import (
	"context"
	"sync"

	"github.com/panjf2000/ants/v2"
	"go.uber.org/zap"

	"infrastructure/logger"
	"internal/core"
)

type Project struct {
	name       string
	components []core.Component
	monitors   map[string]core.Monitor
	isSync     bool
}

func (p *Project) Run(ctx context.Context) []core.MonitorResult {
	if len(p.monitors) == 0 {
		logger.Write(ctx, zap.WarnLevel, `project.monitors is empty`, zap.String(`project`, p.name))
		return nil
	}

	var results []core.MonitorResult
	var mutexMonitorResults sync.Mutex

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(p.monitors))

	for monitorName := range p.monitors {
		runner := func(monitor core.Monitor) func() {
			return func() {
				defer func() {
					if err := recover(); err != nil {
						logger.Write(ctx, zap.ErrorLevel, `monitor error`, zap.Any(`error`, err))
					}
				}()
				logger.Write(ctx, zap.InfoLevel, `monitor start`, zap.String(`name`, monitor.Name()))
				// run monitor
				monitorResults := monitor.Run(ctx, p)

				// append result
				mutexMonitorResults.Lock()
				results = append(results, monitorResults...)
				mutexMonitorResults.Unlock()

				waitGroup.Done()
				logger.Write(ctx, zap.InfoLevel, `monitor end`, zap.String(`name`, monitor.Name()))
			}
		}
		if p.isSync {
			runner(p.monitors[monitorName])()
		} else {
			err := ants.Submit(runner(p.monitors[monitorName]))
			if err != nil {
				logger.Write(ctx, zap.ErrorLevel, `monitor running error.`,
					zap.String(`project`, p.name),
					zap.String(`monitor`, monitorName),
				)
			}
		}
	}

	waitGroup.Wait()

	return results
}

func (p *Project) AppendComponents(components ...core.Component) {
	p.components = append(p.components, components...)
}

func (p *Project) Name() string {
	return p.name
}

func (p *Project) Components() []core.Component {
	return p.components
}

type Option func(project *Project) error

func NewProject(name string, options ...Option) (core.Project, error) {

	project := &Project{
		name:     name,
		monitors: map[string]core.Monitor{},
	}

	for _, option := range options {
		err := option(project)
		if err != nil {
			return nil, err
		}
	}
	return project, nil
}

func MustProject(name string, options ...Option) core.Project {
	project, err := NewProject(name, options...)
	if err != nil {
		panic(err)
	}
	return project
}

// SyncRunMonitor 同步模式运行所有 monitor
// 默认为异步模式
func SyncRunMonitor() Option {
	return func(project *Project) error {
		project.isSync = true
		return nil
	}
}

func MustComponent(name string, options ...ComponentOption) Option {
	return func(project *Project) error {
		component, err := NewComponent(name, options...)
		if err != nil {
			panic(err)
		}
		project.components = append(project.components, component)
		return nil
	}
}

func AppendMonitors(monitors ...core.Monitor) Option {
	return func(project *Project) error {
		for _, monitor := range monitors {
			if _, isExist := project.monitors[monitor.Name()]; isExist {
				logger.Write(context.Background(), zap.FatalLevel,
					`duplicated monitor`, zap.String(`name`, monitor.Name()))
			}
			project.monitors[monitor.Name()] = monitor
		}
		return nil
	}
}
