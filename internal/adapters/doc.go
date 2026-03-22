// Package adapters provides integration with external systems and resources.
//
// This layer implements the ports defined by the application layer, handling:
//   - Template filesystem operations (templatefs)
//   - TOML protocol parsing for manifests and replay files (protocoltoml)
//   - Git command execution (gitexec)
//   - Build information extraction (buildinfo)
//   - Embedded template sources (templatesrc)
//
// Adapters depend on domain types but are isolated from application logic,
// enabling independent testing and implementation changes without affecting business rules.
package adapters
