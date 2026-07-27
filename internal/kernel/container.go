package kernel

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

type BindingType int

const (
	Singleton BindingType = iota
	Transient
	Instance
)

type binding struct {
	concrete    interface{}
	resolver    func(*Container) (interface{}, error)
	bindingType BindingType
	resolved    bool
	instance    interface{}
}

type Container struct {
	mu        sync.RWMutex
	bindings  map[string]*binding
	aliases   map[string]string
	instances map[string]interface{}
}

func NewContainer() *Container {
	return &Container{
		bindings:  make(map[string]*binding),
		aliases:   make(map[string]string),
		instances: make(map[string]interface{}),
	}
}

func (c *Container) Bind(abstract string, resolver func(*Container) (interface{}, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[abstract] = &binding{
		resolver:    resolver,
		bindingType: Transient,
	}
}

func (c *Container) Singleton(abstract string, resolver func(*Container) (interface{}, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[abstract] = &binding{
		resolver:    resolver,
		bindingType: Singleton,
	}
}

func (c *Container) Instance(abstract string, instance interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[abstract] = &binding{
		bindingType: Instance,
		instance:    instance,
		resolved:    true,
	}
	c.instances[abstract] = instance
}

func (c *Container) Alias(abstract string, alias string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aliases[alias] = abstract
}

func (c *Container) getAbstract(abstract string) string {
	if resolved, ok := c.aliases[abstract]; ok {
		return resolved
	}
	return abstract
}

func (c *Container) Resolve(abstract string) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.resolve(abstract)
}

func (c *Container) resolve(abstract string) (interface{}, error) {
	key := c.getAbstract(abstract)
	b, ok := c.bindings[key]
	if !ok {
		return nil, errors.New("no binding found for: " + key)
	}
	switch b.bindingType {
	case Instance:
		return b.instance, nil
	case Singleton:
		if b.resolved {
			return b.instance, nil
		}
		instance, err := b.resolver(c)
		if err != nil {
			return nil, err
		}
		b.instance = instance
		b.resolved = true
		c.instances[key] = instance
		return instance, nil
	case Transient:
		return b.resolver(c)
	default:
		return nil, errors.New("unknown binding type for: " + key)
	}
}

func (c *Container) MustResolve(abstract string) interface{} {
	instance, err := c.Resolve(abstract)
	if err != nil {
		panic(err)
	}
	return instance
}

func (c *Container) Has(abstract string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	key := c.getAbstract(abstract)
	_, ok := c.bindings[key]
	return ok
}

func (c *Container) Call(fn interface{}) ([]reflect.Value, error) {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return nil, errors.New("Call requires a function argument")
	}
	numIn := fnType.NumIn()
	args := make([]reflect.Value, numIn)
	for i := 0; i < numIn; i++ {
		paramType := fnType.In(i)
		resolved, err := c.resolve(paramType.String())
		if err != nil {
			return nil, fmt.Errorf("failed to resolve parameter %d (%s): %w", i, paramType.String(), err)
		}
		args[i] = reflect.ValueOf(resolved)
	}
	return reflect.ValueOf(fn).Call(args), nil
}
