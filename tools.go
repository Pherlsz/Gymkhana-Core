//go:build tools

// Package tools documents the repository tooling boundary.
//
// Tool binaries are installed at pinned versions by the Makefile instead of
// being imported here, keeping them out of the module dependency graph.
package tools
