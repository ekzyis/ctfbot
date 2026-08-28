package client

import (
	"encoding/json"
	"fmt"

	t "github.com/ekzyis/snappy/types"
)

func (c *Client) CreateInvoice(args *t.CreateInvoiceArgs) (*t.Invoice, error) {
	if args == nil {
		args = &t.CreateInvoiceArgs{}
	}

	body := t.GqlBody{
		// TODO: add createdAt
		//   when I wrote this code, createdAt returned null but is non-nullable
		//   so I had to remove it.
		Query: `
		mutation createInvoice($amount: Int!, $expireSecs: Int, $hodlInvoice: Boolean) {
			createInvoice(amount: $amount, expireSecs: $expireSecs, hodlInvoice: $hodlInvoice) {
				id
				hash
				hmac
				bolt11
				satsRequested
				satsReceived
				isHeld
				comment
				confirmedPreimage
				expiresAt
				confirmedAt
			}
		}`,
		Variables: map[string]interface{}{
			"amount": args.Amount,
		},
	}

	resp, err := c.callApi(body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody t.CreateInvoiceResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding items: %w", err)
		return nil, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}
	return &respBody.Data.CreateInvoice, nil
}
