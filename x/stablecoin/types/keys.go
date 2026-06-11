// Package types contains all type definitions, constants, error codes, events,
// and keeper interfaces for the stablecoin module. The stablecoin module manages
// the minting and burning of uvusd (the Vesper USD stablecoin) and maintains
// lifetime supply counters for protocol-level accounting.
package types

import "cosmossdk.io/collections"

const (
	// ModuleName is the canonical name of the stablecoin module.
	ModuleName = "stablecoin"

	// StoreKey is the KV-store key registered for the stablecoin module.
	StoreKey = ModuleName

	// GovModuleName is duplicated here to avoid importing x/gov directly.
	GovModuleName = "gov"
)

var (
	// ParamsKey is the collections prefix for governance-controlled parameters.
	ParamsKey = collections.NewPrefix(0)

	// SupplyKey is the collections prefix for the SupplyState item that tracks
	// lifetime TotalMinted and TotalBurned values.
	SupplyKey = collections.NewPrefix(1)
)
