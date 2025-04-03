package time

import (
	"encoding/json"
	"strconv"
	"time"
)

// UnixNanoTime is a custom time type that marshals/unmarshals to/from a string
// containing the Unix time in nanoseconds.
type UnixNanoTime time.Time

// UnmarshalJSON converts a JSON string with Unix time in nanoseconds into a UnixNanoTime.
func (t *UnixNanoTime) UnmarshalJSON(data []byte) error {
	// Unmarshal the JSON data into a string.
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	// Parse the string as an int64 (nanoseconds).
	ns, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return err
	}
	*t = UnixNanoTime(time.Unix(0, ns))
	return nil
}

// MarshalJSON converts a UnixNanoTime into a JSON string in RFC3339 format.
func (t UnixNanoTime) MarshalJSON() ([]byte, error) {
	// Format the time in RFC3339 format.
	formattedTime := time.Time(t).Format(time.RFC3339)
	// Marshal the formatted time string as JSON.
	return json.Marshal(formattedTime)
}

// String returns the time in a human-readable format.
func (t UnixNanoTime) String() string {
	return time.Time(t).String()
}
