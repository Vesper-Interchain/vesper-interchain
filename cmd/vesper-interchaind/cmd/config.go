package cmd

import (
	"strings"

	cmtcfg "github.com/cometbft/cometbft/config"
	serverconfig "github.com/cosmos/cosmos-sdk/server/config"
)

// initCometBFTConfig helps to override default CometBFT Config values.
// return cmtcfg.DefaultConfig if no custom configuration is required for the application.
func initCometBFTConfig() *cmtcfg.Config {
	cfg := cmtcfg.DefaultConfig()

	// these values put a higher strain on node memory
	// cfg.P2P.MaxNumInboundPeers = 100
	// cfg.P2P.MaxNumOutboundPeers = 40

	return cfg
}

// initAppConfig helps to override default appConfig template and configs.
// return "", nil if no custom configuration is required for the application.
func initAppConfig() (string, interface{}) {
	// The following code snippet is just for reference.
	type CustomAppConfig struct {
		serverconfig.Config `mapstructure:",squash"`
	}

	// Optionally allow the chain developer to overwrite the SDK's default
	// server config.
	srvCfg := serverconfig.DefaultConfig()
	// The SDK's default minimum gas price is set to "" (empty value) inside
	// app.toml. If left empty by validators, the node will halt on startup.
	// However, the chain developer can set a default app.toml value for their
	// validators here.
	//
	// In summary:
	// - if you leave srvCfg.MinGasPrices = "", all validators MUST tweak their
	//   own app.toml config,
	// - if you set srvCfg.MinGasPrices non-empty, validators CAN tweak their
	//   own app.toml to override, or use this default value.
	//
	// Set a default so the node can start without manual app.toml edits.
	// Validators can override this in their own app.toml.
	// "0stake" lets the feemarket module govern dynamic fees while satisfying
	// the SDK's requirement that MinGasPrices is non-empty.
	srvCfg.MinGasPrices = "0stake"

	customAppConfig := CustomAppConfig{
		Config: *srvCfg,
	}

	// cosmos-sdk's interceptConfigs calls rootViper.Unmarshal(&customConfig) where customConfig
	// is typed as `any`. This causes viper to replace the struct with a bare map, zeroing out
	// MinGasPrices in the rendered template. Hardcoding the value in the template string is the
	// only reliable way to ensure app.toml is written with a non-empty minimum-gas-prices on
	// fresh init, avoiding the "set min gas price in app.toml" startup error.
	customAppTemplate := strings.Replace(
		serverconfig.DefaultConfigTemplate,
		`minimum-gas-prices = "{{ .BaseConfig.MinGasPrices }}"`,
		`minimum-gas-prices = "0stake"`,
		1,
	)

	return customAppTemplate, customAppConfig
}
