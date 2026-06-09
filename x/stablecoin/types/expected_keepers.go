package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AuthKeeper defines the expected auth keeper interface
type AuthKeeper interface {
	GetModuleAddress(moduleName string) sdk.AccAddress
}

// BankKeeper defines the expected bank keeper interface
type BankKeeper interface {
	// MintCoins creates new coins and adds them to the module account
	MintCoins(ctx context.Context, moduleName string, amount sdk.Coins) error

	// BurnCoins destroys coins from the module account
	BurnCoins(ctx context.Context, moduleName string, amount sdk.Coins) error

	// SendCoinsFromModuleToAccount sends coins from a module account to an account
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amount sdk.Coins) error

	// SendCoinsFromAccountToModule sends coins from an account to a module account
	SendCoinsFromAccountToModule(ctx context.Context, senderAddr sdk.AccAddress, recipientModule string, amount sdk.Coins) error
}
