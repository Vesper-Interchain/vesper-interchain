// Package keeper implements the rewards module state management.
// The rewards module uses the MasterChef accumulator pattern to distribute
// uvusd rewards to collateral depositors proportionally to their deposited
// collateral (shares).
//
// MasterChef pattern overview:
//   - A global RewardAccumulator (acc) increases by (rewardRate / totalShares) each block.
//   - Each user has a "reward debt" — the value of acc at the time they last
//     updated their shares or claimed rewards.
//   - pending = (acc - userDebt) * userShares
//   - When a user deposits or withdraws, pending is settled immediately to
//     prevent gaming the accumulator, then shares and debt are updated.
//
// This approach avoids iterating all users every block: the accumulator grows
// once per block (O(1)), and individual user accounting happens only on demand.
package keeper

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/store"
	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/Vesper-Interchain/vesper-interchain/x/rewards/types"
)

// Keeper holds the rewards module state and exposes the accumulator logic.
type Keeper struct {
	storeService store.KVStoreService
	bankKeeper   types.BankKeeper

	// authority is the address (x/gov module account) that can update Params.
	authority []byte

	// Schema is the collections schema, needed for genesis import/export.
	Schema collections.Schema

	// RawParams stores module parameters as JSON-encoded bytes.
	// JSON is used instead of protobuf because the rewards module is registered
	// manually (no buf-generated codec), keeping the build simpler.
	RawParams collections.Item[[]byte]

	// RewardAccumulator is the global per-share accumulator value stored as a
	// LegacyDec string. It grows monotonically each block.
	RewardAccumulator collections.Item[string]

	// TotalShares is the sum of all depositor collateral amounts currently tracked.
	// Stored as an Int string. Updated whenever UpdateShares is called.
	TotalShares collections.Item[string]

	// UserShares maps each owner's bech32 address to their current share amount (Int string).
	UserShares collections.Map[string, string]

	// UserRewardDebt maps each owner's bech32 address to their reward debt (LegacyDec string).
	// The debt records the value of RewardAccumulator at the last settlement.
	UserRewardDebt collections.Map[string, string]
}

// NewKeeper constructs a new rewards Keeper and registers all state collections.
func NewKeeper(
	storeService store.KVStoreService,
	bankKeeper types.BankKeeper,
	authority []byte,
) Keeper {
	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService:      storeService,
		bankKeeper:        bankKeeper,
		authority:         authority,
		RawParams:         collections.NewItem(sb, types.ParamsKey, "params", collections.BytesValue),
		RewardAccumulator: collections.NewItem(sb, types.RewardAccumulatorKey, "reward_accumulator", collections.StringValue),
		TotalShares:       collections.NewItem(sb, types.TotalSharesKey, "total_shares", collections.StringValue),
		UserShares:        collections.NewMap(sb, types.UserSharesPrefix, "user_shares", collections.StringKey, collections.StringValue),
		UserRewardDebt:    collections.NewMap(sb, types.UserRewardDebtPrefix, "user_reward_debt", collections.StringKey, collections.StringValue),
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

// GetParams deserialises and returns the current module parameters.
// Falls back to DefaultParams if no params have been written yet (first block).
func (k Keeper) GetParams(ctx sdk.Context) (types.Params, error) {
	bz, err := k.RawParams.Get(ctx)
	if err != nil {
		return types.DefaultParams(), nil
	}
	var p types.Params
	if err := json.Unmarshal(bz, &p); err != nil {
		return types.Params{}, err
	}
	return p, nil
}

// SetParams serialises and stores the module parameters.
func (k Keeper) SetParams(ctx sdk.Context, p types.Params) error {
	bz, err := json.Marshal(p)
	if err != nil {
		return err
	}
	return k.RawParams.Set(ctx, bz)
}

// getRewardAccumulator reads the current global accumulator value.
// Returns zero if no value has been written yet (fresh chain).
func (k Keeper) getRewardAccumulator(ctx sdk.Context) math.LegacyDec {
	s, err := k.RewardAccumulator.Get(ctx)
	if err != nil {
		return math.LegacyZeroDec()
	}
	acc, err := math.LegacyNewDecFromStr(s)
	if err != nil {
		return math.LegacyZeroDec()
	}
	return acc
}

// setRewardAccumulator persists the updated accumulator value.
func (k Keeper) setRewardAccumulator(ctx sdk.Context, acc math.LegacyDec) error {
	return k.RewardAccumulator.Set(ctx, acc.String())
}

// getTotalShares returns the sum of all user share amounts.
// Returns zero if no depositors have registered yet.
func (k Keeper) getTotalShares(ctx sdk.Context) math.Int {
	s, err := k.TotalShares.Get(ctx)
	if err != nil {
		return math.ZeroInt()
	}
	total, ok := math.NewIntFromString(s)
	if !ok {
		return math.ZeroInt()
	}
	return total
}

// setTotalShares persists the updated total shares value.
func (k Keeper) setTotalShares(ctx sdk.Context, total math.Int) error {
	return k.TotalShares.Set(ctx, total.String())
}

// getUserShares returns the share amount currently registered for an owner.
// Returns zero if the owner has no registered shares.
func (k Keeper) getUserShares(ctx sdk.Context, owner string) math.Int {
	s, err := k.UserShares.Get(ctx, owner)
	if err != nil {
		return math.ZeroInt()
	}
	shares, ok := math.NewIntFromString(s)
	if !ok {
		return math.ZeroInt()
	}
	return shares
}

// getUserRewardDebt returns the accumulator snapshot recorded at the last
// share update or claim for the given owner. Returns zero if no debt exists.
func (k Keeper) getUserRewardDebt(ctx sdk.Context, owner string) math.LegacyDec {
	s, err := k.UserRewardDebt.Get(ctx, owner)
	if err != nil {
		return math.LegacyZeroDec()
	}
	debt, err := math.LegacyNewDecFromStr(s)
	if err != nil {
		return math.LegacyZeroDec()
	}
	return debt
}

// calcPendingRewards computes the outstanding claimable rewards for an owner
// without modifying any state. This is a pure read operation.
//
// Formula: pending = (accumulator - userDebt) * userShares
func (k Keeper) calcPendingRewards(ctx sdk.Context, owner string) math.Int {
	shares := k.getUserShares(ctx, owner)
	if shares.IsZero() {
		return math.ZeroInt()
	}
	acc := k.getRewardAccumulator(ctx)
	debt := k.getUserRewardDebt(ctx, owner)

	delta := acc.Sub(debt)
	if delta.IsNegative() {
		// Accumulator should never go backwards; treat negative delta as zero.
		return math.ZeroInt()
	}

	return delta.MulInt(shares).TruncateInt()
}

// AccumulateRewards increases the global accumulator by (rewardRate / totalShares).
// This is called once per block from the module's EndBlocker. If there are no
// depositors (totalShares == 0) or the reward rate is zero, the function is a no-op.
func (k Keeper) AccumulateRewards(ctx sdk.Context) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	total := k.getTotalShares(ctx)
	if total.IsZero() {
		// No depositors; nothing to distribute this block.
		return nil
	}

	rate, err := params.GetRewardRateAsInt()
	if err != nil {
		return err
	}
	if rate.IsZero() {
		return nil
	}

	// Increment per share = total reward this block / total shares.
	increment := math.LegacyNewDecFromInt(rate).QuoInt(total)
	acc := k.getRewardAccumulator(ctx)
	return k.setRewardAccumulator(ctx, acc.Add(increment))
}

// UpdateShares is called by the collateral keeper whenever a user's deposit
// changes (deposit, withdraw, or liquidation). It performs three steps:
//  1. Settle any pending rewards at the old share amount before the change.
//  2. Update TotalShares and UserShares to the new value.
//  3. Reset UserRewardDebt to the current accumulator so the next pending
//     calculation starts from the correct baseline.
func (k Keeper) UpdateShares(ctx sdk.Context, owner sdk.AccAddress, newShares math.Int) error {
	ownerStr := owner.String()

	// Settle pending rewards before changing shares to prevent retroactive
	// over-accrual at the new (higher) share amount.
	pending := k.calcPendingRewards(ctx, ownerStr)
	if pending.IsPositive() {
		if err := k.mintRewards(ctx, owner, pending); err != nil {
			return fmt.Errorf("failed to settle rewards before share update: %w", err)
		}
	}

	// Adjust total shares: remove old contribution, add new contribution.
	oldShares := k.getUserShares(ctx, ownerStr)
	total := k.getTotalShares(ctx)
	total = total.Sub(oldShares).Add(newShares)
	if total.IsNegative() {
		total = math.ZeroInt()
	}
	if err := k.setTotalShares(ctx, total); err != nil {
		return err
	}

	if err := k.UserShares.Set(ctx, ownerStr, newShares.String()); err != nil {
		return err
	}

	// Snap the debt to the current accumulator so the next pending calculation
	// will only count rewards earned after this update.
	acc := k.getRewardAccumulator(ctx)
	if err := k.UserRewardDebt.Set(ctx, ownerStr, acc.String()); err != nil {
		return err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeUpdateShares,
			sdk.NewAttribute(types.AttributeKeyOwner, ownerStr),
			sdk.NewAttribute(types.AttributeKeyShares, newShares.String()),
		),
	)

	return nil
}

// ClaimRewards mints pending uvusd rewards and sends them to the owner.
// After the claim the user's reward debt is reset to prevent double-claiming.
// Returns ErrNoRewardsToClaim if the calculated pending amount is zero.
func (k Keeper) ClaimRewards(ctx sdk.Context, owner sdk.AccAddress) (math.Int, error) {
	ownerStr := owner.String()

	pending := k.calcPendingRewards(ctx, ownerStr)
	if !pending.IsPositive() {
		return math.ZeroInt(), types.ErrNoRewardsToClaim
	}

	if err := k.mintRewards(ctx, owner, pending); err != nil {
		return math.ZeroInt(), err
	}

	// Update the debt snapshot so the user cannot claim the same rewards again.
	acc := k.getRewardAccumulator(ctx)
	if err := k.UserRewardDebt.Set(ctx, ownerStr, acc.String()); err != nil {
		return math.ZeroInt(), err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			types.EventTypeClaimRewards,
			sdk.NewAttribute(types.AttributeKeyOwner, ownerStr),
			sdk.NewAttribute(types.AttributeKeyAmount, pending.String()),
		),
	)

	return pending, nil
}

// mintRewards creates new uvusd tokens and sends them directly to the recipient.
// The stablecoin denom is read from module Params so it remains configurable.
func (k Keeper) mintRewards(ctx sdk.Context, recipient sdk.AccAddress, amount math.Int) error {
	params, err := k.GetParams(ctx)
	if err != nil {
		return err
	}

	coins := sdk.NewCoins(sdk.NewCoin(params.StableDenom, amount))
	if err := k.bankKeeper.MintCoins(ctx, types.ModuleName, coins); err != nil {
		return err
	}
	return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, recipient, coins)
}

// GetPendingRewards is the read-only counterpart to ClaimRewards.
// Used by query handlers and the EVM precompile to report claimable amounts
// without modifying state.
func (k Keeper) GetPendingRewards(ctx sdk.Context, owner sdk.AccAddress) math.Int {
	return k.calcPendingRewards(ctx, owner.String())
}
