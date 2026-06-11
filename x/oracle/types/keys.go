// Package types contains all type definitions, constants, errors, events, and
// interfaces for the oracle module. The oracle module stores on-chain price feeds
// that other modules (e.g. x/collateral) consume to compute collateral ratios.
package types

import "cosmossdk.io/collections"

const (
	// ModuleName is the canonical name used to identify the oracle module in the
	// module manager, store keys, and event attributes.
	ModuleName = "oracle"

	// StoreKey is the KV-store key registered for the oracle module.
	StoreKey = ModuleName

	// GovModuleName is duplicated here to avoid a direct import of x/gov.
	// It must stay in sync with the gov module's own module name constant.
	GovModuleName = "gov"
)

var (
	// ParamsKey is the collections prefix used to store oracle module parameters.
	ParamsKey = collections.NewPrefix(0)

	// PriceKey is the collections prefix used to store the price map (denom → Price).
	PriceKey = collections.NewPrefix(1)
)
