package types

import "time"

type Invoice struct {
	Id                int                    `json:"id,string"`
	Hash              string                 `json:"hash"`
	Hmac              string                 `json:"hmac"`
	Bolt11            string                 `json:"bolt11"`
	SatsRequested     int                    `json:"satsRequested"`
	SatsReceived      int                    `json:"satsReceived"`
	Cancelled         bool                   `json:"cancelled"`
	ConfirmedAt       time.Time              `json:"createdAt"`
	ExpiresAt         time.Time              `json:"expiresAt"`
	Nostr             map[string]interface{} `json:"nostr"`
	IsHeld            bool                   `json:"isHeld"`
	Comment           string                 `json:"comment"`
	Lud18Data         map[string]interface{} `json:"lud18Data"`
	ConfirmedPreimage string                 `json:"confirmedPreimage"`
	ActionState       string                 `json:"actionState"`
	ActionType        string                 `json:"actionType"`
}

type PaymentMethod string

const (
	PaymentMethodFeeCredits  PaymentMethod = "FEE_CREDIT"
	PaymentMethodOptimistic  PaymentMethod = "OPTIMISTIC"
	PaymentMethodPessimistic PaymentMethod = "PESSIMISTIC"
)

type CreateInvoiceArgs struct {
	Amount      int
	ExpireSecs  int
	HodlInvoice bool
}

type CreateInvoiceResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		CreateInvoice Invoice `json:"createInvoice"`
	} `json:"data"`
}
