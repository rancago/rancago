// Package Container provides an IoC service container for the Rancago Framework.
// It supports Singleton, Transient, and Instance bindings with alias support.
package Container

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// BindingType represents the lifecycle of a container binding.
type BindingType int

const (
	Singleton BindingType = iota // One instance for the lifetime of the container.
	Transient                    // New instance on every resolve.
	Instance                     // Pre-existing, pre-built instance.
)

type binding struct {
	resolver    func(*Container) (interface{}, error)
	bindingType BindingType
	resolved    bool
	instance    interface{}
}

// Container is a thread-safe IoC service container.
type Container struct {
	mu       sync.RWMutex
	bindings map[string]*binding
	aliases  map[string]string
}

// NewContainer creates a new empty Container.
func NewContainer() *Container {
	return &Container{
		bindings: make(map[string]*binding),
		aliases:  make(map[string]string),
	}
}

// Bind registers a transient (new instance per resolve) binding.
func (c *Container) Bind(abstract string, resolver func(*Container) (interface{}, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[abstract] = &binding{resolver: resolver, bindingType: Transient}
}

// Singleton registers a singleton (one instance) binding.
func (c *Container) Singleton(abstract string, resolver func(*Container) (interface{}, error)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[abstract] = &binding{resolver: resolver, bindingType: Singleton}
}

// Instance registers a pre-existing instance.
func (c *Container) Instance(abstract string, instance interface{}) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.bindings[abstract] = &binding{bindingType: Instance, instance: instance, resolved: true}
}

// Alias maps an alias name to a concrete binding key.
func (c *Container) Alias(abstract, alias string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.aliases[alias] = abstract
}

// Resolve resolves a binding by its abstract key or alias.
func (c *Container) Resolve(abstract string) (interface{}, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.resolve(abstract)
}

func (c *Container) resolve(abstract string) (interface{}, error) {
	key := abstract
	if resolved, ok := c.aliases[abstract]; ok {
		key = resolved
	}
	b, ok := c.bindings[key]
	if !ok {
		return nil, errors.New("container: no binding found for: " + key)
	}
	switch b.bindingType {
	case Instance:
		return b.instance, nil
	case Singleton:
		if b.resolved {
			return b.instance, nil
		}
		inst, err := b.resolver(c)
		if err != nil {
			return nil, err
		}
		b.instance = inst
		b.resolved = true
		return inst, nil
	case Transient:
		return b.resolver(c)
	default:
		return nil, errors.New("container: unknown binding type for: " + key)
	}
}

// MustResolve resolves a binding and panics if resolution fails.
func (c *Container) MustResolve(abstract string) interface{} {
	inst, err := c.Resolve(abstract)
	if err != nil {
		panic(fmt.Sprintf("container: MustResolve failed for %q: %v", abstract, err))
	}
	return inst
}

// Has reports whether a binding exists for the given key.
func (c *Container) Has(abstract string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	key := abstract
	if resolved, ok := c.aliases[abstract]; ok {
		key = resolved
	}
	_, ok := c.bindings[key]
	return ok
}

// Call calls fn, auto-resolving its parameters from the container via reflection.
func (c *Container) Call(fn interface{}) ([]reflect.Value, error) {
	fnType := reflect.TypeOf(fn)
	if fnType.Kind() != reflect.Func {
		return nil, errors.New("container: Call requires a function argument")
	}
	numIn := fnType.NumIn()
	args := make([]reflect.Value, numIn)
	c.mu.Lock()
	defer c.mu.Unlock()
	for i := 0; i < numIn; i++ {
		paramType := fnType.In(i)
		resolved, err := c.resolve(paramType.String())
		if err != nil {
			return nil, fmt.Errorf("container: failed to resolve parameter %d (%s): %w", i, paramType.String(), err)
		}
		args[i] = reflect.ValueOf(resolved)
	}
	return reflect.ValueOf(fn).Call(args), nil
}
