package stablecoin

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	stablecoinsimulation "github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/simulation"
	"github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/types"
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	stablecoinGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&stablecoinGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	const (
		opWeightMsgMintStablecoin          = "op_weight_msg_stablecoin"
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
		stablecoinsimulation.SimulateMsgMintStablecoin(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))
	const (
		opWeightMsgBurnStablecoin          = "op_weight_msg_stablecoin"
		defaultWeightMsgBurnStablecoin int = 100
	)

	var weightMsgBurnStablecoin int
	simState.AppParams.GetOrGenerate(opWeightMsgBurnStablecoin, &weightMsgBurnStablecoin, nil,
		func(_ *rand.Rand) {
			weightMsgBurnStablecoin = defaultWeightMsgBurnStablecoin
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgBurnStablecoin,
		stablecoinsimulation.SimulateMsgBurnStablecoin(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
