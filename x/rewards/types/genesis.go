package types

// GenesisState defines the rewards module genesis state.
// Only Params are included; the accumulator and per-user shares are initialised
// to zero on every fresh start and rebuilt as deposits arrive.
type GenesisState struct {
	Params Params `json:"params"`
}

// DefaultGenesisState returns a GenesisState with default parameters
// (RewardRate = 1,000,000 uvusd/block, StableDenom = "uvusd").
func DefaultGenesisState() GenesisState {
	return GenesisState{
		Params: DefaultParams(),
	}
}

// Validate performs basic sanity checks on the genesis state.
func (g GenesisState) Validate() error {
	return g.Params.Validate()
}
