package types

import "cosmossdk.io/collections"

const (
	ModuleName = "stablecoin"
	StoreKey   = ModuleName

	GovModuleName = "gov"
)

var (
	ParamsKey = collections.NewPrefix(0)
	SupplyKey = collections.NewPrefix(1)
)
