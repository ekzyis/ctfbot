package client

import (
	"encoding/json"
	"fmt"

	t "github.com/ekzyis/snappy/types"
)

func (c *Client) Item(id int) (*t.Item, error) {
	body := t.GqlBody{
		Query: `
		query item($id: ID!) {
			item(id: $id) {
				id
				parentId
				title
				url
				text
				sats
				cost
				createdAt
				deletedAt
				ncomments
				subName
				subNames
				sub {
					name
					createdAt
					user {
						id
						name
					}
				}
				user {
					id
					name
				}
			}
		}`,
		Variables: map[string]interface{}{
			"id": id,
		},
	}

	resp, err := c.callApi(body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody t.ItemResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding item: %w", err)
		return nil, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}
	return &respBody.Data.Item, nil
}

func (c *Client) Items(query *t.ItemsQuery) (*t.ItemsCursor, error) {
	if query == nil {
		query = &t.ItemsQuery{}
	}

	body := t.GqlBody{
		Query: `
		query items($sub: String, $sort: String, $cursor: String, $type: String, $name: String, $when: String, $by: String, $limit: Limit) {
			items(sub: $sub, sort: $sort, cursor: $cursor, type: $type, name: $name, when: $when, by: $by, limit: $limit) {
				cursor
				items {
					id
					parentId
					title
					url
					text
					sats
					cost
					createdAt
					deletedAt
					ncomments
					subName
					subNames
					sub {
						name
						createdAt
						user {
							id
							name
						}
					}
					user {
						id
						name
					}
				},
			}
		}`,
		Variables: map[string]interface{}{
			"sub":    query.Sub,
			"sort":   query.Sort,
			"type":   query.Type,
			"cursor": query.Cursor,
			"name":   query.Name,
			"when":   query.When,
			"by":     query.By,
			"limit":  query.Limit,
		},
	}
	if query.Limit == 0 {
		body.Variables["limit"] = 21
	}

	resp, err := c.callApi(body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody t.ItemsResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding items: %w", err)
		return nil, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}
	return &respBody.Data.Items, nil
}

func (c *Client) PostDiscussion(title string, text string, subNames []string) (int, error) {
	body := t.GqlBody{
		Query: `
		mutation upsertDiscussion($title: String!, $text: String, $subNames: [String!]) {
			upsertDiscussion(title: $title, text: $text, subNames: $subNames) {
				id
				payerPrivates {
					payInFailureReason
					payInBolt11 {
						id
					}
					result {
						... on Item {
							id
						}
					}
				}
			}
		}`,
		Variables: map[string]interface{}{
			"title":    title,
			"text":     text,
			"subNames": subNames,
		},
	}

	resp, err := c.callApi(body)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var respBody t.UpsertDiscussionResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding upsertDiscussion: %w", err)
		return -1, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return -1, err
	}

	payIn := respBody.Data.UpsertDiscussion
	err = c.checkForPayInErrors(payIn)
	if err != nil {
		return -1, err
	}

	return payIn.PayerPrivates.Result.Id, nil
}

func (c *Client) PostLink(url string, title string, text string, subNames []string) (int, error) {
	body := t.GqlBody{
		Query: `
		mutation upsertLink($url: String!, $title: String!, $text: String, $subNames: [String!]) {
			upsertLink(url: $url, title: $title, text: $text, subNames: $subNames) {
				id
				payerPrivates {
					payInFailureReason
					payInBolt11 {
						id
					}
					result {
						... on Item {
							id
						}
					}
				}
			}
		}`,
		Variables: map[string]interface{}{
			"url":      url,
			"title":    title,
			"text":     text,
			"subNames": subNames,
		},
	}

	resp, err := c.callApi(body)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var respBody t.UpsertLinkResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding upsertLink: %w", err)
		return -1, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return -1, err
	}

	payIn := respBody.Data.UpsertLink
	err = c.checkForPayInErrors(payIn)
	if err != nil {
		return -1, err
	}

	return payIn.PayerPrivates.Result.Id, nil
}

func (c *Client) CreateComment(parentId int, text string) (int, error) {
	body := t.GqlBody{
		Query: `
		mutation upsertComment($parentId: ID!, $text: String!) {
			upsertComment(parentId: $parentId, text: $text) {
				id
				payerPrivates {
					payInFailureReason
					payInBolt11 {
						id
					}
					result {
						... on Item {
							id
						}
					}
				}
			}
		}`,
		Variables: map[string]interface{}{
			"parentId": parentId,
			"text":     text,
		},
	}

	resp, err := c.callApi(body)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	var respBody t.UpsertCommentResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding upsertComment: %w", err)
		return -1, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return -1, err
	}

	payIn := respBody.Data.UpsertComment
	err = c.checkForPayInErrors(payIn)
	if err != nil {
		return -1, err
	}

	return payIn.PayerPrivates.Result.Id, nil
}

func (c *Client) Dupes(url string) (*[]t.Dupe, error) {
	body := t.GqlBody{
		Query: `
		query Dupes($url: String!) {
			dupes(url: $url) {
				id
				url
				title
				user {
					name
				}
				createdAt
				sats
				ncomments
			}
		}`,
		Variables: map[string]interface{}{
			"url": url,
		},
	}
	resp, err := c.callApi(body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody t.DupesResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding dupes: %w", err)
		return nil, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}

	return &respBody.Data.Dupes, nil
}

func (c *Client) HasDupes(url string) (bool, error) {
	dupes, err := c.Dupes(url)
	if err != nil {
		return false, err
	}

	return len(*dupes) > 0, nil
}
