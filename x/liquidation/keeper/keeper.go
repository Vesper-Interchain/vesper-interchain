// Package keeper implements the liquidation module's state management.
// The liquidation module provides a queue-based mechanism for coordinating
// multi-step liquidations. Any account can enqueue an unhealthy position;
// any account holding sufficient uvusd can then execute the queued liquidation.
// This two-step design allows bots and searchers to monitor the queue and
// compete to execute profitable liquidations.
package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/Vesper-Interchain/vesper-interchain/x/liquidation/types"
)

// Keeper is the liquidation module keeper. It holds the queue state and
// references to the keepers it calls when executing a liquidation.
type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec

	// authority is the governance address that can update module parameters.
	authority []byte

	// Schema is the collections schema, needed for genesis import/export.
	Schema collections.Schema

	// Params stores module-level governance parameters.
	Params collections.Item[types.Params]

	// Queue is the pending liquidation map: position_id (uint64) → LiquidationQueue entry.
	// Entries are added by EnqueuePosition and removed by RemoveQueueEntry after execution.
	Queue collections.Map[uint64, types.LiquidationQueue]

	// PositionCounter is a monotonically increasing sequence used to assign
	// unique IDs to queued positions. IDs are never reused.
	PositionCounter collections.Sequence

	collateralKeeper types.CollateralKeeper
	oracleKeeper     types.OracleKeeper
	stablecoinKeeper types.StablecoinKeeper
	bankKeeper       types.BankKeeper
}

// NewKeeper constructs a new liquidation Keeper and registers all state collections.
func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,
	collateralKeeper types.CollateralKeeper,
	oracleKeeper types.OracleKeeper,
	stablecoinKeeper types.StablecoinKeeper,
	bankKeeper types.BankKeeper,
) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService:     storeService,
		cdc:              cdc,
		addressCodec:     addressCodec,
		authority:        authority,
		collateralKeeper: collateralKeeper,
		oracleKeeper:     oracleKeeper,
		stablecoinKeeper: stablecoinKeeper,
		bankKeeper:       bankKeeper,
		Params:           collections.NewItem(sb, types.ParamsKey, "params", codec.CollValue[types.Params](cdc)),
		Queue:            collections.NewMap(sb, types.LiquidationQueuePrefix, "liquidation_queue", collections.Uint64Key, codec.CollValue[types.LiquidationQueue](cdc)),
		PositionCounter:  collections.NewSequence(sb, types.PositionCounterKey, "position_counter"),
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
