package ibc

import (
	"fmt"
)

// ConnectionRegistry manages IBC connections between chains
type ConnectionRegistry struct {
	// Map of source chain ID -> destination chain ID -> connection
	connections map[string]map[string]*Connection
}

// NewConnectionRegistry creates a new registry
func NewConnectionRegistry() *ConnectionRegistry {
	return &ConnectionRegistry{
		connections: make(map[string]map[string]*Connection),
	}
}

// DefaultConnectionRegistry creates a new registry
func DefaultConnectionRegistry() *ConnectionRegistry {
	osmosisToNeutron := &Connection{
		Transfer: &Transfer{
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

	neutronToOsmosis := &Connection{
		Transfer: &Transfer{
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

	dydxToNeutron := &Connection{
		Transfer: &Transfer{
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

	neutronToDydx := &Connection{
		Transfer: &Transfer{
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

	dydxToOsmosis := &Connection{
		Transfer: &Transfer{
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

	osmosisToDydx := &Connection{
		Transfer: &Transfer{
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

	return &ConnectionRegistry{
		connections: map[string]map[string]*Connection{
			"osmosis-1": {
				"neutron-1":      osmosisToNeutron,
				"dydx-mainnet-1": osmosisToDydx,
			},
			"neutron-1": {
				"osmosis-1":      neutronToOsmosis,
				"dydx-mainnet-1": neutronToDydx,
			},
			"dydx-mainnet-1": {
				"neutron-1": dydxToNeutron,
				"osmosis-1": dydxToOsmosis,
			},
		},
	}
}

// RegisterConnection adds a new IBC connection to the registry
func (r *ConnectionRegistry) RegisterConnection(connection *Connection) error {
	sourceChainID := connection.Transfer.SourceChainID
	destChainID := connection.Transfer.DestChainID

	// Initialize the inner map if it doesn't exist
	if _, exists := r.connections[sourceChainID]; !exists {
		r.connections[sourceChainID] = make(map[string]*Connection)
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
func (r *ConnectionRegistry) RegisterConnections(connections []*Connection) error {
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
func (r *ConnectionRegistry) GetConnection(sourceChainID, destChainID string) (*Connection, error) {
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
func (r *ConnectionRegistry) GetAllConnections() []*Connection {
	var allConnections []*Connection

	for _, destMap := range r.connections {
		for _, connection := range destMap {
			allConnections = append(allConnections, connection)
		}
	}

	return allConnections
}
