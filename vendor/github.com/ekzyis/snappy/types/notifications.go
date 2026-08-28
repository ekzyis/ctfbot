package types

import "time"

type Notification struct {
	Id   int    `json:"id,string"`
	Type string `json:"__typename"`
	Item Item   `json:"item"`
}

type NotificationsCursor struct {
	LastChecked   time.Time      `json:"lastChecked"`
	Cursor        string         `json:"cursor"`
	Notifications []Notification `json:"notifications"`
}

type NotificationsResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		Notifications NotificationsCursor `json:"notifications"`
	} `json:"data"`
}
