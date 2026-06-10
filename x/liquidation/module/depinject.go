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

	Config           *types.Module
	StoreService     store.KVStoreService
	Cdc              codec.Codec
	AddressCodec     address.Codec

	BankKeeper       types.BankKeeper
	CollateralKeeper types.CollateralKeeper
	OracleKeeper     types.OracleKeeper
	StablecoinKeeper types.StablecoinKeeper
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
