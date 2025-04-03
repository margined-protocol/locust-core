package health

import (
	"encoding/json"
)

// AccountKind represents the various account types.
type AccountKind struct {
	Type        string          `json:"type"`
	FundManager *FundManagerMsg `json:"fund_manager,omitempty"`
}

// FundManagerMsg represents the details for the `FundManager` variant.
type FundManagerMsg struct {
	VaultAddr string `json:"vault_addr"`
}

// MarshalJSON implements custom JSON marshaling for AccountKind.
func (a AccountKind) MarshalJSON() ([]byte, error) {
	switch a.Type {
	case "FundManager":
		return json.Marshal(&struct {
			Type        string          `json:"type"`
			FundManager *FundManagerMsg `json:"fund_manager,omitempty"`
		}{
			Type:        a.Type,
			FundManager: a.FundManager,
		})
	default:
		// For Default and HighLeveredStrategy
		return json.Marshal(&struct {
			Type string `json:"type"`
		}{
			Type: a.Type,
		})
	}
}

// UnmarshalJSON implements custom JSON unmarshaling for AccountKind.
func (a *AccountKind) UnmarshalJSON(data []byte) error {
	// Define a struct for generic unmarshaling
	var raw struct {
		Type        string          `json:"type"`
		FundManager *FundManagerMsg `json:"fund_manager,omitempty"`
	}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// Populate the AccountKind based on the type
	a.Type = raw.Type
	switch raw.Type {
	case "FundManager":
		a.FundManager = raw.FundManager
	default:
		// Default and HighLeveredStrategy have no additional data
		a.FundManager = nil
	}

	return nil
}
