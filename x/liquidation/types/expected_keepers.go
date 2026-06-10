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

// CollateralKeeper defines the expected collateral keeper interface
type CollateralKeeper interface {
	GetPosition(ctx context.Context, owner string) (collateraltypes.Position, error)
	SetPosition(ctx context.Context, position collateraltypes.Position) error
	DeletePosition(ctx context.Context, owner string) error
	GetCollateralValueUSD(ctx context.Context, amountStr string, denom string) (math.LegacyDec, error)
	GetCollateralRatio(ctx context.Context, position collateraltypes.Position) (math.LegacyDec, error)
	IsPositionHealthy(ctx context.Context, position collateraltypes.Position) (bool, error)
	CalculateLiquidationOutput(ctx context.Context, position collateraltypes.Position) (collateralToGive math.Int, penaltyUSD math.LegacyDec, debtToRepayUVUSD math.Int, err error)
	GetParams(ctx context.Context) (collateraltypes.Params, error)
}

// OracleKeeper defines the expected oracle keeper interface
type OracleKeeper interface {
	GetPriceValue(ctx context.Context, denom string) (math.LegacyDec, error)
	IsPriceStale(ctx context.Context, denom string, maxAgeSeconds int64) (bool, error)
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
