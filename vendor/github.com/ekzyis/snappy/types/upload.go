package types

type GetSignedPOST struct {
	Url    string            `json:"url"`
	Fields map[string]string `json:"fields"`
}

type GetSignedPOSTResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		GetSignedPOST GetSignedPOST `json:"getSignedPOST"`
	} `json:"data"`
}
