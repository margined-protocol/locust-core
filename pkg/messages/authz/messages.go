package authz

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
)

// CreateAuthzMsg creates a wrapped MsgExec message
func CreateAuthzMsg(grantee sdk.AccAddress, msgs []sdk.Msg) sdk.Msg {
	// Generate the authz message using this helper since it wants type
	// Any messages so its easier like this
	msg := authz.NewMsgExec(grantee, msgs)

	return &authz.MsgExec{
		Grantee: msg.Grantee,
		Msgs:    msg.Msgs,
	}
}
