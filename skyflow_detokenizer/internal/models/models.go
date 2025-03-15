// models/models.go

package models

// TokenObject represents a single token in the request
// This matches the SQL Server JSON structure where each token
// is an object with a "token" field
type TokenObject struct {
	Token string `json:"token"`
}

// DetokenizeRequest represents the complete request structure
// coming from SQL Server. It contains an array of TokenObjects
// and a redaction type
type DetokenizeRequest struct {
	Tokens    []TokenObject `json:"tokens"`    // Array of token objects
	Redaction string        `json:"redaction"` // Redaction type (e.g., "PLAIN_TEXT")
}

// SkyflowRequest represents the structure we send to Skyflow API
// We transform our internal request into this format
type SkyflowRequest struct {
	DetokenizationParameters []struct {
		Token     string `json:"token"`
		Redaction string `json:"redaction"`
	} `json:"detokenizationParameters"`
	ContinueOnError bool `json:"continueOnError"` // Continue processing on error
}

// SkyflowResponse represents the response we get back from Skyflow API
type SkyflowResponse struct {
	Records []struct {
		Token     string  `json:"token"`
		Value     string  `json:"value"`
		Error     *string `json:"error"`
		ValueType string  `json:"valueType"`
	} `json:"records"`
}

// TokenResponse represents a single token-value pair in our response
// This is used when building our response back to SQL Server
type TokenResponse struct {
	Token string `json:"token"`
	Value string `json:"value"`
}
