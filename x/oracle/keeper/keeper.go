// Package keeper implements the oracle module's state management and business logic.
// The oracle module is responsible for storing and retrieving on-chain price data.
// Only the designated oracle address (stored in Params.OracleAddress) is authorised
// to submit new prices via MsgUpdatePrice. All other modules that need price feeds
// (primarily x/collateral) read from this keeper.
package keeper

import (
	"fmt"

	"cosmossdk.io/collections"
	"cosmossdk.io/core/address"
	corestore "cosmossdk.io/core/store"
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/Vesper-Interchain/vesper-interchain/x/oracle/types"
)

// Keeper is the oracle module keeper. It holds all persistent state for the module
// and exposes methods that the message server, query server, and other modules use.
type Keeper struct {
	storeService corestore.KVStoreService
	cdc          codec.Codec
	addressCodec address.Codec

	// authority is the address that can update module parameters via governance.
	// This is typically the x/gov module account address.
	authority []byte

	// Schema is the collections schema used to register all state collections.
	Schema collections.Schema

	// Params stores the module's governance-controlled parameters (e.g. oracle address).
	Params collections.Item[types.Params]

	// Prices maps a denom string (e.g. "uatom") to its latest Price entry.
	// Each price entry includes the decimal price string, the block timestamp, and the source.
	Prices collections.Map[string, types.Price]

	// bankKeeper is used for simulation purposes only (checking spendable coins).
	bankKeeper types.BankKeeper
}

// NewKeeper constructs a new oracle Keeper. It registers all collections against the
// provided store service and builds the collections schema. Panics on invalid authority
// or schema build failure, as these are unrecoverable programmer errors.
func NewKeeper(
	storeService corestore.KVStoreService,
	cdc codec.Codec,
	addressCodec address.Codec,
	authority []byte,
	bankKeeper types.BankKeeper,
) Keeper {
	if _, err := addressCodec.BytesToString(authority); err != nil {
		panic(fmt.Sprintf("invalid authority address %s: %s", authority, err))
	}

	sb := collections.NewSchemaBuilder(storeService)

	k := Keeper{
		storeService: storeService,
		cdc:          cdc,
		addressCodec: addressCodec,
		authority:    authority,
		bankKeeper:   bankKeeper,

		Params: collections.NewItem(
			sb,
			types.ParamsKey,
			"params",
			codec.CollValue[types.Params](cdc),
		),

		Prices: collections.NewMap(
			sb,
			types.PriceKey,
			"prices",
			collections.StringKey,
			codec.CollValue[types.Price](cdc),
		),
	}

	schema, err := sb.Build()
	if err != nil {
		panic(err)
	}
	k.Schema = schema

	return k
}

// GetAuthority returns the raw authority address bytes. This is the address that
// must sign MsgUpdateParams proposals for the oracle module.
func (k Keeper) GetAuthority() []byte {
	return k.authority
}
