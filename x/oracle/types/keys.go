package types

import "cosmossdk.io/collections"

const (
    ModuleName = "oracle"
    StoreKey   = ModuleName

    GovModuleName = "gov"
)

var (
    // Params storage
    ParamsKey = collections.NewPrefix("p_oracle")

    // Price storage
    PriceKey = collections.NewPrefix("price")
)