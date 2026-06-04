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
					Short:     "Shows the parameters of the module",
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service:              types.Msg_serviceDesc.ServiceName,
			EnhanceCustomCommand: true, // only required if you want to use the custom command
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      true, // skipped because authority gated
				},
				{
					RpcMethod:      "DepositCollateral",
					Use:            "deposit-collateral [denom] [amount]",
					Short:          "Send a depositCollateral tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "denom"}, {ProtoField: "amount"}},
				},
				{
					RpcMethod:      "WithdrawCollateral",
					Use:            "withdraw-collateral [position-id] [amount]",
					Short:          "Send a withdrawCollateral tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "position_id"}, {ProtoField: "amount"}},
				},
				{
					RpcMethod:      "MintStablecoin",
					Use:            "mint-stablecoin [position-id] [amount]",
					Short:          "Send a mintStablecoin tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "position_id"}, {ProtoField: "amount"}},
				},
				{
					RpcMethod:      "RepayDebt",
					Use:            "repay-debt [position-id] [amount]",
					Short:          "Send a repayDebt tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "position_id"}, {ProtoField: "amount"}},
				},
				{
					RpcMethod:      "LiquidatePosition",
					Use:            "liquidate-position [position-id]",
					Short:          "Send a liquidatePosition tx",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{{ProtoField: "position_id"}},
				},
			},
		},
	}
}
