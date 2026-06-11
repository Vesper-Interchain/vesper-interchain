package app

import (
	storetypes "cosmossdk.io/store/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	rewardskeeper "github.com/Vesper-Interchain/vesper-interchain/x/rewards/keeper"
	rewardsmodule "github.com/Vesper-Interchain/vesper-interchain/x/rewards/module"
	rewardstypes "github.com/Vesper-Interchain/vesper-interchain/x/rewards/types"
)

// registerRewardsModule registers the x/rewards module with the application.
//
// The rewards module is registered manually (not via depinject/app-wiring) because
// it was added after the initial app-wiring configuration and adding it to
// app_config.go would require proto-codec generation. The manual path is:
//  1. Allocate a KV store key.
//  2. Call RegisterStores so the multistore is aware of the key before Load().
//  3. Construct the keeper with the store service, bank keeper, and authority.
//  4. Wire the rewards keeper into the collateral keeper via SetRewardsKeeper so
//     that collateral deposit/withdraw events propagate share updates to rewards.
//  5. Call RegisterModules to include the module in the module manager.
//
// This function must be called BEFORE app.Load(loadLatest) in app.New() to ensure
// the store key is present in the multistore before the state is loaded from disk.
func (app *App) registerRewardsModule() error {
	rewardsStoreKey := storetypes.NewKVStoreKey(rewardstypes.StoreKey)

	if err := app.RegisterStores(rewardsStoreKey); err != nil {
		return err
	}

	app.RewardsKeeper = rewardskeeper.NewKeeper(
		runtime.NewKVStoreService(rewardsStoreKey),
		app.BankKeeper,
		// Use the governance module address as the authority so only governance
		// proposals can update rewards module parameters.
		authtypes.NewModuleAddress(govtypes.ModuleName),
	)

	// Inject the rewards keeper into the collateral keeper. The collateral keeper
	// holds a nil pointer for this dependency until this call, so vault operations
	// are safe to execute even before this wiring is complete.
	app.CollateralKeeper.SetRewardsKeeper(&app.RewardsKeeper)

	return app.RegisterModules(
		rewardsmodule.NewAppModule(app.appCodec, app.RewardsKeeper),
	)
}
