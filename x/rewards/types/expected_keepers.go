package types

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the subset of the bank module keeper that the rewards
// module requires for minting and distributing reward tokens.
type BankKeeper interface {
	// MintCoins creates new coins in the rewards module account. The rewards
	// module mints uvusd directly rather than drawing from a pre-funded pool.
	MintCoins(ctx context.Context, moduleName string, amounts sdk.Coins) error

	// SendCoinsFromModuleToAccount transfers the freshly minted reward tokens
	// from the module account to the claimant's wallet.
	SendCoinsFromModuleToAccount(ctx context.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error

	// GetBalance is available for query-side balance checks (e.g. verifying
	// that the module account holds enough tokens before a transfer).
	GetBalance(ctx context.Context, addr sdk.AccAddress, denom string) sdk.Coin
}
