package types

import "cosmossdk.io/collections"

const (
    ModuleName = "oracle"
    StoreKey   = ModuleName

    GovModuleName = "gov"
)

var (
    // Params storage
    ParamsKey = collections.NewPrefix(0)

    // Price storage
    PriceKey = collections.NewPrefix(1)
)