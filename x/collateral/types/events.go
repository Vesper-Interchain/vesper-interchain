package types

const (
	EventTypeDepositCollateral  = "collateral_deposited"
	EventTypeWithdrawCollateral = "collateral_withdrawn"
	EventTypeMintStablecoin     = "stablecoin_minted"
	EventTypeRepayDebt          = "debt_repaid"
	EventTypeLiquidatePosition  = "position_liquidated"
	EventTypeUpdateParams       = "update_params"

	AttributeKeyOwner           = "owner"
	AttributeKeyAmount          = "amount"
	AttributeKeyCollateral      = "collateral_amount"
	AttributeKeyDebt            = "debt_amount"
	AttributeKeyCollateralRatio = "collateral_ratio"
	AttributeKeyLiquidator      = "liquidator"
	AttributeKeyPenalty         = "penalty"
)
