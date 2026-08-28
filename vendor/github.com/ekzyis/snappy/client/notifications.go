package client

import (
	"encoding/json"
	"fmt"

	t "github.com/ekzyis/snappy/types"
)

func (c *Client) Notifications() (*t.NotificationsCursor, error) {
	body := t.GqlBody{
		Query: `
		fragment ItemFields on Item {
			id
			user {
				id
				name
			}
			parentId
			createdAt
			deletedAt
			title
			text
		}
		query notifications {
			notifications {
				lastChecked
				cursor
				notifications {
					__typename
					... on Reply {
						id
						item {
							...ItemFields
						}
					}
					... on Mention {
						id
						item {
							...ItemFields
						}
					}
				}
			}
		}
		`,
		Variables: map[string]interface{}{},
	}

	resp, err := c.callApi(body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var respBody t.NotificationsResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		err = fmt.Errorf("error decoding notifications: %w", err)
		return nil, err
	}

	err = c.checkForErrors(respBody.Errors)
	if err != nil {
		return nil, err
	}
	return &respBody.Data.Notifications, nil
}

func (c *Client) Mentions() ([]t.Notification, error) {
	return c.filterNotifications(
		func(n t.Notification) bool {
			return n.Type == "Mention"
		},
	)
}

func (c *Client) Replies() ([]t.Notification, error) {
	return c.filterNotifications(
		func(n t.Notification) bool {
			return n.Type == "Reply"
		},
	)
}

func (c *Client) filterNotifications(f func(t.Notification) bool) ([]t.Notification, error) {
	var (
		n   *t.NotificationsCursor
		err error
	)

	if n, err = c.Notifications(); err != nil {
		return nil, err
	}

	return filter(n.Notifications, f), nil
}

func filter[T any](s []T, f func(T) bool) []T {
	var r []T
	for _, v := range s {
		if f(v) {
			r = append(r, v)
		}
	}
	return r
}
