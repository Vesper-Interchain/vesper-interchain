package types

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:      DefaultParams(),
		TotalMinted: "0",
		TotalBurned: "0",
	}
}

// Validate performs basic genesis state validation
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	return nil
}
