package oracle

import (
	"math/rand"

	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	oraclesimulation "github.com/Vesper-Interchain/vesper-interchain/x/oracle/simulation"
	"github.com/Vesper-Interchain/vesper-interchain/x/oracle/types"
)

// GenerateGenesisState creates a randomized GenState of the module.
func (AppModule) GenerateGenesisState(simState *module.SimulationState) {
	accs := make([]string, len(simState.Accounts))
	for i, acc := range simState.Accounts {
		accs[i] = acc.Address.String()
	}
	oracleGenesis := types.GenesisState{
		Params: types.DefaultParams(),
	}
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(&oracleGenesis)
}

// RegisterStoreDecoder registers a decoder.
func (am AppModule) RegisterStoreDecoder(_ simtypes.StoreDecoderRegistry) {}

// WeightedOperations returns the all the gov module operations with their respective weights.
func (am AppModule) WeightedOperations(simState module.SimulationState) []simtypes.WeightedOperation {
	operations := make([]simtypes.WeightedOperation, 0)
	const (
		opWeightMsgUpdatePrice          = "op_weight_msg_oracle"
		defaultWeightMsgUpdatePrice int = 100
	)

	var weightMsgUpdatePrice int
	simState.AppParams.GetOrGenerate(opWeightMsgUpdatePrice, &weightMsgUpdatePrice, nil,
		func(_ *rand.Rand) {
			weightMsgUpdatePrice = defaultWeightMsgUpdatePrice
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsgUpdatePrice,
		oraclesimulation.SimulateMsgUpdatePrice(am.authKeeper, am.bankKeeper, am.keeper, simState.TxConfig),
	))

	return operations
}

// ProposalMsgs returns msgs used for governance proposals for simulations.
func (am AppModule) ProposalMsgs(simState module.SimulationState) []simtypes.WeightedProposalMsg {
	return []simtypes.WeightedProposalMsg{}
}
