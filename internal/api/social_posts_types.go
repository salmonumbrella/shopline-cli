package api

// --- Social Posts: Query Options ---

// SocialChannelPostsOptions are query parameters for getting channel posts.
type SocialChannelPostsOptions struct {
	PartyChannelID string `json:"partyChannelId,omitempty"`
	PageSize       int    `json:"pageSize,omitempty"`
	Type           string `json:"type,omitempty"`
	Since          string `json:"since,omitempty"`
	Before         string `json:"before,omitempty"`
	After          string `json:"after,omitempty"`
}

// SocialProductSearchOptions are query parameters for searching social products.
type SocialProductSearchOptions struct {
	Query       string   `json:"query,omitempty"`
	Page        int      `json:"page,omitempty"`
	PageSize    int      `json:"pageSize,omitempty"`
	SearchType  string   `json:"searchType,omitempty"`
	CategoryIDs []string `json:"categoryIds,omitempty"`
}

// SalesEventListOptions are query parameters for listing sales events.
type SalesEventListOptions struct {
	PageNum   int    `json:"pageNum,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
	SalesType string `json:"salesType,omitempty"`
}

// --- Social Posts: Request Bodies ---

// CreateSalesEventRequest is the body for creating a sales event.
type CreateSalesEventRequest struct {
	Platform     string   `json:"platform"`
	Type         int      `json:"type"`
	Platforms    []string `json:"platforms"`
	Title        string   `json:"title"`
	PatternModel string   `json:"patternModel"`
	PostSubType  string   `json:"postSubType,omitempty"`
}

// ScheduleSalesEventRequest is the body for scheduling a sales event.
type ScheduleSalesEventRequest struct {
	StartTime int64 `json:"start_time"`
	EndTime   int64 `json:"end_time"`
}

// SalesEventSPU describes a product (SPU) to add to a sales event.
type SalesEventSPU struct {
	SPUID         string          `json:"spuId"`
	DefaultKey    string          `json:"defaultKey,omitempty"`
	MissCommonKey bool            `json:"missCommonKey,omitempty"`
	CustomNumbers []int           `json:"customNumbers,omitempty"`
	SKUList       []SalesEventSKU `json:"skuList"`
}

// SalesEventSKU describes a product variant (SKU) for a sales event.
type SalesEventSKU struct {
	SKUID         string   `json:"skuId"`
	MissCommonKey bool     `json:"missCommonKey,omitempty"`
	KeyList       []string `json:"keyList"`
}

// AddSalesEventProductsRequest is the body for adding products to a sales event.
type AddSalesEventProductsRequest struct {
	SPUList []SalesEventSPU `json:"spuList"`
}

// UpdateProductKeysRequest is the body for updating product keywords in a sales event.
type UpdateProductKeysRequest struct {
	SPUList []SalesEventSPU `json:"spuList"`
}

// SalesEventPost describes a social media post to link to a sales event.
type SalesEventPost struct {
	PostID          string `json:"postId"`
	PostTitle       string `json:"postTitle"`
	PostDescription string `json:"postDescription,omitempty"`
	PostImageURL    string `json:"postImageUrl,omitempty"`
	PermalinkURL    string `json:"permalinkUrl,omitempty"`
	StatusType      string `json:"statusType,omitempty"`
}

// LinkFacebookPostRequest is the body for linking a Facebook post to a sales event.
type LinkFacebookPostRequest struct {
	PageID   string           `json:"pageId"`
	PageName string           `json:"pageName"`
	PostList []SalesEventPost `json:"postList"`
}

// LinkInstagramPostRequest is the body for linking an Instagram post to a sales event.
type LinkInstagramPostRequest struct {
	PageID   string           `json:"pageId"`
	PageName string           `json:"pageName"`
	PostList []SalesEventPost `json:"postList"`
}

// LinkFBGroupPostRequest is the body for linking a Facebook Group post to a sales event.
type LinkFBGroupPostRequest struct {
	PageID      string `json:"pageId"`
	RelationURL string `json:"relationUrl"`
}
