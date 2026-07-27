// Package Drivers contains concrete StorageDriver implementations.
package Drivers

// This file re-exports the memory driver from the parent Storage package
// to keep the Drivers/ folder as the canonical location for driver code
// while the manager itself lives one level up.
//
// The in-memory driver is defined in framework/Storage/memory.go and is
// embedded here via the Storage package import so framework users can
// reference it from Drivers/ if they prefer that path.
//
// Actual concrete implementations (MinIO, S3, GDrive) go in separate files
// in this package once the external dependencies are added to go.mod.
