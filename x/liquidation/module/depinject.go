package liquidation

import (
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/Vesper-Interchain/vesper-interchain/x/liquidation/keeper"
	"github.com/Vesper-Interchain/vesper-interchain/x/liquidation/types"
	/*
	@desc: Concrete keeper types imported for depinject field resolution.
	@fix:  depinject resolves wire types by their exact Go type. Declaring
	       CollateralKeeper / OracleKeeper / StablecoinKeeper as their interface types
	       causes "no provider for type" at startup because depinject only knows the
	       concrete types emitted by other modules' ProvideModule functions.
	       Concrete types satisfy the interfaces declared in expected_keepers.go, so
	       the keeper constructor still receives values through the interface contract.
	*/
	collateralkeeper "github.com/Vesper-Interchain/vesper-interchain/x/collateral/keeper"
	oraclekeeper "github.com/Vesper-Interchain/vesper-interchain/x/oracle/keeper"
	stablecoinkeeper "github.com/Vesper-Interchain/vesper-interchain/x/stablecoin/keeper"
)

var _ depinject.OnePerModuleType = AppModule{}

func (AppModule) IsOnePerModuleType() {}

func init() {
	appconfig.Register(
		&types.Module{},
		appconfig.Provide(ProvideModule),
	)
}

type ModuleInputs struct {
	depinject.In

	Config       *types.Module
	StoreService store.KVStoreService
	Cdc          codec.Codec
	AddressCodec address.Codec

	BankKeeper types.BankKeeper
	// @desc: Concrete types — depinject matches by exact Go type, not by interface.
	// The keeper constructor accepts the corresponding interfaces; Go's structural
	// typing handles the implicit conversion.
	CollateralKeeper collateralkeeper.Keeper
	OracleKeeper     oraclekeeper.Keeper
	StablecoinKeeper stablecoinkeeper.Keeper
}

type ModuleOutputs struct {
	depinject.Out

	LiquidationKeeper keeper.Keeper
	Module            appmodule.AppModule
}

func ProvideModule(in ModuleInputs) ModuleOutputs {
	authority := authtypes.NewModuleAddress(types.GovModuleName)
	if in.Config.Authority != "" {
		authority = authtypes.NewModuleAddressOrBech32Address(in.Config.Authority)
	}
	k := keeper.NewKeeper(
		in.StoreService,
		in.Cdc,
		in.AddressCodec,
		authority,
		in.CollateralKeeper,
		in.OracleKeeper,
		in.StablecoinKeeper,
		in.BankKeeper,
	)
	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{LiquidationKeeper: k, Module: m}
}
