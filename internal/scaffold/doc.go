// Package scaffold defines the core domain types and validation logic
// for project scaffolding operations.
//
// This package is dependency-free and contains pure business rules including:
//   - CreateRequest and CreateResult types for scaffolding operations
//   - GitMode enumeration for Git initialization behavior
//   - TemplateVars for template variable substitution
//   - Validation functions for project names and parameters
//
// The scaffold package represents the domain layer in a hexagonal architecture,
// with no dependencies on external systems or frameworks.
package scaffold
