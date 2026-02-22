package api

// AdminCommentRequest is the body for adding a comment to an order.
type AdminCommentRequest struct {
	Comment   string `json:"comment"`
	IsPrivate bool   `json:"isPrivate,omitempty"`
}

// AdminOrderComment represents a single comment on an order from the Admin API.
type AdminOrderComment struct {
	ID        string `json:"id"`
	Comment   string `json:"comment"`
	IsPrivate bool   `json:"isPrivate"`
	Author    string `json:"author"`
	CreatedAt string `json:"created_at"`
}

// AdminRefundRequest is the body for an admin refund on an order.
type AdminRefundRequest struct {
	PerformerID           string `json:"performer_id"`
	Amount                int    `json:"amount"`
	OrderPaymentUpdatedAt string `json:"order_payment_updated_at"`
	RefundRemark          string `json:"refund_remark,omitempty"`
}

// AdminShipmentStatus is the response from GET /orders/{id}/shipment/status.
type AdminShipmentStatus struct {
	OrderNumber      string `json:"order_number"`
	Executed         bool   `json:"executed"`
	DeliveryPlatform string `json:"delivery_platform"`
	Shipment         string `json:"shipment"`
	TrackingNumber   string `json:"tracking_number,omitempty"`
}

// AdminTrackingResponse is the response from tracking-related endpoints.
type AdminTrackingResponse struct {
	TrackingNumber string `json:"tracking_number"`
}

// AdminExecuteShipmentRequest is the body for executing a shipment.
type AdminExecuteShipmentRequest struct {
	OrderNumber string `json:"orderNumber"`
	PerformerID string `json:"performerId"`
}

// AdminPrintLabelRequest is the body for printing a packing label.
type AdminPrintLabelRequest struct {
	Upsert bool `json:"upsert,omitempty"`
}

// AdminPackingLabel is the response from the print label endpoint.
type AdminPackingLabel struct {
	PackingLabel   string `json:"packing_label"`
	DeliveryMethod string `json:"delivery_method"`
	OriginalURL    string `json:"original_url,omitempty"`
}

// AdminListStreamsOptions are query parameters for listing livestreams.
type AdminListStreamsOptions struct {
	PageNum   int    `json:"pageNum,omitempty"`
	PageSize  int    `json:"pageSize,omitempty"`
	SalesType string `json:"salesType,omitempty"`
}

// AdminCreateStreamRequest is the body for creating a livestream.
type AdminCreateStreamRequest struct {
	Title             string `json:"title"`
	SalesOwner        string `json:"salesOwner"`
	SalesDescription  string `json:"salesDescription"`
	StartDate         string `json:"start_date"`
	EndDate           string `json:"end_date"`
	LockInventoryTime string `json:"lock_inventory_time"`
	CheckoutTime      string `json:"checkout_time"`
	CheckoutMessage   string `json:"checkout_message"`
	Platform          string `json:"platform"`
	ImageServePath    string `json:"image_serve_path,omitempty"`
}

// AdminUpdateStreamRequest is the body for updating a livestream.
type AdminUpdateStreamRequest struct {
	PostSalesTitle            string `json:"post_sales_title,omitempty"`
	PostSalesOwner            string `json:"post_sales_owner,omitempty"`
	PostSalesDescription      string `json:"post_sales_description,omitempty"`
	CheckoutTime              string `json:"checkout_time,omitempty"`
	LockInventoryTime         string `json:"lock_inventory_time,omitempty"`
	ArchivedStreamVisibleTime string `json:"archived_stream_visible_time,omitempty"`
}

// AdminStreamProduct describes a product to add to a livestream.
type AdminStreamProduct struct {
	ProductID  string                 `json:"product_id"`
	Variations []AdminStreamVariation `json:"variations"`
}

// AdminStreamVariation describes a product variation for a livestream.
type AdminStreamVariation struct {
	VariationID string   `json:"variation_id"`
	CustomKeys  []string `json:"custom_keys"`
}

// AdminAddStreamProductsRequest is the body for adding products to a livestream.
type AdminAddStreamProductsRequest struct {
	Products []AdminStreamProduct `json:"products"`
}

// AdminRemoveStreamProductsRequest is the body for removing products from a livestream.
type AdminRemoveStreamProductsRequest struct {
	ProductIDs []string `json:"productIds"`
}

// AdminStartStreamRequest is the body for starting a livestream.
type AdminStartStreamRequest struct {
	Platform  string          `json:"platform"`
	VideoData *AdminVideoData `json:"videoData,omitempty"`
}

// AdminVideoData describes the video data for a livestream start.
type AdminVideoData struct {
	PostID       string `json:"postId"`
	LiveVideoID  string `json:"liveVideoId"`
	PageID       string `json:"pageId"`
	PageName     string `json:"pageName"`
	PermalinkURL string `json:"permalinkUrl"`
	Status       string `json:"status"`
}

// AdminListConversationsOptions are query parameters for listing conversations.
type AdminListConversationsOptions struct {
	Platform    string `json:"platform,omitempty"`
	PageNum     int    `json:"page_num,omitempty"`
	PageSize    int    `json:"page_size,omitempty"`
	StateFilter string `json:"state_filter,omitempty"`
	IsArchived  *bool  `json:"is_archived,omitempty"`
	SearchType  string `json:"search_type,omitempty"`
	Query       string `json:"query,omitempty"`
}

// AdminSendMessageRequest is the body for sending a message in a conversation.
// ConversationID is injected by AdminClient.SendMessage and should not be set by callers.
type AdminSendMessageRequest struct {
	Platform       string `json:"platform"`
	Type           string `json:"type"`
	Content        string `json:"content"`
	ConversationID string `json:"conversation_id"`
}

// AdminListInstantMessagesOptions are query parameters for listing instant message conversations.
type AdminListInstantMessagesOptions struct {
	Page               int      `json:"page,omitempty"`
	SearchType         string   `json:"search_type,omitempty"`
	Route              string   `json:"route,omitempty"`
	UnreadType         string   `json:"unread_type,omitempty"`
	PageSize           int      `json:"page_size,omitempty"`
	PartyChannelIDList []string `json:"party_channel_id_list,omitempty"`
}

// AdminInstantMessagesQuery are query parameters for paginating instant conversation messages.
type AdminInstantMessagesQuery struct {
	SearchType   string `json:"search_type"`
	UseMessageID string `json:"use_message_id"`
	CreateTime   string `json:"create_time"`
}

// AdminSendInstantMessageRequest is the body for sending instant messages.
type AdminSendInstantMessageRequest struct {
	ConversationID string `json:"conversation_id"`
	Content        string `json:"content"`
	SenderTypeEnum string `json:"sender_type_enum,omitempty"`
	MessageSource  string `json:"message_source,omitempty"`
}

// AdminToggleStreamProductRequest is the body for toggling livestream product display status.
type AdminToggleStreamProductRequest struct {
	Status string `json:"status"`
}

// AdminExpressLinkProduct represents a product in express link creation.
type AdminExpressLinkProduct struct {
	ID          string `json:"_id"`
	VariationID string `json:"variation_id"`
}

// AdminExpressLinkCampaign represents a campaign in express link creation.
type AdminExpressLinkCampaign struct {
	ID string `json:"_id"`
}

// AdminCreateExpressLinkRequest is the request body for creating an express link.
type AdminCreateExpressLinkRequest struct {
	Products []AdminExpressLinkProduct `json:"products"`
	UserID   string                    `json:"user_id"`
	Campaign AdminExpressLinkCampaign  `json:"campaign"`
}

// AdminPaymentsPayoutsOptions are query parameters for payment payouts.
type AdminPaymentsPayoutsOptions struct {
	From int64 `json:"from"`
}

// AdminShoplyticsNewReturningOptions are query parameters for new/returning customers.
type AdminShoplyticsNewReturningOptions struct {
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}
