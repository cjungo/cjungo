package cjungo

import "go.uber.org/dig"

type DiContainer interface {
	ProvideController(controllers []any) error
	Decorate(decorator interface{}, opts ...dig.DecorateOption) error
	Invoke(function interface{}, opts ...dig.InvokeOption) error
	Provide(constructor interface{}, opts ...dig.ProvideOption) error
	Scope(name string, opts ...dig.ScopeOption) *dig.Scope
	String() string
}

type DiSimpleContainer struct {
	*dig.Container
}

func (container *DiSimpleContainer) ProvideController(controllers []any) error {
	for _, c := range controllers {
		if err := container.Provide(c); err != nil {
			return err
		}
	}
	return nil
}
