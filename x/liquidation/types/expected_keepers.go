package types

import (
	"context"

	"cosmossdk.io/core/address"
	"cosmossdk.io/math"
	collateraltypes "github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AuthKeeper defines the expected auth keeper interface
type AuthKeeper interface {
	AddressCodec() address.Codec
	GetAccount(ctx context.Context, addr sdk.AccAddress) sdk.AccountI
}

/*
@desc: CollateralKeeper defines the subset of collateral module keeper methods the
       liquidation module calls during position inspection and liquidation execution.
@fix:  All position-mutating methods use sdk.Context (not context.Context) to match
       the collateral keeper's actual method signatures. Go interface satisfaction is
       structural and exact — a mismatch in the context type causes a compile error
       because sdk.Context and context.Context are distinct types (sdk.Context wraps it).
       GetParams keeps context.Context since that method was authored with the standard
       interface-friendly type.
*/
type CollateralKeeper interface {
	GetPosition(ctx sdk.Context, owner string) (collateraltypes.Position, error)
	SetPosition(ctx sdk.Context, position collateraltypes.Position) error
	DeletePosition(ctx sdk.Context, owner string) error
	GetCollateralValueUSD(ctx sdk.Context, amountStr string, denom string) (math.LegacyDec, error)
	GetCollateralRatio(ctx sdk.Context, position collateraltypes.Position) (math.LegacyDec, error)
	IsPositionHealthy(ctx sdk.Context, position collateraltypes.Position) (bool, error)
	CalculateLiquidationOutput(ctx sdk.Context, position collateraltypes.Position) (collateralToGive math.Int, penaltyUSD math.LegacyDec, debtToRepayUVUSD math.Int, err error)
	GetParams(ctx context.Context) (collateraltypes.Params, error)
}

/*
@desc: OracleKeeper defines the price-lookup interface the liquidation module needs to
       evaluate whether a collateral position has fallen below its liquidation ratio.
@fix:  Methods use sdk.Context instead of context.Context to match oraclekeeper.Keeper's
       actual signatures, enabling the concrete keeper to satisfy this interface.
*/
type OracleKeeper interface {
	GetPriceValue(ctx sdk.Context, denom string) (math.LegacyDec, error)
	IsPriceStale(ctx sdk.Context, denom string, maxAgeSeconds int64) (bool, error)
}

// StablecoinKeeper defines the expected stablecoin keeper interface
type StablecoinKeeper interface {
	MintStablecoin(ctx context.Context, recipient sdk.AccAddress, amount math.Int) error
	BurnStablecoin(ctx context.Context, owner sdk.AccAddress, amount math.Int) error
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}
