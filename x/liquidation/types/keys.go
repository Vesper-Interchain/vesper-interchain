// Package types contains all type definitions, constants, error codes, events,
// and keeper interfaces for the liquidation module. The liquidation module
// provides a queue-based liquidation pipeline: unhealthy positions are enqueued
// by any account, and any account holding sufficient uvusd can execute them.
package types

import "cosmossdk.io/collections"

const (
	// ModuleName is the canonical name of the liquidation module.
	ModuleName = "liquidation"

	// StoreKey is the KV-store key registered for the liquidation module.
	StoreKey = ModuleName

	// GovModuleName is duplicated here to avoid importing x/gov directly.
	// Must stay in sync with x/gov's own ModuleName constant.
	GovModuleName = "gov"
)

var (
	// ParamsKey is the collections prefix for governance-controlled module parameters.
	ParamsKey = collections.NewPrefix("p_liquidation")

	// LiquidationQueuePrefix is the collections prefix for the pending liquidation queue
	// map (uint64 position_id → LiquidationQueue entry).
	LiquidationQueuePrefix = collections.NewPrefix("lq_")

	/*
	@fix: PositionCounterKey prefix changed from "lq_seq" to "pc_".
	      The previous value "lq_seq" starts with "lq_" which is also the prefix of
	      LiquidationQueuePrefix ("lq_"). cosmossdk.io/collections uses byte-prefix
	      iteration to scan a collection; a sequence key whose bytes are a prefix of
	      queue entries (or vice-versa) causes the iterator to walk into the wrong
	      collection, silently returning wrong data or corrupting both.
	      "pc_" (position counter) is entirely disjoint from "lq_" so there is no overlap.
	*/
	PositionCounterKey = collections.NewPrefix("pc_")
)
