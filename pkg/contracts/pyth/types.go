package pyth

// APIResponse represents the structure of the JSON response from the API.
type APIResponse struct {
	Binary BinaryData `json:"binary"`
}

// BinaryData represents the "binary" field of the API response.
type BinaryData struct {
	Encoding string   `json:"encoding"`
	Data     []string `json:"data"`
}

// UpdatePriceFeeds feeds holds the payload for the update_price_feeds endpoint
type UpdatePriceFeeds struct {
	Data []string `json:"data"`
}

// UpdatePriceFeedsMsg is the wrapper for UpdatePriceFeeds
type UpdatePriceFeedsMsg struct {
	UpdatePriceFeeds UpdatePriceFeeds `json:"update_price_feeds"`
}
