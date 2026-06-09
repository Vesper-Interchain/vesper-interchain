package types

import "cosmossdk.io/collections"

const (
	ModuleName = "stablecoin"
	StoreKey   = ModuleName

	// Used by MsgUpdateParams authority configuration
	GovModuleName = "gov"
)

var (
	ParamsKey      = collections.NewPrefix(0)
	TotalMintedKey = collections.NewPrefix(1)
	TotalBurnedKey = collections.NewPrefix(2)
)