package cmd

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"cosmossdk.io/log"
	confixcmd "cosmossdk.io/tools/confix/cmd"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/pruning"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/client/snapshot"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	authcmd "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"
	cosmosevmcmd "github.com/cosmos/evm/client"
	cosmosevmserver "github.com/cosmos/evm/server"

	"github.com/Vesper-Interchain/vesper-interchain/app"
)

// patchGenesisFeemarketNoBaseFee sets feemarket.params.no_base_fee = true in the
// written genesis.json so that the EVM dynamic fee checker falls back to the
// standard minimum-gas-prices check for native Cosmos txs. Without this, the
// NewMinGasPriceDecorator and NewDynamicFeeChecker ante decorators require fees
// in the EVM denom (atest) even for bank/staking txs, making the chain unusable
// without a cross-denom price oracle.
func patchGenesisFeemarketNoBaseFee(cmd *cobra.Command, _ []string) error {
	homeDir, _ := cmd.Flags().GetString(flags.FlagHome)
	if homeDir == "" {
		homeDir = app.DefaultNodeHome
	}

	genesisPath := filepath.Join(homeDir, "config", "genesis.json")
	bz, err := os.ReadFile(genesisPath)
	if err != nil {
		return err
	}

	var genesis map[string]interface{}
	if err := json.Unmarshal(bz, &genesis); err != nil {
		return err
	}

	appState, ok := genesis["app_state"].(map[string]interface{})
	if !ok {
		return nil
	}
	feemarket, ok := appState["feemarket"].(map[string]interface{})
	if !ok {
		return nil
	}
	params, ok := feemarket["params"].(map[string]interface{})
	if !ok {
		return nil
	}
	params["no_base_fee"] = true

	patched, err := json.MarshalIndent(genesis, "", " ")
	if err != nil {
		return err
	}
	return os.WriteFile(genesisPath, patched, 0o644)
}

func initRootCmd(
	rootCmd *cobra.Command,
	txConfig client.TxConfig,
	basicManager module.BasicManager,
) {
	initCmd := genutilcli.InitCmd(basicManager, app.DefaultNodeHome)
	initCmd.PostRunE = patchGenesisFeemarketNoBaseFee

	rootCmd.AddCommand(
		initCmd,
		NewInPlaceTestnetCmd(),
		NewTestnetMultiNodeCmd(basicManager, banktypes.GenesisBalancesIterator{}),
		debug.Cmd(),
		confixcmd.ConfigCommand(),
		pruning.Cmd(newApp, app.DefaultNodeHome),
		snapshot.Cmd(newApp),
	)
	cosmosevmserver.AddCommands(
		rootCmd,
		cosmosevmserver.NewDefaultStartOptions(func(l log.Logger, d dbm.DB, w io.Writer, ao servertypes.AppOptions) cosmosevmserver.Application {
			return newApp(l, d, w, ao).(cosmosevmserver.Application)
		}, app.DefaultNodeHome),
		appExport,
		addModuleInitFlags,
	)

	// add keybase, auxiliary RPC, query, genesis, and tx child commands
	rootCmd.AddCommand(
		server.StatusCommand(),
		genutilcli.Commands(txConfig, basicManager, app.DefaultNodeHome),
		queryCommand(),
		txCommand(),
		cosmosevmcmd.KeyCommands(app.DefaultNodeHome, false),
	)
}

// addModuleInitFlags adds more flags to the start command.
func addModuleInitFlags(startCmd *cobra.Command) {
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.WaitTxCmd(),
		rpc.ValidatorCommand(),
		server.QueryBlockCmd(),
		authcmd.QueryTxsByEventsCmd(),
		server.QueryBlocksCmd(),
		authcmd.QueryTxCmd(),
		server.QueryBlockResultsCmd(),
	)

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         false,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcmd.GetSignCommand(),
		authcmd.GetSignBatchCommand(),
		authcmd.GetMultiSignCommand(),
		authcmd.GetMultiSignBatchCmd(),
		authcmd.GetValidateSignaturesCommand(),
		flags.LineBreak,
		authcmd.GetBroadcastCommand(),
		authcmd.GetEncodeCommand(),
		authcmd.GetDecodeCommand(),
		authcmd.GetSimulateCmd(),
	)

	return cmd
}

// newApp creates the application
func newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	baseappOptions := server.DefaultBaseappOptions(appOpts)

	return app.New(
		logger, db, traceStore, true,
		appOpts,
		baseappOptions...,
	)
}

// appExport creates a new app (optionally at a given height) and exports state.
func appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	var bApp *app.App

	// this check is necessary as we use the flag in x/upgrade.
	// we can exit more gracefully by checking the flag here.
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home not set")
	}

	viperAppOpts, ok := appOpts.(*viper.Viper)
	if !ok {
		return servertypes.ExportedApp{}, errors.New("appOpts is not viper.Viper")
	}

	appOpts = viperAppOpts
	if height != -1 {
		bApp = app.New(logger, db, traceStore, false, appOpts)
		if err := bApp.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	} else {
		bApp = app.New(logger, db, traceStore, true, appOpts)
	}

	return bApp.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}
