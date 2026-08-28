package types

import (
	"fmt"
	"time"

	"gopkg.in/guregu/null.v4"
)

type Item struct {
	Id        int       `json:"id,string"`
	ParentId  int       `json:"parentId"`
	Title     string    `json:"title"`
	Url       string    `json:"url"`
	Text      string    `json:"text"`
	Sats      int       `json:"sats"`
	Cost      int       `json:"cost"`
	CreatedAt time.Time `json:"createdAt"`
	DeletedAt null.Time `json:"deletedAt"`
	Comments  []Comment `json:"comments"`
	NComments int       `json:"ncomments"`
	User      User      `json:"user"`
	SubName string `json:"subName"`
	SubNames []string `json:"subNames"`
	Sub Sub `json:"sub"`
}

type Sub struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	User User `json:"user"`
}

type Comment struct {
	Id        int       `json:"id,string"`
	ParentId  int       `json:"parentId"`
	CreatedAt time.Time `json:"createdAt"`
	Text      string    `json:"text"`
	User      User      `json:"user"`
	Comments  []Comment `json:"comments"`
}

type ItemsQuery struct {
	Sub    string
	Sort   string
	Type   string
	Cursor string
	Name   string
	When   string
	By     string
	Limit  int
}

type ItemsCursor struct {
	Items  []Item `json:"items"`
	Cursor string `json:"cursor"`
}

type ItemResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		Item Item `json:"item"`
	} `json:"data"`
}

type ItemsResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		Items ItemsCursor `json:"items"`
	} `json:"data"`
}

type PayIn struct {
	Id            int `json:"id"`
	PayerPrivates struct {
		PayInFailureReason string `json:"payInFailureReason"`
		PayInBolt11        struct {
			Id int `json:"id"`
		} `json:"payInBolt11"`
		Result struct {
			Id int `json:"id,string"`
		} `json:"result"`
	} `json:"payerPrivates"`
}

type UpsertDiscussionResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		UpsertDiscussion PayIn `json:"upsertDiscussion"`
	} `json:"data"`
}

type UpsertLinkResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		UpsertLink PayIn `json:"upsertLink"`
	} `json:"data"`
}

type UpsertCommentResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		UpsertComment PayIn `json:"upsertComment"`
	} `json:"data"`
}

type Dupe struct {
	Id        int       `json:"id,string"`
	Url       string    `json:"url"`
	Title     string    `json:"title"`
	User      User      `json:"user"`
	CreatedAt time.Time `json:"createdAt"`
	Sats      int       `json:"sats"`
	NComments int       `json:"ncomments"`
}

type DupesResponse struct {
	Errors []GqlError `json:"errors"`
	Data   struct {
		Dupes []Dupe `json:"dupes"`
	} `json:"data"`
}

type DupesError struct {
	Url   string
	Dupes []Dupe
}

func (e *DupesError) Error() string {
	return fmt.Sprintf("found %d dupes for %s", len(e.Dupes), e.Url)
}
