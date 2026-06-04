package collateral

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	collateralsimulation "github.com/Vesper-Interchain/vesper-interchain/x/collateral/simulation"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	collateralGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&collateralGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	const (
		opWeightMsgDepositCollateral          = "op_weight_msg_collateral"
		defaultWeightMsgDepositCollateral int = 100
	)

	var weightMsgDepositCollateral int
	simState.AppParams.GetOrGenerate(opWeightMsgDepositCollateral, &weightMsgDepositCollateral, nil,
		func(_ *rand.Rand) {
			weightMsgDepositCollateral = defaultWeightMsgDepositCollateral
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgDepositCollateral,
		collateralsimulation.SimulateMsgDepositCollateral(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgWithdrawCollateral          = "op_weight_msg_collateral"
		defaultWeightMsgWithdrawCollateral int = 100
	)

	var weightMsgWithdrawCollateral int
	simState.AppParams.GetOrGenerate(opWeightMsgWithdrawCollateral, &weightMsgWithdrawCollateral, nil,
		func(_ *rand.Rand) {
			weightMsgWithdrawCollateral = defaultWeightMsgWithdrawCollateral
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgWithdrawCollateral,
		collateralsimulation.SimulateMsgWithdrawCollateral(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgMintStablecoin          = "op_weight_msg_collateral"
		defaultWeightMsgMintStablecoin int = 100
	)

	var weightMsgMintStablecoin int
	simState.AppParams.GetOrGenerate(opWeightMsgMintStablecoin, &weightMsgMintStablecoin, nil,
		func(_ *rand.Rand) {
			weightMsgMintStablecoin = defaultWeightMsgMintStablecoin
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgMintStablecoin,
		collateralsimulation.SimulateMsgMintStablecoin(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
