package cmd

import (
	"os"

	"cosmossdk.io/client/v2/autocli"
	"cosmossdk.io/depinject"
	"cosmossdk.io/log"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtxconfig "github.com/cosmos/cosmos-sdk/x/auth/tx/config"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	evmcryptocodec "github.com/cosmos/evm/crypto/codec"
	"github.com/cosmos/evm/crypto/ethsecp256k1"
	cosmosevmkeyring "github.com/cosmos/evm/crypto/keyring"
	ibctransferevm "github.com/cosmos/evm/x/ibc/transfer"
	ibctransfer "github.com/cosmos/ibc-go/v10/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v10/modules/apps/transfer/types"
	"github.com/spf13/cobra"

	"github.com/Vesper-Interchain/vesper-interchain/app"
)

// NewRootCmd creates a new root command for vesper-interchaind. It is called once in the main function.
func NewRootCmd() *cobra.Command {
	var (
		autoCliOpts        autocli.AppOptions
		moduleBasicManager module.BasicManager
		clientCtx          client.Context
	)

	if err := depinject.Inject(
		depinject.Configs(app.AppConfig(),
			depinject.Supply(log.NewNopLogger()),
			depinject.Provide(
				ProvideClientContext,
				ProvideCodecWithEVMCrypto,
			), depinject.Provide(app.ProvideMsgEthereumTxCustomGetSigner),
		),
		&autoCliOpts,
		&moduleBasicManager,
		&clientCtx,
	); err != nil {
		panic(err)
	}

	rootCmd := &cobra.Command{
		Use:           app.Name + "d",
		Short:         "vesperinterchain node",
		SilenceErrors: true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			clientCtx = clientCtx.WithCmdContext(cmd.Context()).WithViper(app.Name)
			clientCtx, err := client.ReadPersistentCommandFlags(clientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			clientCtx, err = config.ReadFromClientConfig(clientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(clientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()
			customCMTConfig := initCometBFTConfig()

			if err := server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, customCMTConfig); err != nil {
				return err
			}

			// cosmos-sdk's interceptConfigs calls rootViper.Unmarshal(&customConfig) typed as `any`,
			// causing viper/mapstructure to replace the struct with a bare map — MinGasPrices is
			// lost and app.toml gets minimum-gas-prices = "". This fallback covers both existing
			// nodes with broken app.toml and edge cases where the template fix doesn't apply.
			serverCtx := server.GetServerContextFromCmd(cmd)
			if serverCtx.Viper.GetString(server.FlagMinGasPrices) == "" {
				serverCtx.Viper.Set(server.FlagMinGasPrices, "0stake")
			}

			return nil
		},
	}

	// Since the IBC modules don't support dependency injection, we need to
	// manually register the modules on the client side.
	// This needs to be removed after IBC supports App Wiring.
	ibcModules := app.RegisterIBC(clientCtx.Codec)
	for name, mod := range ibcModules {
		moduleBasicManager[name] = module.CoreAppModuleBasicAdaptor(name, mod)
		autoCliOpts.Modules[name] = mod
	}
	evmModules := app.RegisterEVM(clientCtx.Codec, clientCtx.InterfaceRegistry)
	for name, mod := range evmModules {
		moduleBasicManager[name] = module.CoreAppModuleBasicAdaptor(name, mod)
		autoCliOpts.Modules[name] = mod
	}

	moduleBasicManager[ibctransfertypes.ModuleName] = ibctransferevm.AppModuleBasic{
		AppModuleBasic: &ibctransfer.AppModuleBasic{},
	}

	initRootCmd(rootCmd, clientCtx.TxConfig, moduleBasicManager)

	if err := autoCliOpts.EnhanceRootCommand(rootCmd); err != nil {
		panic(err)
	}

	return rootCmd
}

/*
@desc: EVMCryptoConfig is a zero-size marker type used as a depinject dependency edge.
       ProvideCodecWithEVMCrypto returns it; ProvideClientContext consumes it (via blank
       parameter _). This forces depinject to sequence the EVM registration before the
       client context is built — without this ordering guarantee, the keyring would be
       constructed before ethsecp256k1 is known, causing "unsupported signing algo" errors.

@desc: ProvideCodecWithEVMCrypto registers ethsecp256k1 in two places on the client side:
       1. interfaceRegistry — so protobuf Any-unpacking of pubkeys works in CLI responses.
       2. legacyAmino — so amino-JSON signing mode and offline tx building work correctly.
       We do NOT call evmcryptocodec.RegisterCrypto here because it internally calls
       cryptocodec.RegisterCrypto, which re-registers secp256k1/ed25519 that depinject
       already wired, causing a "TypeInfo already exists" panic at startup.
*/

// EVMCryptoConfig is a marker type that indicates EVM crypto types have been registered
type EVMCryptoConfig struct{}

// ProvideCodecWithEVMCrypto registers EVM crypto types in the codec and returns a marker
// This ensures the registration happens before the codec is used for the keyring
func ProvideCodecWithEVMCrypto(
	interfaceRegistry codectypes.InterfaceRegistry,
	legacyAmino *codec.LegacyAmino,
) EVMCryptoConfig {
	// @desc: Register ethsecp256k1 in the interface registry for protobuf Any unpacking
	evmcryptocodec.RegisterInterfaces(interfaceRegistry)

	// @desc: Register ethsecp256k1 in legacy amino for amino-JSON sign mode and offline tx building
	// @fix:  Must NOT call evmcryptocodec.RegisterCrypto — it re-registers standard SDK types → panic
	legacyAmino.RegisterConcrete(&ethsecp256k1.PubKey{}, ethsecp256k1.PubKeyName, nil)
	legacyAmino.RegisterConcrete(&ethsecp256k1.PrivKey{}, ethsecp256k1.PrivKeyName, nil)

	return EVMCryptoConfig{}
}

/*
@desc: ProvideClientContext builds the client.Context for all CLI commands.
       WithKeyringOptions(cosmosevmkeyring.Option()) sets eth_secp256k1 as the default
       keyring signing algorithm so that `keys add` without --algo still uses the
       EVM-compatible key type.
       The _ EVMCryptoConfig parameter is intentionally blank — it exists solely to
       create a depinject dependency edge that forces ProvideCodecWithEVMCrypto to run
       before this function, guaranteeing ethsecp256k1 is registered before the context
       is handed to any command.
*/
// ProvideClientContext creates and provides a fully initialized client.Context,
// allowing it to be used for dependency injection and CLI operations.
func ProvideClientContext(
	appCodec codec.Codec,
	interfaceRegistry codectypes.InterfaceRegistry,
	txConfigOpts tx.ConfigOptions,
	legacyAmino *codec.LegacyAmino,
	_ EVMCryptoConfig, // @fix: force dependency edge — ensures EVM crypto is registered first
) client.Context {
	clientCtx := client.Context{}.
		WithCodec(appCodec).
		WithInterfaceRegistry(interfaceRegistry).
		WithLegacyAmino(legacyAmino).
		WithInput(os.Stdin).
		WithAccountRetriever(types.AccountRetriever{}).
		WithHomeDir(app.DefaultNodeHome).
		WithViper(app.Name).WithKeyringOptions(cosmosevmkeyring.Option()).WithLedgerHasProtobuf(true)

	// Read the config again to overwrite the default values with the values from the config file
	clientCtx, _ = config.ReadFromClientConfig(clientCtx)

	// textual is enabled by default, we need to re-create the tx config grpc instead of bank keeper.
	txConfigOpts.TextualCoinMetadataQueryFn = authtxconfig.NewGRPCCoinMetadataQueryFn(clientCtx)
	txConfig, err := tx.NewTxConfigWithOptions(clientCtx.Codec, txConfigOpts)
	if err != nil {
		panic(err)
	}
	clientCtx = clientCtx.WithTxConfig(txConfig)

	return clientCtx
}
