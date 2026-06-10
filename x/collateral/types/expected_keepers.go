package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type BankKeeper interface {
	SendCoinsFromAccountToModule(
		ctx context.Context,
		senderAddr sdk.AccAddress,
		recipientModule string,
		amt sdk.Coins,
	) error

	SendCoinsFromModuleToAccount(
		ctx context.Context,
		senderModule string,
		recipientAddr sdk.AccAddress,
		amt sdk.Coins,
	) error

	GetBalance(
		ctx context.Context,
		addr sdk.AccAddress,
		denom string,
	) sdk.Coin
}

type OracleKeeper interface {
	GetPriceValue(
		ctx context.Context,
		denom string,
	) (math.LegacyDec, error)

	IsPriceStale(
		ctx context.Context,
		denom string,
		maxAgeSeconds int64,
	) (bool, error)
}

type StablecoinKeeper interface {
	MintStablecoin(
		ctx context.Context,
		recipient sdk.AccAddress,
		amount math.Int,
	) error

	BurnStablecoin(
		ctx context.Context,
		owner sdk.AccAddress,
		amount math.Int,
	) error
}
