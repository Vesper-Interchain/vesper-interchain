package collateral

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"

	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: types.Query_serviceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query collateral module parameters",
				},
				{
					RpcMethod: "Position",
					Use:       "position [owner]",
					Short:     "Query a collateral position by owner address",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "owner"},
					},
				},
				{
					RpcMethod: "CollateralRatio",
					Use:       "collateral-ratio [owner]",
					Short:     "Query the collateral ratio for a position",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "owner"},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // authority gated
				},
				{
					RpcMethod: "DepositCollateral",
					Use:       "deposit-collateral [amount]",
					Short:     "Deposit collateral to open or increase a vault position",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "amount"},
					},
				},
				{
					RpcMethod: "WithdrawCollateral",
					Use:       "withdraw-collateral [amount]",
					Short:     "Withdraw collateral from a vault position",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "amount"},
					},
				},
				{
					RpcMethod: "MintStablecoin",
					Use:       "mint-stablecoin [amount]",
					Short:     "Mint uVUSD stablecoin against deposited collateral",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "amount"},
					},
				},
				{
					RpcMethod: "RepayDebt",
					Use:       "repay-debt [amount]",
					Short:     "Repay uVUSD debt to reduce or close a vault position",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "amount"},
					},
				},
				{
					RpcMethod: "LiquidatePosition",
					Use:       "liquidate [owner]",
					Short:     "Liquidate an undercollateralised vault position",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "owner"},
					},
				},
			},
		},
	}
}
