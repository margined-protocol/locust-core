package connection

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/zap"

	"github.com/margined-protocol/locust-core/pkg/types"
)

func TestAddressPrefixCaching(t *testing.T) {
	ctx := context.Background()
	logger := zap.NewExample()

	// Simple chain configs
	dydxChain := &types.Chain{
		RPCServerAddress: "https://dydx-rpc.publicnode.com:443",
		ChainID:          "dydx-1",
		Prefix:           "dydx",
		Fees:             ptr("15000000000adydx"),
		GasPrices:        ptr("15000000000adydx"),
	}

	neutronChain := &types.Chain{
		RPCServerAddress: "https://rpc-neutron.margined.io:443",
		ChainID:          "neutron-1",
		Prefix:           "neutron",
		Fees:             ptr("0.0053untrn"),
		GasPrices:        ptr("0.0053untrn"),
	}

	// Simple test key
	testKey := &types.SigningKey{
		AppName: "neutron",
		Backend: "os",
		RootDir: "/home/margined/.neutrond",
	}

	t.Run("demonstrate prefix caching", func(t *testing.T) {
		// Create dYdX client
		// _, err := InitCosmosClient(ctx, logger, dydxChain, testKey)
		dydxClient, err := InitCosmosClient(ctx, logger, dydxChain, testKey)
		require.NoError(t, err)

		// sdkConfig := sdk.GetConfig()
		// require.Equal(t, "dydx", sdkConfig.GetBech32AccountAddrPrefix())

		// Get dYdX address
		// dydxAccount, dydxAddr, err := GetSignerAccountAndAddress(dydxClient, "deployer", dydxChain.Prefix)
		dydxAccount, dydxAddr, err := GetSignerAccountAndAddress(dydxClient, "deployer", dydxChain.Prefix)
		require.NoError(t, err)
		require.Contains(t, dydxAddr, "dydx")

		address, err := dydxAccount.Record.GetAddress()
		fmt.Println("dydx address", address.Bytes())
		fmt.Println("dydx address", address.String())
		require.NoError(t, err)
		require.Contains(t, address.String(), "dydx")

		// Create Neutron client
		neutronClient, err := InitCosmosClient(ctx, logger, neutronChain, testKey)
		require.NoError(t, err)

		// sdkConfig = sdk.GetConfig()
		// require.Equal(t, "neutron", sdkConfig.GetBech32AccountAddrPrefix())

		// Get Neutron address - this is where the cache issue shows up
		neutronAccount, neutronAddr, err := GetSignerAccountAndAddress(neutronClient, "deployer", neutronChain.Prefix)
		require.NoError(t, err)
		require.Contains(t, neutronAddr, "neutron", "Address should have neutron prefix but cache might affect this")

		// // Verify we can get both address formats from the same account
		// dydxVerify, err := neutronAccount.Address(dydxChain.Prefix)
		// require.NoError(t, err)
		// require.Contains(t, dydxVerify, "dydx", "Should be able to get dYdX format")

		// neutronVerify, err := neutronAccount.Address(neutronChain.Prefix)
		// require.NoError(t, err)
		// require.Contains(t, neutronVerify, "neutron", "Should be able to get Neutron format")

		address, err = neutronAccount.Record.GetAddress()
		fmt.Println("neutron address", address.String())
		require.NoError(t, err)
		require.Contains(t, address.String(), "neutron")
	})
}

func ptr(s string) *string {
	return &s
}
