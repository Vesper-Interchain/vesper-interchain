package types

import (
	"context"

	"cosmossdk.io/core/address"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AuthKeeper defines the subset of the authentication keeper that the oracle
// module requires. Keeping this as a narrow interface avoids a full dependency
// on the auth module and makes the oracle keeper easier to test in isolation.
type AuthKeeper interface {
	AddressCodec() address.Codec
	// GetAccount is used only during simulation to validate account existence.
	GetAccount(context.Context, sdk.AccAddress) sdk.AccountI
}

// BankKeeper defines the subset of the bank keeper that the oracle module
// requires. Currently only used for simulation (checking spendable balances).
type BankKeeper interface {
	SpendableCoins(context.Context, sdk.AccAddress) sdk.Coins
}

// ParamSubspace defines the expected interface for reading and writing module
// parameters via the legacy params subspace. Retained for migration compatibility
// only; new parameter changes go through governance MsgUpdateParams.
type ParamSubspace interface {
	Get(context.Context, []byte, interface{})
	Set(context.Context, []byte, interface{})
}
