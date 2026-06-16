package collateral

import (
	"cosmossdk.io/core/address"
	"cosmossdk.io/core/appmodule"
	"cosmossdk.io/core/store"
	"cosmossdk.io/depinject"
	"cosmossdk.io/depinject/appconfig"
	"github.com/cosmos/cosmos-sdk/codec"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"

	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/keeper"
	"github.com/Vesper-Interchain/vesper-interchain/x/collateral/types"
	/*
	@desc: Import concrete keeper types instead of interface types for depinject injection.
	@fix:  depinject resolves providers by their exact Go type, not by interface. Declaring
	       fields as types.OracleKeeper / types.StablecoinKeeper (interfaces) causes depinject
	       to fail with "no provider for type" because the container only knows the concrete
	       oraclekeeper.Keeper and stablecoinkeeper.Keeper types that other modules export.
	       Using concrete types lets depinject wire them correctly while the keeper constructor
	       still accepts the interface — satisfying both depinject and the type system.
	*/
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
	// @desc: Concrete types — depinject matches providers by exact type, not by interface.
	// The keeper constructor accepts the interface; depinject hands in the concrete value.
	OracleKeeper     oraclekeeper.Keeper
	StablecoinKeeper stablecoinkeeper.Keeper
}

type ModuleOutputs struct {
	depinject.Out

	CollateralKeeper keeper.Keeper
	Module           appmodule.AppModule
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
		in.BankKeeper,
		in.OracleKeeper,
		in.StablecoinKeeper,
	)
	m := NewAppModule(in.Cdc, k)

	return ModuleOutputs{CollateralKeeper: k, Module: m}
}
