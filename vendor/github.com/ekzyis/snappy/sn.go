package sn

import (
	"github.com/ekzyis/snappy/client"
	t "github.com/ekzyis/snappy/types"
)

type Client = client.Client

var (
	NewClient    = client.NewClient
	WithApiKey   = client.WithApiKey
	WithNsec     = client.WithNsec
	WithBaseUrl  = client.WithBaseUrl
	WithMediaUrl = client.WithMediaUrl
)

type (
	GqlBody  = t.GqlBody
	GqlError = t.GqlError

	Item                     = t.Item
	Comment                  = t.Comment
	ItemsQuery               = t.ItemsQuery
	ItemsCursor              = t.ItemsCursor
	ItemResponse             = t.ItemResponse
	ItemsResponse            = t.ItemsResponse
	PayIn                    = t.PayIn
	UpsertDiscussionResponse = t.UpsertDiscussionResponse
	UpsertLinkResponse       = t.UpsertLinkResponse
	UpsertCommentResponse    = t.UpsertCommentResponse
	Dupe                     = t.Dupe
	DupesResponse            = t.DupesResponse
	DupesError               = t.DupesError

	Notification          = t.Notification
	NotificationsCursor   = t.NotificationsCursor
	NotificationsResponse = t.NotificationsResponse
	User                  = t.User
	UserPrivates          = t.UserPrivates
	MeResponse            = t.MeResponse
	Invoice               = t.Invoice
	PaymentMethod         = t.PaymentMethod
	CreateInvoiceArgs     = t.CreateInvoiceArgs
	CreateInvoiceResponse = t.CreateInvoiceResponse
	RssItem               = t.RssItem
	RssChannel            = t.RssChannel
	Rss                   = t.Rss
	RssDate               = t.RssDate
	RssAuthor             = t.RssAuthor
	GetSignedPOST         = t.GetSignedPOST
	GetSignedPOSTResponse = t.GetSignedPOSTResponse
)

const (
	PaymentMethodFeeCredits  = t.PaymentMethodFeeCredits
	PaymentMethodOptimistic  = t.PaymentMethodOptimistic
	PaymentMethodPessimistic = t.PaymentMethodPessimistic
)
