package creditmanager

import (
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/types"
)

type ActionAmount struct {
	Exact *string `json:"exact,omitempty"`
}

type ActionCoin struct {
	Denom  string       `json:"denom"`
	Amount ActionAmount `json:"amount"`
}

func (ac ActionCoin) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"denom":  ac.Denom,
		"amount": ac.Amount,
	})
}

func (ac *ActionCoin) UnmarshalJSON(data []byte) error {
	var raw struct {
		Denom  string       `json:"denom"`
		Amount ActionAmount `json:"amount"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	ac.Denom = raw.Denom
	ac.Amount = raw.Amount
	return nil
}

// AccountKind represents different types of accounts
type AccountKind struct {
	Type      string  `json:"type"`
	VaultAddr *string `json:"vault_addr,omitempty"`
}

// Predefined AccountKind values
var (
	Default             = AccountKind{Type: "Default"}
	HighLeveredStrategy = AccountKind{Type: "HighLeveredStrategy"}
)

// NewFundManagerAccount creates a FundManager account kind
func NewFundManagerAccount(vaultAddr string) AccountKind {
	return AccountKind{
		Type:      "FundManager",
		VaultAddr: &vaultAddr,
	}
}

// MarshalJSON custom serialization to match Rust's serde behavior
func (a AccountKind) MarshalJSON() ([]byte, error) {
	if a.Type == "FundManager" {
		return json.Marshal(map[string]interface{}{
			"FundManager": map[string]string{
				"vault_addr": *a.VaultAddr,
			},
		})
	}
	return json.Marshal(a.Type)
}

// UnmarshalJSON custom deserialization
func (a *AccountKind) UnmarshalJSON(data []byte) error {
	var typeStr string
	if err := json.Unmarshal(data, &typeStr); err == nil {
		a.Type = typeStr
		return nil
	}

	var fundManagerData map[string]map[string]string
	if err := json.Unmarshal(data, &fundManagerData); err == nil {
		if vault, ok := fundManagerData["FundManager"]["vault_addr"]; ok {
			a.Type = "FundManager"
			a.VaultAddr = &vault
		}
		return nil
	}

	return json.Unmarshal(data, a)
}

type VaultPositionType string

type LiquidateRequest struct {
	Type        string            `json:"type"`
	Deposit     *string           `json:"deposit,omitempty"`
	Lend        *string           `json:"lend,omitempty"`
	Vault       *VaultLiquidation `json:"vault,omitempty"`
	StakedAstro *string           `json:"staked_astro_lp,omitempty"`
}

type VaultLiquidation struct {
	RequestVault string            `json:"request_vault"`
	PositionType VaultPositionType `json:"position_type"`
}

func (lr LiquidateRequest) MarshalJSON() ([]byte, error) {
	switch lr.Type {
	case "Deposit":
		return json.Marshal(map[string]interface{}{
			"type":    "Deposit",
			"deposit": lr.Deposit,
		})
	case "Lend":
		return json.Marshal(map[string]interface{}{
			"type": "Lend",
			"lend": lr.Lend,
		})
	case "Vault":
		return json.Marshal(map[string]interface{}{
			"type":  "Vault",
			"vault": lr.Vault,
		})
	case "StakedAstro":
		return json.Marshal(map[string]interface{}{
			"type":            "StakedAstro",
			"staked_astro_lp": lr.StakedAstro,
		})
	default:
		return nil, fmt.Errorf("unknown LiquidateRequest type: %s", lr.Type)
	}
}

func (lr *LiquidateRequest) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if t, ok := raw["type"].(string); ok {
		lr.Type = t
	} else {
		return fmt.Errorf("missing or invalid type field in LiquidateRequest")
	}
	switch lr.Type {
	case "Deposit":
		if deposit, ok := raw["deposit"].(string); ok {
			lr.Deposit = &deposit
		} else {
			return fmt.Errorf("missing or invalid deposit field in LiquidateRequest")
		}
	case "Lend":
		if lend, ok := raw["lend"].(string); ok {
			lr.Lend = &lend
		} else {
			return fmt.Errorf("missing or invalid lend field in LiquidateRequest")
		}
	case "Vault":
		var vault VaultLiquidation
		if err := json.Unmarshal([]byte(raw["vault"].(string)), &vault); err != nil {
			return err
		}
		lr.Vault = &vault
	case "StakedAstro":
		if stakedAstro, ok := raw["staked_astro_lp"].(string); ok {
			lr.StakedAstro = &stakedAstro
		} else {
			return fmt.Errorf("missing or invalid staked_astro_lp field in LiquidateRequest")
		}
	default:
		return fmt.Errorf("unknown type field in LiquidateRequest: %s", lr.Type)
	}
	return nil
}

type Comparison string

const (
	GreaterThan Comparison = "GreaterThan"
	LessThan    Comparison = "LessThan"
)

type Condition struct {
	Type          string         `json:"type"`
	OraclePrice   *OraclePrice   `json:"oracle_price,omitempty"`
	RelativePrice *RelativePrice `json:"relative_price,omitempty"`
	HealthFactor  *HealthFactor  `json:"health_factor,omitempty"`
}

type OraclePrice struct {
	Denom      string     `json:"denom"`
	Price      string     `json:"price"`
	Comparison Comparison `json:"comparison"`
}

type RelativePrice struct {
	BasePriceDenom  string     `json:"base_price_denom"`
	QuotePriceDenom string     `json:"quote_price_denom"`
	Price           string     `json:"price"`
	Comparison      Comparison `json:"comparison"`
}

type HealthFactor struct {
	Threshold  string     `json:"threshold"`
	Comparison Comparison `json:"comparison"`
}

func (c Condition) MarshalJSON() ([]byte, error) {
	switch c.Type {
	case "OraclePrice":
		return json.Marshal(map[string]interface{}{
			"type":         "OraclePrice",
			"oracle_price": c.OraclePrice,
		})
	case "RelativePrice":
		return json.Marshal(map[string]interface{}{
			"type":           "RelativePrice",
			"relative_price": c.RelativePrice,
		})
	case "HealthFactor":
		return json.Marshal(map[string]interface{}{
			"type":          "HealthFactor",
			"health_factor": c.HealthFactor,
		})
	default:
		return nil, fmt.Errorf("unknown Condition type: %s", c.Type)
	}
}

func (c *Condition) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if t, ok := raw["type"].(string); ok {
		c.Type = t
	} else {
		return fmt.Errorf("missing or invalid type field in Condition")
	}
	switch c.Type {
	case "OraclePrice":
		var oracle OraclePrice
		if err := json.Unmarshal(data, &oracle); err != nil {
			return err
		}
		c.OraclePrice = &oracle
	case "RelativePrice":
		var relative RelativePrice
		if err := json.Unmarshal(data, &relative); err != nil {
			return err
		}
		c.RelativePrice = &relative
	case "HealthFactor":
		var health HealthFactor
		if err := json.Unmarshal(data, &health); err != nil {
			return err
		}
		c.HealthFactor = &health
	default:
		return fmt.Errorf("unknown Condition type: %s", c.Type)
	}
	return nil
}

// PositionsRequest represents the query payload for `Positions`.
type PositionsRequest struct {
	AccountID string `json:"account_id"`
}

// DebtAmount represents the debt details.
type DebtAmount struct {
	Denom  string `json:"denom"`
	Shares string `json:"shares"`
	Amount string `json:"amount"`
}

// PositionsResponse represents the result of the `Positions` query.
type PositionsResponse struct {
	AccountID      string          `json:"account_id"`
	AccountKind    string          `json:"account_kind"` // Assuming AccountKind as a string here
	Deposits       []types.Coin    `json:"deposits"`
	Debts          []DebtAmount    `json:"debts"`
	Lends          []types.Coin    `json:"lends"`
	Vaults         []VaultPosition `json:"vaults"`
	StakedAstroLPs []types.Coin    `json:"staked_astro_lps"`
}

// VaultPosition represents a position in a vault.
type VaultPosition struct {
	Denom   string        `json:"denom"`
	Deposit VaultDeposit  `json:"deposit"`
	Unlocks []VaultUnlock `json:"unlocks"`
}

// VaultDeposit represents the vault deposit details.
type VaultDeposit struct {
	Shares string `json:"shares"`
	Amount string `json:"amount"`
}

// VaultUnlock represents details of an unlocked vault.
type VaultUnlock struct {
	CreatedAt   uint64 `json:"created_at"`
	CooldownEnd uint64 `json:"cooldown_end"`
	Shares      string `json:"shares"`
	Amount      string `json:"amount"`
}

// Action represents the list of actions that users can perform on their positions
type Action struct {
	Deposit               *Coin         `json:"deposit,omitempty"`
	Withdraw              *ActionCoin   `json:"withdraw,omitempty"`
	WithdrawToWallet      *WithdrawData `json:"withdraw_to_wallet,omitempty"`
	Borrow                *Coin         `json:"borrow,omitempty"`
	Lend                  *ActionCoin   `json:"lend,omitempty"`
	Reclaim               *ActionCoin   `json:"reclaim,omitempty"`
	Repay                 *RepayData    `json:"repay,omitempty"`
	ExecutePerpOrder      *PerpOrder    `json:"execute_perp_order,omitempty"`
	CreateTriggerOrder    *TriggerOrder `json:"create_trigger_order,omitempty"`
	DeleteTriggerOrder    *string       `json:"delete_trigger_order,omitempty"`
	RefundAllCoinBalances bool          `json:"refund_all_coin_balances,omitempty"`
}

// WithdrawData represents the withdraw action to a wallet address
type WithdrawData struct {
	Coin      ActionCoin `json:"coin"`
	Recipient string     `json:"recipient"`
}

// RepayData represents a repayment action
type RepayData struct {
	RecipientAccountID *string    `json:"recipient_account_id,omitempty"`
	Coin               ActionCoin `json:"coin"`
}

// PerpOrder represents a perpetual position order execution
type PerpOrder struct {
	Denom      string  `json:"denom"`
	OrderSize  *string `json:"order_size"`
	ReduceOnly *bool   `json:"reduce_only,omitempty"`
}

// TriggerOrder represents the structure for creating a trigger order
type TriggerOrder struct {
	Actions    []Action    `json:"actions"`
	Conditions []Condition `json:"conditions"`
	KeeperFee  Coin        `json:"keeper_fee"`
}
