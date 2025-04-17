package ibc

import (
	"fmt"
)

// IBCConnectionRegistry manages IBC connections between chains
type IBCConnectionRegistry struct {
	// Map of source chain ID -> destination chain ID -> connection
	connections map[string]map[string]*IBCConnection
}

// NewIBCConnectionRegistry creates a new registry
func NewIBCConnectionRegistry() *IBCConnectionRegistry {
	return &IBCConnectionRegistry{
		connections: make(map[string]map[string]*IBCConnection),
	}
}

// DefaultIBCConnectionRegistry creates a new registry
func DefaultIBCConnectionRegistry() *IBCConnectionRegistry {
	dydxToOsmosis := &IBCConnection{
		Transfer: &IBCTransfer{
			SourceChainID: "dydx-mainnet-1",
			DestChainID:   "osmosis-1",
			Channel:       "channel-0",
			Port:          "transfer",
			Forward: &Forward{
				ChainID: "noble-1",
				Port:    "transfer",
				Channel: "channel-1",
			},
		},
		SourcePrefix:  "dydx",
		DestPrefix:    "neutron",
		ForwardPrefix: "noble",
	}

	dydxToNeutron := &IBCConnection{
		Transfer: &IBCTransfer{
			SourceChainID: "dydx-mainnet-1",
			DestChainID:   "neutron-1",
			Channel:       "channel-0",
			Port:          "transfer",
			Forward: &Forward{
				ChainID: "noble-1",
				Port:    "transfer",
				Channel: "channel-18",
			},
		},
		SourcePrefix:  "dydx",
		DestPrefix:    "neutron",
		ForwardPrefix: "noble",
	}

	neutronToOsmosis := &IBCConnection{
		Transfer: &IBCTransfer{
			SourceChainID: "neutron-1",
			DestChainID:   "osmosis-1",
			Channel:       "channel-30",
			Port:          "transfer",
			Forward: &Forward{
				ChainID: "noble-1",
				Port:    "transfer",
				Channel: "channel-1",
			},
		},
		SourcePrefix:  "neutron",
		DestPrefix:    "osmo",
		ForwardPrefix: "noble",
	}

	neutronToDydx := &IBCConnection{
		Transfer: &IBCTransfer{
			SourceChainID: "neutron-1",
			DestChainID:   "dydx-mainnet-1",
			Channel:       "channel-30",
			Port:          "transfer",
			Forward: &Forward{
				ChainID: "noble-1",
				Port:    "transfer",
				Channel: "channel-33",
			},
		},
		SourcePrefix:  "neutron",
		DestPrefix:    "dydx",
		ForwardPrefix: "noble",
	}

	neutronToUmee := &IBCConnection{
		Transfer: &IBCTransfer{
			SourceChainID: "neutron-1",
			DestChainID:   "umee-1",
			Channel:       "channel-30",
			Port:          "transfer",
			Forward: &Forward{
				ChainID: "noble-1",
				Port:    "transfer",
				Channel: "channel-33",
			},
		},
		SourcePrefix:  "neutron",
		DestPrefix:    "dydx",
		ForwardPrefix: "noble",
	}

	osmosisToDydx := &IBCConnection{
		Transfer: &IBCTransfer{
			SourceChainID: "osmosis-1",
			DestChainID:   "dydx-mainnet-1",
			Channel:       "channel-750",
			Port:          "transfer",
			Forward: &Forward{
				ChainID: "noble-1",
				Port:    "transfer",
				Channel: "channel-33",
			},
		},
		SourcePrefix:  "neutron",
		DestPrefix:    "dydx",
		ForwardPrefix: "noble",
	}

	osmosisToNeutron := &IBCConnection{
		Transfer: &IBCTransfer{
			SourceChainID: "osmosis-1",   // Mainnet chain ID
			DestChainID:   "neutron-1",   // Mainnet chain ID
			Channel:       "channel-750", // Noble
			Port:          "transfer",
			Forward: &Forward{
				ChainID: "noble-1",
				Port:    "transfer",
				Channel: "channel-18",
			},
		},
		SourcePrefix:  "osmo",
		DestPrefix:    "neutron",
		ForwardPrefix: "noble",
	}

	osmosisToUmee := &IBCConnection{
		Transfer: &IBCTransfer{
			SourceChainID: "osmosis-1",   // Mainnet chain ID
			DestChainID:   "umee-1",      // Mainnet chain ID
			Channel:       "channel-750", // Noble
			Port:          "transfer",
			Forward: &Forward{
				ChainID: "noble-1",
				Port:    "transfer",
				Channel: "channel-51",
			},
		},
		SourcePrefix:  "osmo",
		DestPrefix:    "umee",
		ForwardPrefix: "noble",
	}

	umeeToOsmosis := &IBCConnection{
		Transfer: &IBCTransfer{
			SourceChainID: "umee-1",
			DestChainID:   "osmosis-1",
			Channel:       "channel-120",
			Port:          "transfer",
			Forward: &Forward{
				ChainID: "noble-1",
				Port:    "transfer",
				Channel: "channel-1",
			},
		},
		SourcePrefix:  "umee",
		DestPrefix:    "osmo",
		ForwardPrefix: "noble",
	}

	umeeToNeutron := &IBCConnection{
		Transfer: &IBCTransfer{
			SourceChainID: "umee-1",
			DestChainID:   "osmosis-1",
			Channel:       "channel-51",
			Port:          "transfer",
			Forward: &Forward{
				ChainID: "noble-1",
				Port:    "transfer",
				Channel: "channel-1",
			},
		},
		SourcePrefix:  "umee",
		DestPrefix:    "osmo",
		ForwardPrefix: "noble",
	}

	return &IBCConnectionRegistry{
		connections: map[string]map[string]*IBCConnection{
			"osmosis-1": {
				"neutron-1":      osmosisToNeutron,
				"dydx-mainnet-1": osmosisToDydx,
				"umee-1":         osmosisToUmee,
			},
			"neutron-1": {
				"osmosis-1":      neutronToOsmosis,
				"dydx-mainnet-1": neutronToDydx,
				"umee-1":         neutronToUmee,
			},
			"dydx-mainnet-1": {
				"neutron-1": dydxToNeutron,
				"osmosis-1": dydxToOsmosis,
			},
			"umee-1": {
				"osmosis-1": umeeToOsmosis,
				"neutron-1": umeeToNeutron,
			},
		},
	}
}

// RegisterConnection adds a new IBC connection to the registry
func (r *IBCConnectionRegistry) RegisterConnection(connection *IBCConnection) error {
	sourceChainID := connection.Transfer.SourceChainID
	destChainID := connection.Transfer.DestChainID

	// Initialize the inner map if it doesn't exist
	if _, exists := r.connections[sourceChainID]; !exists {
		r.connections[sourceChainID] = make(map[string]*IBCConnection)
	}

	// Check if connection already exists
	if _, exists := r.connections[sourceChainID][destChainID]; exists {
		return fmt.Errorf("connection from %s to %s already exists", sourceChainID, destChainID)
	}

	// Store the connection
	r.connections[sourceChainID][destChainID] = connection
	return nil
}

// RegisterConnections adds multiple IBC connections to the registry
func (r *IBCConnectionRegistry) RegisterConnections(connections []*IBCConnection) error {
	for _, conn := range connections {
		if err := r.RegisterConnection(conn); err != nil {
			return fmt.Errorf("failed to register connection from %s to %s: %w",
				conn.Transfer.SourceChainID,
				conn.Transfer.DestChainID,
				err)
		}
	}
	return nil
}

// GetConnection retrieves an IBC connection from the registry
func (r *IBCConnectionRegistry) GetConnection(sourceChainID, destChainID string) (*IBCConnection, error) {
	// Check if source chain exists
	sourceMap, exists := r.connections[sourceChainID]
	if !exists {
		return nil, fmt.Errorf("no connections registered for source chain %s", sourceChainID)
	}

	// Check if destination chain exists
	connection, exists := sourceMap[destChainID]
	if !exists {
		return nil, fmt.Errorf("no connection registered from %s to %s", sourceChainID, destChainID)
	}

	return connection, nil
}

// GetAllConnections returns all registered connections
func (r *IBCConnectionRegistry) GetAllConnections() []*IBCConnection {
	var allConnections []*IBCConnection

	for _, destMap := range r.connections {
		for _, connection := range destMap {
			allConnections = append(allConnections, connection)
		}
	}

	return allConnections
}
