package types

// Event type and attribute constants for the collateral module.
// These are emitted as ABCI events on every successful state-changing operation
// so that external systems (block explorers, analytics, front-ends) can index
// vault activity without scanning raw transaction bodies.
const (
	// EventTypeDepositCollateral is emitted when uatom is added to a vault.
	EventTypeDepositCollateral = "collateral_deposited"

	// EventTypeWithdrawCollateral is emitted when uatom is returned from a vault.
	EventTypeWithdrawCollateral = "collateral_withdrawn"

	// EventTypeMintStablecoin is emitted when new uvusd is minted against collateral.
	EventTypeMintStablecoin = "stablecoin_minted"

	// EventTypeRepayDebt is emitted when uvusd is burned to reduce vault debt.
	EventTypeRepayDebt = "debt_repaid"

	// EventTypeLiquidatePosition is emitted when a vault is closed by a liquidator.
	EventTypeLiquidatePosition = "position_liquidated"

	// EventTypeUpdateParams is emitted when module parameters are changed via governance.
	EventTypeUpdateParams = "update_params"

	// AttributeKeyOwner is the bech32 address of the vault owner.
	AttributeKeyOwner = "owner"

	// AttributeKeyAmount is the token amount involved in the operation.
	AttributeKeyAmount = "amount"

	// AttributeKeyCollateral is the total collateral amount after the operation.
	AttributeKeyCollateral = "collateral_amount"

	// AttributeKeyDebt is the outstanding debt amount after the operation.
	AttributeKeyDebt = "debt_amount"

	// AttributeKeyCollateralRatio is the resulting collateral ratio after minting.
	AttributeKeyCollateralRatio = "collateral_ratio"

	// AttributeKeyLiquidator is the address of the account that executed the liquidation.
	AttributeKeyLiquidator = "liquidator"

	// AttributeKeyPenalty is the liquidation penalty in USD reported during a liquidation event.
	AttributeKeyPenalty = "penalty"
)
