package types

type GqlBody struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GqlError struct {
	Message string `json:"message"`
}
