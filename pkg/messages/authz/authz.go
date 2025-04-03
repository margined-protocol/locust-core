package authz

import (
	"context"
	"fmt"
	"time"

	"github.com/cenkalti/backoff/v4"
	locustbackoff "github.com/margined-protocol/locust-core/pkg/backoff"
	"go.uber.org/zap"

	"github.com/cosmos/cosmos-sdk/x/authz"
)

const (
	MsgCreatePosition   = "/osmosis.concentratedliquidity.v1beta1.MsgCreatePosition"
	MsgWithdrawPosition = "/osmosis.concentratedliquidity.v1beta1.MsgWithdrawPosition"
	MsgPlaceLimitOrder  = "/neutron.dex.MsgPlaceLimitOrder"
	MsgDeposit          = "/neutron.dex.MsgDeposit"
	MsgWithdrawal       = "/neutron.dex.MsgWithdrawal"
)

var NeutronGrants = []string{
	MsgPlaceLimitOrder,
	MsgDeposit,
	MsgWithdrawal,
}

var OsmosisGrants = []string{
	MsgCreatePosition,
	MsgWithdrawPosition,
}

func GetValidGrantersNeutron(ctx context.Context, client authz.QueryClient, address string, l *zap.Logger) ([]string, error) {
	return GetValidGrantersWithRequiredGrants(ctx, client, NeutronGrants, address, l)
}

func GetValidGrantersOsmosis(ctx context.Context, client authz.QueryClient, address string, l *zap.Logger) ([]string, error) {
	return GetValidGrantersWithRequiredGrants(ctx, client, OsmosisGrants, address, l)
}

func GetValidGrantersWithRequiredGrants(ctx context.Context, client authz.QueryClient, requiredGrants []string, address string, l *zap.Logger) ([]string, error) {
	req := authz.QueryGranteeGrantsRequest{
		Grantee: address,
	}
	var res *authz.QueryGranteeGrantsResponse
	var err error

	exponentialBackoff := locustbackoff.NewBackoff(ctx)

	retryableRequest := func() error {
		res, err = client.GranteeGrants(ctx, &req)
		return err
	}

	err = backoff.Retry(retryableRequest, exponentialBackoff)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch grants: %w", err)
	}

	granteeGrants := make(map[string]map[string]bool)
	for _, grant := range res.Grants {
		// Check grant is valid
		if grant.Expiration != nil && !grant.Expiration.After(time.Now()) {
			l.Error("Grant is expired or invalid",
				zap.String("granter", grant.Granter),
				zap.String("grantee", grant.Grantee),
				zap.String("expiration", grant.Expiration.String()))
			continue
		}

		// Check if Authorization data is present and non-empty
		if grant.Authorization == nil || len(grant.Authorization.Value) == 0 {
			l.Error("Authorization data is missing or empty")
			continue
		}

		// Check if the type URL matches the expected GenericAuthorization type
		if grant.Authorization.TypeUrl != "/cosmos.authz.v1beta1.GenericAuthorization" {
			// Skip this grant as its type URL does not match the expected type
			continue
		}

		// Since the type URL matches, proceed to unmarshal
		var typ authz.GenericAuthorization
		if err := typ.Unmarshal(grant.Authorization.Value); err != nil {
			// Log the error and continue with the next grant
			l.Error("Failed to unmarshal authorization data", zap.Error(err))
			continue
		}

		// Log the details of the grant with the successfully unmarshaled GenericAuthorization type
		l.Debug("Grant details",
			zap.String("granter", grant.Granter),
			zap.String("grantee", grant.Grantee),
			zap.String("type url", grant.Authorization.TypeUrl),
			zap.String("msg", typ.Msg),
		)

		if grant.Expiration != nil {
			l.Debug("expiry date",
				zap.Time("expiry", *grant.Expiration),
			)
		}

		// Initialize the nested map for the grantee if it doesn't exist
		if granteeGrants[grant.Granter] == nil {
			granteeGrants[grant.Granter] = make(map[string]bool)
		}

		// Record the authorization msg as a granted permission
		granteeGrants[grant.Granter][typ.Msg] = true

	}

	var grantersWithAllRequiredGrants []string

	l.Debug("Grantee grants", zap.Any("grants", granteeGrants))

	// Determine which grantees have all required grants
	for granter, grants := range granteeGrants {
		l.Debug("Checking granter", zap.String("granter", granter))
		hasAllRequired := true
		for _, requiredGrant := range requiredGrants {
			if !grants[requiredGrant] {
				hasAllRequired = false

				l.Debug("Did not find required grant",
					zap.String("granter", granter),
				)
				break
			}
		}

		if hasAllRequired {
			l.Debug("Has all required", zap.String("granter", granter))
			grantersWithAllRequiredGrants = append(grantersWithAllRequiredGrants, granter)
		}
	}

	l.Debug("Granters with all required grants", zap.Any("grants", grantersWithAllRequiredGrants))

	return grantersWithAllRequiredGrants, nil
}
