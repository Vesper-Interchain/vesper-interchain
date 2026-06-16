package types

import (
	"context"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// RewardsKeeper is the optional interface that the collateral keeper uses to
// propagate share changes to the rewards module. Defined here (not in x/rewards)
// to avoid a circular import: collateral → rewards → bank; rewards → collateral
// would create a cycle. The collateral keeper holds a nil pointer for this until
// app.go wires it via SetRewardsKeeper after both modules are initialised.
type RewardsKeeper interface {
	// UpdateShares is called after every deposit or withdrawal. newShares is the
	// user's updated collateral balance in uatom; it becomes their reward weight.
	UpdateShares(ctx sdk.Context, owner sdk.AccAddress, newShares math.Int) error
}

// BankKeeper defines the subset of the bank module's keeper that the collateral
// module requires for moving tokens between user accounts and the module escrow.
type BankKeeper interface {
	// SendCoinsFromAccountToModule transfers coins from a user's account to the
	// named module account (used when depositing collateral).
	SendCoinsFromAccountToModule(
		ctx context.Context,
		senderAddr sdk.AccAddress,
		recipientModule string,
		amt sdk.Coins,
	) error

	// SendCoinsFromModuleToAccount transfers coins from the module account back
	// to a user's account (used when withdrawing collateral or paying liquidators).
	SendCoinsFromModuleToAccount(
		ctx context.Context,
		senderModule string,
		recipientAddr sdk.AccAddress,
		amt sdk.Coins,
	) error

	// GetBalance returns the current balance of a specific denom for an address.
	GetBalance(
		ctx context.Context,
		addr sdk.AccAddress,
		denom string,
	) sdk.Coin
}

/*
@desc: OracleKeeper defines the subset of the oracle module's keeper that the
       collateral module requires for price lookups and staleness checks.
@fix:  Methods use sdk.Context instead of context.Context to match the oracle keeper's
       actual Go signatures. Go interface satisfaction is structural (exact signature match)
       — if the interface declares context.Context but the oracle keeper uses sdk.Context,
       oraclekeeper.Keeper does NOT implement OracleKeeper and compilation fails.
       sdk.Context embeds context.Context so it is a strict superset.
*/
type OracleKeeper interface {
	GetPriceValue(
		ctx sdk.Context,
		denom string,
	) (math.LegacyDec, error)

	IsPriceStale(
		ctx sdk.Context,
		denom string,
		maxAgeSeconds int64,
	) (bool, error)
}

// StablecoinKeeper defines the subset of the stablecoin module's keeper that
// the collateral module uses to create and destroy uvusd tokens.
type StablecoinKeeper interface {
	// MintStablecoin creates new uvusd and credits them to the recipient.
	MintStablecoin(
		ctx context.Context,
		recipient sdk.AccAddress,
		amount math.Int,
	) error

	// BurnStablecoin transfers uvusd from the owner to the module account and
	// then burns them. Both steps happen atomically inside this call.
	BurnStablecoin(
		ctx context.Context,
		owner sdk.AccAddress,
		amount math.Int,
	) error
}
