package keeper

import (
	"bytes"
	"context"

	errorsmod "cosmossdk.io/errors"

	"github.com/Vesper-Interchain/vesper-interchain/x/oracle/types"
)

// UpdateParams processes a governance proposal to update oracle module parameters.
// The caller must be the module's designated authority (x/gov module account).
// This is the only mechanism by which the oracle address can be changed after genesis.
func (k msgServer) UpdateParams(ctx context.Context, req *types.MsgUpdateParams) (*types.MsgUpdateParamsResponse, error) {
	// Decode the request's authority string into raw bytes for comparison.
	authority, err := k.addressCodec.StringToBytes(req.Authority)
	if err != nil {
		return nil, errorsmod.Wrap(err, "invalid authority address")
	}

	// Ensure the signer is exactly the stored authority; any other signer is rejected.
	if !bytes.Equal(k.GetAuthority(), authority) {
		expectedAuthorityStr, _ := k.addressCodec.BytesToString(k.GetAuthority())
		return nil, errorsmod.Wrapf(types.ErrInvalidSigner, "invalid authority; expected %s, got %s", expectedAuthorityStr, req.Authority)
	}

	if err := req.Params.Validate(); err != nil {
		return nil, err
	}

	if err := k.Params.Set(ctx, req.Params); err != nil {
		return nil, err
	}

	return &types.MsgUpdateParamsResponse{}, nil
}
