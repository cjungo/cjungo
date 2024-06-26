package cjungo

import "go.uber.org/dig"

type DiContainer interface {
	Provides(constructors ...any) error
	Decorate(decorator any, opts ...dig.DecorateOption) error
	Invoke(function any, opts ...dig.InvokeOption) error
	Provide(constructor any, opts ...dig.ProvideOption) error
	Scope(name string, opts ...dig.ScopeOption) *dig.Scope
	String() string
}

type DiSimpleContainer struct {
	*dig.Container
}

func (container *DiSimpleContainer) Provides(controllers ...any) error {
	for _, c := range controllers {
		if err := container.Provide(c); err != nil {
			return err
		}
	}
	return nil
}
