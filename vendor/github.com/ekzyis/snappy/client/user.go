package client

import (
	"encoding/json"
	"fmt"

	t "github.com/ekzyis/snappy/types"
)

func (c *Client) Me() (*t.User, error) {
	body := t.GqlBody{
		Query: `
		query me {
			me {
				id
				name
				privates {
					sats
				}
			}
		}`,
	}

	resp, err := c.callApi(body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody t.MeResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding me: %w", err)
		return nil, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}
	return &respBody.Data.Me, nil
}
