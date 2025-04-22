package ibc

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateForwardMemo(t *testing.T) {
	tests := []struct {
		name             string
		conn             *Transfer
		receiver         string
		destChainID      string
		expectedMemo     string
		expectedReceiver string
		shouldError      bool
		errorContains    string
	}{
		{
			name: "Direct transfer without forwarding",
			conn: &Transfer{
				SourceChainID: "osmosis-1",
				DestChainID:   "umee-1",
				Channel:       "channel-42",
				Port:          "transfer",
				// No Forward field
			},
			receiver:         "umee1p5ms3fycq6mcvylmp0asstppj8khvtf99eeypz",
			destChainID:      "umee-1",
			expectedMemo:     "",                                            // No memo for direct transfers
			expectedReceiver: "umee1p5ms3fycq6mcvylmp0asstppj8khvtf99eeypz", // Receiver should remain unchanged
			shouldError:      false,
		},
		{
			name: "Transfer with forwarding",
			conn: &Transfer{
				SourceChainID: "osmosis-1",
				DestChainID:   "umee-1",
				Channel:       "channel-120",
				Port:          "transfer",
				Forward: &Forward{
					ChainID:  "noble-1",
					Channel:  "channel-1",
					Port:     "transfer",
					Receiver: "umee1p5ms3fycq6mcvylmp0asstppj8khvtf99eeypz",
				},
			},
			receiver:         "neutron1p5ms3fycq6mcvylmp0asstppj8khvtf9nsdelh",
			destChainID:      "umee-1",
			expectedMemo:     `{"forward":{"channel":"channel-1","port":"transfer","receiver":"neutron1p5ms3fycq6mcvylmp0asstppj8khvtf9nsdelh"}}`,
			expectedReceiver: "umee1p5ms3fycq6mcvylmp0asstppj8khvtf99eeypz", // Should use Forward's receiver
			shouldError:      false,
		},
		{
			name: "Forward with empty receiver",
			conn: &Transfer{
				SourceChainID: "osmosis-1",
				DestChainID:   "umee-1",
				Channel:       "channel-120",
				Port:          "transfer",
				Forward: &Forward{
					ChainID:  "noble-1",
					Channel:  "channel-1",
					Port:     "transfer",
					Receiver: "", // Empty receiver in forward config
				},
			},
			receiver:         "neutron1p5ms3fycq6mcvylmp0asstppj8khvtf9nsdelh",
			destChainID:      "umee-1",
			expectedMemo:     "",
			expectedReceiver: "",
			shouldError:      true,
			errorContains:    "forward receiver address is empty",
		},
		{
			name:             "Nil connection",
			conn:             nil,
			receiver:         "umee1p5ms3fycq6mcvylmp0asstppj8khvtf99eeypz",
			destChainID:      "umee-1",
			expectedMemo:     "",
			expectedReceiver: "",
			shouldError:      true,
			errorContains:    "connection configuration is nil",
		},
		{
			name: "Real-world example from logs",
			conn: &Transfer{
				SourceChainID: "osmosis-1",
				DestChainID:   "umee-1",
				Channel:       "channel-120",
				Port:          "transfer",
				Forward: &Forward{
					ChainID:  "noble-1",
					Channel:  "channel-1",
					Port:     "transfer",
					Receiver: "noble1p5ms3fycq6mcvylmp0asstppj8khvtf9lv3na7",
				},
			},
			receiver:         "umee1p5ms3fycq6mcvylmp0asstppj8khvtf99eeypz",
			destChainID:      "umee-1",
			expectedMemo:     `{"forward":{"channel":"channel-1","port":"transfer","receiver":"umee1p5ms3fycq6mcvylmp0asstppj8khvtf99eeypz"}}`,
			expectedReceiver: "noble1p5ms3fycq6mcvylmp0asstppj8khvtf9lv3na7",
			shouldError:      false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			memo, receiver, err := CreateForwardMemo(tc.conn, tc.receiver)

			if tc.shouldError {
				require.Error(t, err)
				if tc.errorContains != "" {
					assert.Contains(t, err.Error(), tc.errorContains)
				}
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expectedReceiver, receiver)

			if tc.expectedMemo != "" {
				// Compare JSON objects for equality regardless of field order
				var expectedObj map[string]interface{}
				var actualObj map[string]interface{}

				err = json.Unmarshal([]byte(tc.expectedMemo), &expectedObj)
				require.NoError(t, err, "Expected memo is not valid JSON")

				err = json.Unmarshal([]byte(memo), &actualObj)
				require.NoError(t, err, "Actual memo is not valid JSON")

				assert.Equal(t, expectedObj, actualObj)
			} else {
				assert.Equal(t, tc.expectedMemo, memo)
			}
		})
	}
}

func TestForwardMemoMarshalJSON(t *testing.T) {
	tests := []struct {
		name         string
		forwardMemo  ForwardMemo
		expectedJSON string
	}{
		{
			name: "Basic forward memo",
			forwardMemo: ForwardMemo{
				Forward: ForwardInfo{
					Channel:  "channel-1",
					Port:     "transfer",
					Receiver: "cosmos1abc",
				},
			},
			expectedJSON: `{"forward":{"channel":"channel-1","port":"transfer","receiver":"cosmos1abc"}}`,
		},
		{
			name: "Forward memo with retries",
			forwardMemo: ForwardMemo{
				Forward: ForwardInfo{
					Channel:  "channel-20",
					Port:     "transfer",
					Receiver: "cosmos1xyz",
					Retries:  intPtr(3),
				},
			},
			expectedJSON: `{"forward":{"channel":"channel-20","port":"transfer","receiver":"cosmos1xyz","retries":3}}`,
		},
		{
			name: "Forward memo with timeout",
			forwardMemo: ForwardMemo{
				Forward: ForwardInfo{
					Channel:  "channel-5",
					Port:     "transfer",
					Receiver: "osmo1abc",
					Timeout:  int64Ptr(600),
				},
			},
			expectedJSON: `{"forward":{"channel":"channel-5","port":"transfer","receiver":"osmo1abc","timeout":600}}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			bytes, err := tc.forwardMemo.MarshalJSON()
			require.NoError(t, err)

			// Compare as JSON objects to avoid field order issues
			var expectedObj map[string]interface{}
			var actualObj map[string]interface{}

			err = json.Unmarshal([]byte(tc.expectedJSON), &expectedObj)
			require.NoError(t, err)

			err = json.Unmarshal(bytes, &actualObj)
			require.NoError(t, err)

			assert.Equal(t, expectedObj, actualObj)
		})
	}
}

// Helper functions to create pointers
func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}
