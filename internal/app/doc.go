// Package app contains the application layer services that orchestrate
// business workflows and use cases.
//
// This layer coordinates between the domain layer (scaffold) and adapters layer,
// implementing the core application logic for:
//   - Project creation (create package)
//   - Template catalog operations (catalog package)
//   - Version information (version package)
//
// Services in this layer depend on both domain types and adapter interfaces,
// but remain independent of specific adapter implementations through dependency injection.
package app
