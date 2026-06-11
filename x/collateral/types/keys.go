// Package types contains all type definitions, constants, error codes, events,
// and keeper interfaces for the collateral module. The collateral module is the
// heart of the Vesper protocol: it manages user vaults (Positions), enforces
// collateral ratio rules, and coordinates with the oracle and stablecoin modules.
package types

import "cosmossdk.io/collections"

const (
	// ModuleName is the canonical name of the collateral module, used for store
	// registration, event emission, and module account naming.
	ModuleName = "collateral"

	// StoreKey is the KV-store key for the collateral module.
	StoreKey = ModuleName

	// GovModuleName is duplicated here to avoid importing x/gov directly.
	GovModuleName = "gov"
)

var (
	// ParamsKey is the collections prefix for the module's governance parameters.
	ParamsKey = collections.NewPrefix(0)

	// PositionKey is the collections prefix for the Positions map (owner → Position).
	PositionKey = collections.NewPrefix(1)
)
