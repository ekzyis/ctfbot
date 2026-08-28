package types

type User struct {
	Id       int          `json:"id,string"`
	Name     string       `json:"name"`
	Privates UserPrivates `json:"privates"`
}

type UserPrivates struct {
	Sats int `json:"sats"`
}

type MeResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		Me User `json:"me"`
	} `json:"data"`
}
