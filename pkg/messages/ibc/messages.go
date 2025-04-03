package ibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	transfertypes "github.com/cosmos/ibc-go/v8/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v8/modules/core/02-client/types"
)

func CreateTransferMsg(port, channel, memo, sender, receiver string, token sdk.Coin, height, timeout uint64) sdk.Msg {
	return transfertypes.NewMsgTransfer(
		port,
		channel,
		token,
		sender, receiver,
		clienttypes.NewHeight(1, height+10),
		timeout,
		memo,
	)
}
