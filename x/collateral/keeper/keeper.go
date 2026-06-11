// Package keeper implements the collateral module's state management and core
// vault logic. A "vault" (stored as a Position) is the on-chain record of a
// single user's collateral deposit and outstanding stablecoin debt.
//
// Lifecycle of a vault:
//  1. DepositCollateral  — user sends uatom to the module, a Position is created or updated.
//  2. MintStablecoin     — user borrows uvusd against their collateral (up to MaxLTV).
//  3. RepayDebt          — user burns uvusd to reduce their debt.
//  4. WithdrawCollateral — user retrieves uatom once the collateral ratio stays healthy.
//  5. LiquidatePosition  — a third party closes an under-collateralised vault.
package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// Keeper is the collateral module keeper. It owns all vault state and coordinates
// with the oracle, stablecoin, and rewards keepers.
type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec

	// authority is the address (x/gov module account) allowed to update Params.
	authority []byte

	// Schema is the collections schema, required for genesis import/export.
	Schema collections.Schema

	// Params holds module-level configuration: liquidation ratio, max LTV,
	// minimum collateral amount, supported denom, and oracle staleness window.
	Params collections.Item[types.Params]

	// Positions maps owner bech32 address → Position.
	// A position exists as long as the user has non-zero collateral or debt.
	Positions collections.Map[string, types.Position]

	bankKeeper       types.BankKeeper
	oracleKeeper     types.OracleKeeper
	stablecoinKeeper types.StablecoinKeeper

	// rewardsKeeper is an optional hook; it is nil until registerRewardsModule
	// wires it via SetRewardsKeeper. Vault operations still succeed without it.
	rewardsKeeper types.RewardsKeeper
}

// NewKeeper constructs a new collateral Keeper. The rewards keeper is not injected
// here because of dependency ordering; it is wired afterwards via SetRewardsKeeper.
func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,
	bankKeeper types.BankKeeper,
	oracleKeeper types.OracleKeeper,
	stablecoinKeeper types.StablecoinKeeper,
) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address: %s", authority))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService:     storeService,
		cdc:              cdc,
		addressCodec:     addressCodec,
		authority:        authority,
		bankKeeper:       bankKeeper,
		oracleKeeper:     oracleKeeper,
		stablecoinKeeper: stablecoinKeeper,
		Params:           collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Positions:        collections.NewMap(sb, types.PositionKey, "positions", collections.StringKey, codec.CollValue[types.Position](cdc)),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the authority address bytes.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}

// SetRewardsKeeper injects the rewards keeper after both the collateral and
// rewards modules have been initialised. This avoids a circular dependency
// at construction time: collateral needs rewards, rewards needs bank, bank
// needs nothing. Called from app/rewards.go after both keepers exist.
func (k *Keeper) SetRewardsKeeper(rk types.RewardsKeeper) {
	k.rewardsKeeper = rk
}

// notifyRewards propagates a collateral share change to the rewards module.
// newShares should equal the user's updated CollateralAmount in uatom.
// If the rewards keeper has not been wired yet the call is a no-op so that
// the collateral module can operate independently during testing.
func (k Keeper) notifyRewards(ctx sdk.Context, owner sdk.AccAddress, newShares math.Int) {
	if k.rewardsKeeper == nil {
		return
	}
	// Errors are swallowed intentionally: a rewards accounting failure must not
	// cause vault operations to revert for the user.
	_ = k.rewardsKeeper.UpdateShares(ctx, owner, newShares)
}
