package connection

import (
	"github.com/ignite/cli/v28/ignite/pkg/cosmosaccount"
	"github.com/ignite/cli/v28/ignite/pkg/cosmosclient"
	"go.uber.org/zap"

	cometbft "github.com/cosmos/cosmos-sdk/client"

	"github.com/cometbft/cometbft/rpc/client/http"
)

// Define a function to get the signer account and address
func GetSignerAccountAndAddress(c *cosmosclient.Client, signer, prefix string) (*cosmosaccount.Account, string, error) {
	// Get the signer account
	account, err := c.Account(signer)
	if err != nil {
		return nil, "", err
	}

	// Get the signer address
	address, err := account.Address(prefix)
	if err != nil {
		return nil, "", err
	}

	return &account, address, nil
}

func GetGranteeAddress(c *cosmosclient.Client, signer, prefix string) (string, error) {
	_, granteeAddress, err := GetSignerAccountAndAddress(c, signer, prefix)
	if err != nil {
		return "", err
	}
	return granteeAddress, nil
}

func GetGranteeAccount(c *cosmosclient.Client, signer, prefix string) (cosmosaccount.Account, error) {
	granteeAccount, _, err := GetSignerAccountAndAddress(c, signer, prefix)
	if err != nil {
		return cosmosaccount.Account{}, err
	}
	return *granteeAccount, nil
}

func GetTendermintClient(l *zap.Logger, rpchttp *http.HTTP) cometbft.CometRPC {
	var tmClient cometbft.CometRPC
	if tc, ok := interface{}(rpchttp).(cometbft.CometRPC); ok {
		tmClient = tc
	} else {
		l.Error("Client does not implement CometRPC interface")
		return nil
	}

	return tmClient
}
