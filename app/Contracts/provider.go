// Package Contracts holds all interface definitions for the Rancago Framework.
// Every cross-module dependency MUST go through these contracts (DIP compliant).
package Contracts

import "github.com/rancago/framework/framework/Container"

// ServiceProvider is the interface every module's provider must implement.
// Register is called first for all providers, Boot second.
type ServiceProvider interface {
	// Register binds services into the container. No heavy initialization here.
	Register(c *Container.Container) error
	// Boot runs after all providers have been registered. Wire-up drivers, seed data, etc.
	Boot(c *Container.Container) error
}
