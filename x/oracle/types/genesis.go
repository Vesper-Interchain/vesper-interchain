package types

// DefaultGenesis returns the default genesis state for the oracle module.
// On a fresh chain the Params are initialised to their defaults and the
// Prices map is empty; the oracle feeder must submit at least one price
// before any collateral vault operations can proceed.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic sanity checks on the genesis state, ensuring that
// the embedded Params satisfy all invariants before the chain starts.
func (gs GenesisState) Validate() error {
	return gs.Params.Validate()
}
