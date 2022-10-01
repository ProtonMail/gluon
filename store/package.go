// package store implements types that store message literals.
//
// Messages may be stored either in-memory or on-disk.
// When stored on disk, they are stored encrypted and optionally compressed.
package store

//go:generate mockgen -destination mock_store/store.go . Store
