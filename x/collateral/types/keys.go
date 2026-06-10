package types

import "cosmossdk.io/collections"

const (
    ModuleName = "collateral"
    StoreKey   = ModuleName
    GovModuleName = "gov"
)

var (
    ParamsKey   = collections.NewPrefix(0)
    PositionKey = collections.NewPrefix(1)  // owner → Position
)
