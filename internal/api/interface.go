package api

import (
	"context"
	"encoding/json"
)

// APIClient defines the interface for Shopline API operations.
// This interface is implemented by *Client and can be mocked for testing.
type APIClient interface {
	AcceptDispute(ctx context.Context, id string) (*Dispute, error)
	ActivateCoupon(ctx context.Context, id string) (*Coupon, error)
	ActivateFlashPrice(ctx context.Context, id string) (*FlashPrice, error)
	ActivateGift(ctx context.Context, id string) (*Gift, error)
	ActivatePromotion(ctx context.Context, id string) (*Promotion, error)
	ActivateSale(ctx context.Context, id string) (*Sale, error)
	AddProductImages(ctx context.Context, productID string, req *ProductAddImagesRequest) ([]ProductImage, error)
	AddProductVariation(ctx context.Context, productID string, req *ProductVariationCreateRequest) (*ProductVariation, error)
	AddProductsToCollection(ctx context.Context, id string, productIDs []string) error
	AddWishListItem(ctx context.Context, wishListID string, req *WishListItemCreateRequest) (*WishListItem, error)
	AdjustCompanyCredit(ctx context.Context, id string, req *CompanyCreditAdjustRequest) (*CompanyCredit, error)
	AdjustInventory(ctx context.Context, id string, delta int) (*InventoryLevel, error)
	AdjustInventoryLevel(ctx context.Context, req *InventoryLevelAdjustRequest) (*InventoryLevel, error)
	AdjustMemberPoints(ctx context.Context, customerID string, points int, description string) (*MemberPoints, error)
	AssignUserCoupon(ctx context.Context, req *UserCouponAssignRequest) (*UserCoupon, error)
	ClaimUserCoupon(ctx context.Context, couponCode string, body any) (json.RawMessage, error)
	CancelBulkOperation(ctx context.Context, id string) (*BulkOperation, error)
	CancelFulfillmentOrder(ctx context.Context, id string) (*FulfillmentOrder, error)
	BulkExecuteShipment(ctx context.Context, orderIDs []string) (*BulkShipmentResponse, error)
	CancelOrder(ctx context.Context, id string) error
	CreateArchivedOrdersReport(ctx context.Context, body any) (json.RawMessage, error)
	CreateOrder(ctx context.Context, req *OrderCreateRequest) (*Order, error)
	CancelPurchaseOrder(ctx context.Context, id string) (*PurchaseOrder, error)
	CancelReturnOrder(ctx context.Context, id string) error
	CapturePayment(ctx context.Context, id string, amount string) (*Payment, error)
	CloseFulfillmentOrder(ctx context.Context, id string) (*FulfillmentOrder, error)
	CompleteDraftOrder(ctx context.Context, id string) (*DraftOrder, error)
	CompleteReturnOrder(ctx context.Context, id string) (*ReturnOrder, error)
	CreateAddonProduct(ctx context.Context, req *AddonProductCreateRequest) (*AddonProduct, error)
	CreateAffiliateCampaign(ctx context.Context, req *AffiliateCampaignCreateRequest) (*AffiliateCampaign, error)
	ExportAffiliateCampaignReport(ctx context.Context, id string, body any) (json.RawMessage, error)
	CreateArticle(ctx context.Context, req *ArticleCreateRequest) (*Article, error)
	CreateBlog(ctx context.Context, req *BlogCreateRequest) (*Blog, error)
	CreateBulkMutation(ctx context.Context, req *BulkOperationMutationRequest) (*BulkOperation, error)
	CreateBulkQuery(ctx context.Context, req *BulkOperationCreateRequest) (*BulkOperation, error)
	CreateCarrierService(ctx context.Context, req *CarrierServiceCreateRequest) (*CarrierService, error)
	CreateCatalogPricing(ctx context.Context, req *CatalogPricingCreateRequest) (*CatalogPricing, error)
	CreateCategory(ctx context.Context, req *CategoryCreateRequest) (*Category, error)
	CreateChannel(ctx context.Context, req *ChannelCreateRequest) (*Channel, error)
	CreateChannelProductPrice(ctx context.Context, channelID, productID string, body any) (json.RawMessage, error)
	CreateCollection(ctx context.Context, req *CollectionCreateRequest) (*Collection, error)
	CreateCompanyCatalog(ctx context.Context, req *CompanyCatalogCreateRequest) (*CompanyCatalog, error)
	CreateCompanyCredit(ctx context.Context, req *CompanyCreditCreateRequest) (*CompanyCredit, error)
	CreateConversation(ctx context.Context, req *ConversationCreateRequest) (*Conversation, error)
	CreateConversationShopMessage(ctx context.Context, body any) (json.RawMessage, error)
	CreateCoupon(ctx context.Context, req *CouponCreateRequest) (*Coupon, error)
	CreateCustomer(ctx context.Context, req *CustomerCreateRequest) (*Customer, error)
	CreateCustomerAddress(ctx context.Context, customerID string, req *CustomerAddressCreateRequest) (*CustomerAddress, error)
	CreateCustomerBlacklist(ctx context.Context, req *CustomerBlacklistCreateRequest) (*CustomerBlacklist, error)
	CreateCustomerGroup(ctx context.Context, req *CustomerGroupCreateRequest) (*CustomerGroup, error)
	CreateCustomerSavedSearch(ctx context.Context, req *CustomerSavedSearchCreateRequest) (*CustomerSavedSearch, error)
	CreateCustomField(ctx context.Context, req *CustomFieldCreateRequest) (*CustomField, error)
	CreateDiscountCode(ctx context.Context, req *DiscountCodeCreateRequest) (*DiscountCode, error)
	CreateDomain(ctx context.Context, req *DomainCreateRequest) (*Domain, error)
	CreateDraftOrder(ctx context.Context, req *DraftOrderCreateRequest) (*DraftOrder, error)
	CreateFile(ctx context.Context, req *FileCreateRequest) (*File, error)
	CreateFlashPrice(ctx context.Context, req *FlashPriceCreateRequest) (*FlashPrice, error)
	CreateFlashPriceCampaign(ctx context.Context, body any) (json.RawMessage, error)
	CreateFulfillmentService(ctx context.Context, req *FulfillmentServiceCreateRequest) (*FulfillmentService, error)
	CreateGift(ctx context.Context, req *GiftCreateRequest) (*Gift, error)
	CreateGiftCard(ctx context.Context, req *GiftCardCreateRequest) (*GiftCard, error)
	CreateLabel(ctx context.Context, req *LabelCreateRequest) (*Label, error)
	CreateLocalDeliveryOption(ctx context.Context, req *LocalDeliveryCreateRequest) (*LocalDeliveryOption, error)
	CreateLocation(ctx context.Context, req *LocationCreateRequest) (*Location, error)
	CreateMarket(ctx context.Context, req *MarketCreateRequest) (*Market, error)
	CreateMarketingEvent(ctx context.Context, req *MarketingEventCreateRequest) (*MarketingEvent, error)
	CreateMedia(ctx context.Context, req *MediaCreateRequest) (*Media, error)
	// Media (documented endpoint)
	CreateMediaImage(ctx context.Context, body any) (json.RawMessage, error)
	CreateMembershipTier(ctx context.Context, req *MembershipTierCreateRequest) (*MembershipTier, error)
	CreateMetafield(ctx context.Context, req *MetafieldCreateRequest) (*Metafield, error)
	CreateMetafieldDefinition(ctx context.Context, req *MetafieldDefinitionCreateRequest) (*MetafieldDefinition, error)
	CreateOrderRisk(ctx context.Context, orderID string, req *OrderRiskCreateRequest) (*OrderRisk, error)
	CreatePage(ctx context.Context, req *PageCreateRequest) (*Page, error)
	CreatePickupLocation(ctx context.Context, req *PickupCreateRequest) (*PickupLocation, error)
	CreatePriceRule(ctx context.Context, req *PriceRuleCreateRequest) (*PriceRule, error)
	CreateProduct(ctx context.Context, req *ProductCreateRequest) (*Product, error)
	CreateProductListing(ctx context.Context, productID string) (*ProductListing, error)
	CreateProductReview(ctx context.Context, req *ProductReviewCreateRequest) (*ProductReview, error)
	// Product review comments (documented endpoints)
	CreateProductReviewComment(ctx context.Context, body any) (json.RawMessage, error)
	BulkCreateProductReviewComments(ctx context.Context, body any) (json.RawMessage, error)
	CreateProductSubscription(ctx context.Context, req *ProductSubscriptionCreateRequest) (*ProductSubscription, error)
	CreatePromotion(ctx context.Context, req *PromotionCreateRequest) (*Promotion, error)
	CreatePurchaseOrder(ctx context.Context, req *PurchaseOrderCreateRequest) (*PurchaseOrder, error)
	// POS purchase orders (documented endpoints)
	CreatePOSPurchaseOrder(ctx context.Context, body any) (json.RawMessage, error)
	CreatePOSPurchaseOrderChild(ctx context.Context, id string, body any) (json.RawMessage, error)
	CreateRedirect(ctx context.Context, req *RedirectCreateRequest) (*Redirect, error)
	CreateRefund(ctx context.Context, req *RefundCreateRequest) (*Refund, error)
	CreateReturnOrder(ctx context.Context, req *ReturnOrderCreateRequest) (*ReturnOrder, error)
	CreateSale(ctx context.Context, req *SaleCreateRequest) (*Sale, error)
	DeleteSaleProducts(ctx context.Context, saleID string, req *SaleDeleteProductsRequest) error
	CreateScriptTag(ctx context.Context, req *ScriptTagCreateRequest) (*ScriptTag, error)
	CreateSellingPlan(ctx context.Context, req *SellingPlanCreateRequest) (*SellingPlan, error)
	CreateShipment(ctx context.Context, req *ShipmentCreateRequest) (*Shipment, error)
	CreateShippingZone(ctx context.Context, req *ShippingZoneCreateRequest) (*ShippingZone, error)
	CreateSizeChart(ctx context.Context, req *SizeChartCreateRequest) (*SizeChart, error)
	CreateSmartCollection(ctx context.Context, req *SmartCollectionCreateRequest) (*SmartCollection, error)
	// Carts (Open API, /carts/...)
	ExchangeCart(ctx context.Context, body any) (json.RawMessage, error)
	PrepareCart(ctx context.Context, cartID string, body any) (json.RawMessage, error)
	AddCartItems(ctx context.Context, cartID string, body any) (json.RawMessage, error)
	UpdateCartItems(ctx context.Context, cartID string, body any) (json.RawMessage, error)
	DeleteCartItems(ctx context.Context, cartID string, body any) (json.RawMessage, error)
	// Cart item metafields (Open API, /carts/{cart_id}/items/*_metafields)
	ListCartItemMetafields(ctx context.Context, cartID string) (json.RawMessage, error)
	BulkCreateCartItemMetafields(ctx context.Context, cartID string, body any) error
	BulkUpdateCartItemMetafields(ctx context.Context, cartID string, body any) error
	BulkDeleteCartItemMetafields(ctx context.Context, cartID string, body any) error
	ListCartItemAppMetafields(ctx context.Context, cartID string) (json.RawMessage, error)
	BulkCreateCartItemAppMetafields(ctx context.Context, cartID string, body any) error
	BulkUpdateCartItemAppMetafields(ctx context.Context, cartID string, body any) error
	BulkDeleteCartItemAppMetafields(ctx context.Context, cartID string, body any) error
	CreateStorefrontCart(ctx context.Context, req *StorefrontCartCreateRequest) (*StorefrontCart, error)
	CreateStorefrontOAuthClient(ctx context.Context, req *StorefrontOAuthClientCreateRequest) (*StorefrontOAuthClient, error)
	CreateStorefrontOAuthApplication(ctx context.Context, req *StorefrontOAuthApplicationCreateRequest) (*StorefrontOAuthApplication, error)
	CreateStorefrontToken(ctx context.Context, req *StorefrontTokenCreateRequest) (*StorefrontToken, error)
	CreateSubscription(ctx context.Context, req *SubscriptionCreateRequest) (*Subscription, error)
	CreateTag(ctx context.Context, req *TagCreateRequest) (*Tag, error)
	CreateTax(ctx context.Context, req *TaxCreateRequest) (*Tax, error)
	CreateTaxonomy(ctx context.Context, req *TaxonomyCreateRequest) (*Taxonomy, error)
	CreateTaxService(ctx context.Context, req *TaxServiceCreateRequest) (*TaxService, error)
	CreateTheme(ctx context.Context, req *ThemeCreateRequest) (*Theme, error)
	CreateToken(ctx context.Context, req *TokenCreateRequest) (*Token, error)
	CreateWarehouse(ctx context.Context, req *WarehouseCreateRequest) (*Warehouse, error)
	CreateWebhook(ctx context.Context, req *WebhookCreateRequest) (*Webhook, error)
	CreateWishList(ctx context.Context, req *WishListCreateRequest) (*WishList, error)
	CreateWishListItem(ctx context.Context, body any) (json.RawMessage, error)
	DeactivateCoupon(ctx context.Context, id string) (*Coupon, error)
	DeactivateFlashPrice(ctx context.Context, id string) (*FlashPrice, error)
	DeactivateGift(ctx context.Context, id string) (*Gift, error)
	DeactivatePromotion(ctx context.Context, id string) (*Promotion, error)
	DeactivateSale(ctx context.Context, id string) (*Sale, error)
	Delete(ctx context.Context, path string) error
	DeleteAddonProduct(ctx context.Context, id string) error
	DeleteAffiliateCampaign(ctx context.Context, id string) error
	DeleteArticle(ctx context.Context, id string) error
	DeleteAsset(ctx context.Context, themeID, key string) error
	DeleteBlog(ctx context.Context, id string) error
	DeleteCarrierService(ctx context.Context, id string) error
	DeleteCatalogPricing(ctx context.Context, id string) error
	DeleteCategory(ctx context.Context, id string) error
	DeleteChannel(ctx context.Context, id string) error
	DeleteCollection(ctx context.Context, id string) error
	DeleteCompanyCatalog(ctx context.Context, id string) error
	DeleteCompanyCredit(ctx context.Context, id string) error
	DeleteConversation(ctx context.Context, id string) error
	DeleteCoupon(ctx context.Context, id string) error
	DeleteCustomer(ctx context.Context, id string) error
	DeleteCustomerAddress(ctx context.Context, customerID, addressID string) error
	DeleteCustomerBlacklist(ctx context.Context, id string) error
	DeleteCustomerGroup(ctx context.Context, id string) error
	DeleteCustomerSavedSearch(ctx context.Context, id string) error
	DeleteCustomField(ctx context.Context, id string) error
	DeleteDiscountCode(ctx context.Context, id string) error
	DeleteDomain(ctx context.Context, id string) error
	DeleteDraftOrder(ctx context.Context, id string) error
	DeleteFile(ctx context.Context, id string) error
	DeleteFlashPrice(ctx context.Context, id string) error
	DeleteFlashPriceCampaign(ctx context.Context, id string) error
	DeleteFulfillmentService(ctx context.Context, id string) error
	DeleteGift(ctx context.Context, id string) error
	DeleteGiftCard(ctx context.Context, id string) error
	DeleteLabel(ctx context.Context, id string) error
	DeleteLocalDeliveryOption(ctx context.Context, id string) error
	DeleteLocation(ctx context.Context, id string) error
	DeleteMarket(ctx context.Context, id string) error
	DeleteMarketingEvent(ctx context.Context, id string) error
	DeleteMedia(ctx context.Context, id string) error
	DeleteMembershipTier(ctx context.Context, id string) error
	DeleteMetafield(ctx context.Context, id string) error
	DeleteMetafieldDefinition(ctx context.Context, id string) error
	DeleteOrderRisk(ctx context.Context, orderID, riskID string) error
	DeletePage(ctx context.Context, id string) error
	DeletePickupLocation(ctx context.Context, id string) error
	DeletePriceRule(ctx context.Context, id string) error
	DeleteProduct(ctx context.Context, id string) error
	DeleteProductImages(ctx context.Context, productID string, imageIDs []string) error
	DeleteProductListing(ctx context.Context, id string) error
	DeleteProductReview(ctx context.Context, id string) error
	DeleteProductReviewComment(ctx context.Context, id string) (json.RawMessage, error)
	DeleteProductSubscription(ctx context.Context, id string) error
	DeleteProductVariation(ctx context.Context, productID string, variationID string) error
	DeletePromotion(ctx context.Context, id string) error
	DeletePurchaseOrder(ctx context.Context, id string) error
	DeleteRedirect(ctx context.Context, id string) error
	DeleteSale(ctx context.Context, id string) error
	DeleteScriptTag(ctx context.Context, id string) error
	DeleteSellingPlan(ctx context.Context, id string) error
	DeleteShipment(ctx context.Context, id string) error
	DeleteShippingZone(ctx context.Context, id string) error
	DeleteSizeChart(ctx context.Context, id string) error
	DeleteSmartCollection(ctx context.Context, id string) error
	DeleteStaff(ctx context.Context, id string) error
	DeleteStorefrontCart(ctx context.Context, id string) error
	DeleteStorefrontOAuthClient(ctx context.Context, id string) error
	DeleteStorefrontOAuthApplication(ctx context.Context, id string) error
	DeleteStorefrontToken(ctx context.Context, id string) error
	DeleteSubscription(ctx context.Context, id string) error
	DeleteTag(ctx context.Context, id string) error
	DeleteTax(ctx context.Context, id string) error
	DeleteTaxonomy(ctx context.Context, id string) error
	DeleteTaxService(ctx context.Context, id string) error
	DeleteTheme(ctx context.Context, id string) error
	DeleteToken(ctx context.Context, id string) error
	DeleteWarehouse(ctx context.Context, id string) error
	DeleteWebhook(ctx context.Context, id string) error
	DeleteWishList(ctx context.Context, id string) error
	DeleteWishListItems(ctx context.Context, body any) (json.RawMessage, error)
	DisableMultipass(ctx context.Context) error
	EnableMultipass(ctx context.Context) (*Multipass, error)
	GenerateMultipassToken(ctx context.Context, req *MultipassTokenRequest) (*MultipassToken, error)
	Get(ctx context.Context, path string, result interface{}) error
	GetAbandonedCheckout(ctx context.Context, id string) (*AbandonedCheckout, error)
	GetAddonProduct(ctx context.Context, id string) (*AddonProduct, error)
	GetAddonProductStocks(ctx context.Context, id string) (*AddonProductStocksResponse, error)
	GetAffiliateCampaign(ctx context.Context, id string) (*AffiliateCampaign, error)
	GetAffiliateCampaignOrders(ctx context.Context, id string, opts *AffiliateCampaignOrdersOptions) (json.RawMessage, error)
	GetAffiliateCampaignProductsSalesRanking(ctx context.Context, id string, opts *AffiliateCampaignProductsSalesRankingOptions) (json.RawMessage, error)
	GetAffiliateCampaignSummary(ctx context.Context, id string) (json.RawMessage, error)
	GetArticle(ctx context.Context, id string) (*Article, error)
	GetAsset(ctx context.Context, themeID, key string) (*Asset, error)
	GetBalance(ctx context.Context) (*Balance, error)
	GetBalanceTransaction(ctx context.Context, id string) (*BalanceTransaction, error)
	GetBlog(ctx context.Context, id string) (*Blog, error)
	GetBulkOperation(ctx context.Context, id string) (*BulkOperation, error)
	GetCarrierService(ctx context.Context, id string) (*CarrierService, error)
	GetCatalogPricing(ctx context.Context, id string) (*CatalogPricing, error)
	GetCategory(ctx context.Context, id string) (*Category, error)
	GetCDPEvent(ctx context.Context, id string) (*CDPEvent, error)
	GetCDPProfile(ctx context.Context, id string) (*CDPCustomerProfile, error)
	GetCDPSegment(ctx context.Context, id string) (*CDPSegment, error)
	GetChannel(ctx context.Context, id string) (*Channel, error)
	GetChannelPrices(ctx context.Context, channelID string) (json.RawMessage, error)
	GetChannelProductListing(ctx context.Context, channelID, productID string) (*ChannelProductListing, error)
	GetCheckoutSettings(ctx context.Context) (*CheckoutSettings, error)
	GetCollection(ctx context.Context, id string) (*Collection, error)
	GetCompanyCatalog(ctx context.Context, id string) (*CompanyCatalog, error)
	GetCompanyCredit(ctx context.Context, id string) (*CompanyCredit, error)
	GetConversation(ctx context.Context, id string) (*Conversation, error)
	GetCountry(ctx context.Context, code string) (*Country, error)
	GetCoupon(ctx context.Context, id string) (*Coupon, error)
	GetCouponByCode(ctx context.Context, code string) (*Coupon, error)
	GetCurrency(ctx context.Context, code string) (*Currency, error)
	GetCurrentBulkOperation(ctx context.Context) (*BulkOperation, error)
	GetCustomer(ctx context.Context, id string) (*Customer, error)
	// Customer coupon promotions (documented)
	GetCustomerCouponPromotions(ctx context.Context, id string) (json.RawMessage, error)
	GetCustomerPromotions(ctx context.Context, id string) (*CustomerPromotionsResponse, error)
	GetLineCustomer(ctx context.Context, lineID string) (*Customer, error)
	GetCustomerAddress(ctx context.Context, customerID, addressID string) (*CustomerAddress, error)
	GetCustomerBlacklist(ctx context.Context, id string) (*CustomerBlacklist, error)
	GetCustomerGroup(ctx context.Context, id string) (*CustomerGroup, error)
	// Customer group children (documented)
	GetCustomerGroupChildren(ctx context.Context, parentGroupID string) (json.RawMessage, error)
	GetCustomerGroupChildCustomerIDs(ctx context.Context, parentGroupID, childGroupID string) (*CustomerGroupIDsResponse, error)
	GetCustomerSavedSearch(ctx context.Context, id string) (*CustomerSavedSearch, error)
	GetCustomField(ctx context.Context, id string) (*CustomField, error)
	GetDeliveryOption(ctx context.Context, id string) (*DeliveryOption, error)
	// Delivery options (documented endpoints)
	GetDeliveryConfig(ctx context.Context, opts *DeliveryConfigOptions) (json.RawMessage, error)
	GetDeliveryTimeSlotsOpenAPI(ctx context.Context, id string) (json.RawMessage, error)
	UpdateDeliveryOptionStoresInfo(ctx context.Context, id string, body any) (json.RawMessage, error)
	GetDiscountCode(ctx context.Context, id string) (*DiscountCode, error)
	GetDiscountCodeByCode(ctx context.Context, code string) (*DiscountCode, error)
	GetDispute(ctx context.Context, id string) (*Dispute, error)
	GetDomain(ctx context.Context, id string) (*Domain, error)
	GetDraftOrder(ctx context.Context, id string) (*DraftOrder, error)
	GetFile(ctx context.Context, id string) (*File, error)
	GetFlashPrice(ctx context.Context, id string) (*FlashPrice, error)
	GetFlashPriceCampaign(ctx context.Context, id string) (json.RawMessage, error)
	GetFulfillment(ctx context.Context, id string) (*Fulfillment, error)
	GetFulfillmentOrder(ctx context.Context, id string) (*FulfillmentOrder, error)
	GetFulfillmentService(ctx context.Context, id string) (*FulfillmentService, error)
	GetGift(ctx context.Context, id string) (*Gift, error)
	GetGiftStocks(ctx context.Context, id string) (json.RawMessage, error)
	GetGiftCard(ctx context.Context, id string) (*GiftCard, error)
	GetInventoryLevel(ctx context.Context, id string) (*InventoryLevel, error)
	GetLabel(ctx context.Context, id string) (*Label, error)
	GetLockedInventoryCount(ctx context.Context, productID string) (*LockedInventoryCount, error)
	GetLocalDeliveryOption(ctx context.Context, id string) (*LocalDeliveryOption, error)
	GetLocation(ctx context.Context, id string) (*Location, error)
	GetMarket(ctx context.Context, id string) (*Market, error)
	GetMarketingEvent(ctx context.Context, id string) (*MarketingEvent, error)
	GetMedia(ctx context.Context, id string) (*Media, error)
	GetMemberPoints(ctx context.Context, customerID string) (*MemberPoints, error)
	GetCustomersMembershipInfo(ctx context.Context) (json.RawMessage, error)
	GetCustomerMemberPointsHistory(ctx context.Context, customerID string) (json.RawMessage, error)
	UpdateCustomerMemberPoints(ctx context.Context, customerID string, body any) (json.RawMessage, error)
	GetCustomerMembershipTierActionLogs(ctx context.Context, customerID string) (json.RawMessage, error)
	GetMembershipTier(ctx context.Context, id string) (*MembershipTier, error)
	GetMerchant(ctx context.Context) (*Merchant, error)
	// Merchants (documented endpoints)
	GetMerchantByID(ctx context.Context, merchantID string) (json.RawMessage, error)
	GenerateMerchantExpressLink(ctx context.Context, body any) (json.RawMessage, error)
	// Merchant metafields (/merchants/current/*)
	ListMerchantMetafields(ctx context.Context) (json.RawMessage, error)
	GetMerchantMetafield(ctx context.Context, metafieldID string) (json.RawMessage, error)
	CreateMerchantMetafield(ctx context.Context, body any) (json.RawMessage, error)
	UpdateMerchantMetafield(ctx context.Context, metafieldID string, body any) (json.RawMessage, error)
	DeleteMerchantMetafield(ctx context.Context, metafieldID string) error
	BulkCreateMerchantMetafields(ctx context.Context, body any) error
	BulkUpdateMerchantMetafields(ctx context.Context, body any) error
	BulkDeleteMerchantMetafields(ctx context.Context, body any) error
	// Merchant app metafields (/merchants/current/app_metafields*)
	ListMerchantAppMetafields(ctx context.Context) (json.RawMessage, error)
	GetMerchantAppMetafield(ctx context.Context, metafieldID string) (json.RawMessage, error)
	CreateMerchantAppMetafield(ctx context.Context, body any) (json.RawMessage, error)
	UpdateMerchantAppMetafield(ctx context.Context, metafieldID string, body any) (json.RawMessage, error)
	DeleteMerchantAppMetafield(ctx context.Context, metafieldID string) error
	BulkCreateMerchantAppMetafields(ctx context.Context, body any) error
	BulkUpdateMerchantAppMetafields(ctx context.Context, body any) error
	BulkDeleteMerchantAppMetafields(ctx context.Context, body any) error
	GetMerchantStaff(ctx context.Context, id string) (*MerchantStaff, error)
	GetMetafield(ctx context.Context, id string) (*Metafield, error)
	GetMetafieldDefinition(ctx context.Context, id string) (*MetafieldDefinition, error)
	GetMultipass(ctx context.Context) (*Multipass, error)
	// Multipass (documented endpoints)
	GetMultipassSecret(ctx context.Context) (json.RawMessage, error)
	CreateMultipassSecret(ctx context.Context, body any) (json.RawMessage, error)
	ListMultipassLinkings(ctx context.Context, customerIDs []string) (json.RawMessage, error)
	UpdateMultipassCustomerLinking(ctx context.Context, customerID string, body any) (json.RawMessage, error)
	DeleteMultipassCustomerLinking(ctx context.Context, customerID string) (json.RawMessage, error)
	GetOperationLog(ctx context.Context, id string) (*OperationLog, error)
	ExecuteShipment(ctx context.Context, id string, body any) (json.RawMessage, error)
	GetOrder(ctx context.Context, id string) (*Order, error)
	GetOrderActionLogs(ctx context.Context, id string) (json.RawMessage, error)
	GetOrderDelivery(ctx context.Context, orderID string) (*OrderDelivery, error)
	GetOrderAttribution(ctx context.Context, orderID string) (*OrderAttribution, error)
	GetOrderLabels(ctx context.Context, opts any) (json.RawMessage, error)
	GetOrderTags(ctx context.Context, id string) (*OrderTagsResponse, error)
	GetOrderTransactions(ctx context.Context, opts any) (json.RawMessage, error)
	// Order metafields (nested endpoints under /orders/{id}/...)
	ListOrderMetafields(ctx context.Context, orderID string) (json.RawMessage, error)
	GetOrderMetafield(ctx context.Context, orderID, metafieldID string) (json.RawMessage, error)
	CreateOrderMetafield(ctx context.Context, orderID string, body any) (json.RawMessage, error)
	UpdateOrderMetafield(ctx context.Context, orderID, metafieldID string, body any) (json.RawMessage, error)
	DeleteOrderMetafield(ctx context.Context, orderID, metafieldID string) error
	BulkCreateOrderMetafields(ctx context.Context, orderID string, body any) error
	BulkUpdateOrderMetafields(ctx context.Context, orderID string, body any) error
	BulkDeleteOrderMetafields(ctx context.Context, orderID string, body any) error

	// Order app metafields (nested endpoints under /orders/{id}/...)
	ListOrderAppMetafields(ctx context.Context, orderID string) (json.RawMessage, error)
	GetOrderAppMetafield(ctx context.Context, orderID, metafieldID string) (json.RawMessage, error)
	CreateOrderAppMetafield(ctx context.Context, orderID string, body any) (json.RawMessage, error)
	UpdateOrderAppMetafield(ctx context.Context, orderID, metafieldID string, body any) (json.RawMessage, error)
	DeleteOrderAppMetafield(ctx context.Context, orderID, metafieldID string) error
	BulkCreateOrderAppMetafields(ctx context.Context, orderID string, body any) error
	BulkUpdateOrderAppMetafields(ctx context.Context, orderID string, body any) error
	BulkDeleteOrderAppMetafields(ctx context.Context, orderID string, body any) error

	// Order item metafields (nested endpoints under /orders/{id}/items/...)
	ListOrderItemMetafields(ctx context.Context, orderID string) (json.RawMessage, error)
	BulkCreateOrderItemMetafields(ctx context.Context, orderID string, body any) error
	BulkUpdateOrderItemMetafields(ctx context.Context, orderID string, body any) error
	BulkDeleteOrderItemMetafields(ctx context.Context, orderID string, body any) error

	// Order item app metafields (nested endpoints under /orders/{id}/items/...)
	ListOrderItemAppMetafields(ctx context.Context, orderID string) (json.RawMessage, error)
	BulkCreateOrderItemAppMetafields(ctx context.Context, orderID string, body any) error
	BulkUpdateOrderItemAppMetafields(ctx context.Context, orderID string, body any) error
	BulkDeleteOrderItemAppMetafields(ctx context.Context, orderID string, body any) error

	// Customer metafields (nested endpoints under /customers/{id}/...)
	ListCustomerMetafields(ctx context.Context, customerID string) (json.RawMessage, error)
	GetCustomerMetafield(ctx context.Context, customerID, metafieldID string) (json.RawMessage, error)
	CreateCustomerMetafield(ctx context.Context, customerID string, body any) (json.RawMessage, error)
	UpdateCustomerMetafield(ctx context.Context, customerID, metafieldID string, body any) (json.RawMessage, error)
	DeleteCustomerMetafield(ctx context.Context, customerID, metafieldID string) error
	BulkCreateCustomerMetafields(ctx context.Context, customerID string, body any) error
	BulkUpdateCustomerMetafields(ctx context.Context, customerID string, body any) error
	BulkDeleteCustomerMetafields(ctx context.Context, customerID string, body any) error

	// Customer app metafields (nested endpoints under /customers/{id}/...)
	ListCustomerAppMetafields(ctx context.Context, customerID string) (json.RawMessage, error)
	GetCustomerAppMetafield(ctx context.Context, customerID, metafieldID string) (json.RawMessage, error)
	CreateCustomerAppMetafield(ctx context.Context, customerID string, body any) (json.RawMessage, error)
	UpdateCustomerAppMetafield(ctx context.Context, customerID, metafieldID string, body any) (json.RawMessage, error)
	DeleteCustomerAppMetafield(ctx context.Context, customerID, metafieldID string) error
	BulkCreateCustomerAppMetafields(ctx context.Context, customerID string, body any) error
	BulkUpdateCustomerAppMetafields(ctx context.Context, customerID string, body any) error
	BulkDeleteCustomerAppMetafields(ctx context.Context, customerID string, body any) error

	// Product metafields (nested endpoints under /products/{id}/...)
	ListProductMetafields(ctx context.Context, productID string) (json.RawMessage, error)
	GetProductMetafield(ctx context.Context, productID, metafieldID string) (json.RawMessage, error)
	CreateProductMetafield(ctx context.Context, productID string, body any) (json.RawMessage, error)
	UpdateProductMetafield(ctx context.Context, productID, metafieldID string, body any) (json.RawMessage, error)
	DeleteProductMetafield(ctx context.Context, productID, metafieldID string) error
	BulkCreateProductMetafields(ctx context.Context, productID string, body any) error
	BulkUpdateProductMetafields(ctx context.Context, productID string, body any) error
	BulkDeleteProductMetafields(ctx context.Context, productID string, body any) error

	// Product app metafields (nested endpoints under /products/{id}/...)
	ListProductAppMetafields(ctx context.Context, productID string) (json.RawMessage, error)
	GetProductAppMetafield(ctx context.Context, productID, metafieldID string) (json.RawMessage, error)
	CreateProductAppMetafield(ctx context.Context, productID string, body any) (json.RawMessage, error)
	UpdateProductAppMetafield(ctx context.Context, productID, metafieldID string, body any) (json.RawMessage, error)
	DeleteProductAppMetafield(ctx context.Context, productID, metafieldID string) error
	BulkCreateProductAppMetafields(ctx context.Context, productID string, body any) error
	BulkUpdateProductAppMetafields(ctx context.Context, productID string, body any) error
	BulkDeleteProductAppMetafields(ctx context.Context, productID string, body any) error

	GetOrderRisk(ctx context.Context, orderID, riskID string) (*OrderRisk, error)
	GetPage(ctx context.Context, id string) (*Page, error)
	GetPayment(ctx context.Context, id string) (*Payment, error)
	GetPayout(ctx context.Context, id string) (*Payout, error)
	GetPickupLocation(ctx context.Context, id string) (*PickupLocation, error)
	GetPriceRule(ctx context.Context, id string) (*PriceRule, error)
	GetProduct(ctx context.Context, id string) (*Product, error)
	GetProductListing(ctx context.Context, id string) (*ProductListing, error)
	GetProductReview(ctx context.Context, id string) (*ProductReview, error)
	GetProductReviewComment(ctx context.Context, id string) (json.RawMessage, error)
	GetProductPromotions(ctx context.Context, productID string) (json.RawMessage, error)
	GetProductStocks(ctx context.Context, productID string) (json.RawMessage, error)
	GetProductSubscription(ctx context.Context, id string) (*ProductSubscription, error)
	GetPromotion(ctx context.Context, id string) (*Promotion, error)
	GetPurchaseOrder(ctx context.Context, id string) (*PurchaseOrder, error)
	GetPOSPurchaseOrder(ctx context.Context, id string) (json.RawMessage, error)
	GetRedirect(ctx context.Context, id string) (*Redirect, error)
	GetRefund(ctx context.Context, id string) (*Refund, error)
	GetReturnOrder(ctx context.Context, id string) (*ReturnOrder, error)
	GetSale(ctx context.Context, id string) (*Sale, error)
	GetSaleProducts(ctx context.Context, saleID string, opts *SaleProductsListOptions) (*SaleProductsListResponse, error)
	AddSaleProducts(ctx context.Context, saleID string, req *SaleAddProductsRequest) (*SaleProductsListResponse, error)
	UpdateSaleProducts(ctx context.Context, saleID string, req *SaleUpdateProductsRequest) (*SaleProductsListResponse, error)
	GetSaleComments(ctx context.Context, saleID string, opts *SaleCommentsListOptions) (*SaleCommentsListResponse, error)
	GetSaleCustomers(ctx context.Context, saleID string, opts *SaleCustomersListOptions) (*SaleCustomersListResponse, error)
	GetScriptTag(ctx context.Context, id string) (*ScriptTag, error)
	GetSellingPlan(ctx context.Context, id string) (*SellingPlan, error)
	GetSettings(ctx context.Context) (*SettingsResponse, error)
	// Documented /settings/* endpoints (raw JSON)
	GetSettingsCheckout(ctx context.Context) (json.RawMessage, error)
	GetSettingsDomains(ctx context.Context) (json.RawMessage, error)
	GetSettingsLayouts(ctx context.Context) (json.RawMessage, error)
	GetSettingsLayoutsDraft(ctx context.Context) (json.RawMessage, error)
	GetSettingsOrders(ctx context.Context) (json.RawMessage, error)
	GetSettingsPayments(ctx context.Context) (json.RawMessage, error)
	GetSettingsPOS(ctx context.Context) (json.RawMessage, error)
	GetSettingsProductReview(ctx context.Context) (json.RawMessage, error)
	GetSettingsProducts(ctx context.Context) (json.RawMessage, error)
	GetSettingsPromotions(ctx context.Context) (json.RawMessage, error)
	GetSettingsShop(ctx context.Context) (json.RawMessage, error)
	GetSettingsTax(ctx context.Context) (json.RawMessage, error)
	GetSettingsTheme(ctx context.Context) (json.RawMessage, error)
	GetSettingsThemeDraft(ctx context.Context) (json.RawMessage, error)
	GetSettingsThirdPartyAds(ctx context.Context) (json.RawMessage, error)
	GetSettingsUsers(ctx context.Context) (json.RawMessage, error)
	GetShipment(ctx context.Context, id string) (*Shipment, error)
	GetShippingZone(ctx context.Context, id string) (*ShippingZone, error)
	GetShop(ctx context.Context) (*Shop, error)
	GetShopSettings(ctx context.Context) (*ShopSettings, error)
	GetSizeChart(ctx context.Context, id string) (*SizeChart, error)
	GetSmartCollection(ctx context.Context, id string) (*SmartCollection, error)
	GetStaff(ctx context.Context, id string) (*Staff, error)
	GetStaffPermissions(ctx context.Context, staffID string) (json.RawMessage, error)
	GetStorefrontCart(ctx context.Context, id string) (*StorefrontCart, error)
	GetStorefrontOAuthClient(ctx context.Context, id string) (*StorefrontOAuthClient, error)
	GetStorefrontOAuthApplication(ctx context.Context, id string) (*StorefrontOAuthApplication, error)
	GetStorefrontProduct(ctx context.Context, id string) (*StorefrontProduct, error)
	GetStorefrontProductByHandle(ctx context.Context, handle string) (*StorefrontProduct, error)
	GetStorefrontPromotion(ctx context.Context, id string) (*StorefrontPromotion, error)
	GetStorefrontPromotionByCode(ctx context.Context, code string) (*StorefrontPromotion, error)
	GetStorefrontToken(ctx context.Context, id string) (*StorefrontToken, error)
	GetSubscription(ctx context.Context, id string) (*Subscription, error)
	GetTag(ctx context.Context, id string) (*Tag, error)
	GetTax(ctx context.Context, id string) (*Tax, error)
	GetTaxonomy(ctx context.Context, id string) (*Taxonomy, error)
	GetTaxService(ctx context.Context, id string) (*TaxService, error)
	GetTheme(ctx context.Context, id string) (*Theme, error)
	GetToken(ctx context.Context, id string) (*Token, error)
	GetTokenInfo(ctx context.Context) (json.RawMessage, error)
	GetTransaction(ctx context.Context, id string) (*Transaction, error)
	GetUserCoupon(ctx context.Context, id string) (*UserCoupon, error)
	GetWarehouse(ctx context.Context, id string) (*Warehouse, error)
	GetWebhook(ctx context.Context, id string) (*Webhook, error)
	GetWishList(ctx context.Context, id string) (*WishList, error)
	ListAbandonedCheckouts(ctx context.Context, opts *AbandonedCheckoutsListOptions) (*AbandonedCheckoutsListResponse, error)
	ListAddonProducts(ctx context.Context, opts *AddonProductsListOptions) (*AddonProductsListResponse, error)
	ListAffiliateCampaigns(ctx context.Context, opts *AffiliateCampaignsListOptions) (*AffiliateCampaignsListResponse, error)
	ListArchivedOrders(ctx context.Context, opts *ArchivedOrdersListOptions) (*OrdersListResponse, error)
	ListArticles(ctx context.Context, opts *ArticlesListOptions) (*ArticlesListResponse, error)
	ListAssets(ctx context.Context, themeID string) (*AssetsListResponse, error)
	ListBalanceTransactions(ctx context.Context, opts *BalanceTransactionsListOptions) (*BalanceTransactionsListResponse, error)
	ListBlogs(ctx context.Context, opts *BlogsListOptions) (*BlogsListResponse, error)
	ListBulkOperations(ctx context.Context, opts *BulkOperationsListOptions) (*BulkOperationsListResponse, error)
	ListCarrierServices(ctx context.Context, opts *CarrierServicesListOptions) (*CarrierServicesListResponse, error)
	ListCatalogPricing(ctx context.Context, opts *CatalogPricingListOptions) (*CatalogPricingListResponse, error)
	ListCategories(ctx context.Context, opts *CategoriesListOptions) (*CategoriesListResponse, error)
	ListCDPEvents(ctx context.Context, opts *CDPEventsListOptions) (*CDPEventsListResponse, error)
	ListCDPProfiles(ctx context.Context, opts *CDPProfilesListOptions) (*CDPProfilesListResponse, error)
	ListCDPSegments(ctx context.Context, opts *CDPSegmentsListOptions) (*CDPSegmentsListResponse, error)
	ListChannelProductListings(ctx context.Context, channelID string, opts *ChannelProductsListOptions) (*ChannelProductsListResponse, error)
	ListChannelProducts(ctx context.Context, channelID string, page, pageSize int) (*ChannelProductsResponse, error)
	ListChannels(ctx context.Context, opts *ChannelsListOptions) (*ChannelsListResponse, error)
	ListCollections(ctx context.Context, opts *CollectionsListOptions) (*CollectionsListResponse, error)
	ListCompanyCatalogs(ctx context.Context, opts *CompanyCatalogsListOptions) (*CompanyCatalogsListResponse, error)
	ListCompanyCredits(ctx context.Context, opts *CompanyCreditsListOptions) (*CompanyCreditsListResponse, error)
	ListCompanyCreditTransactions(ctx context.Context, creditID string, page, pageSize int) (*CompanyCreditTransactionsListResponse, error)
	ListConversationMessages(ctx context.Context, conversationID string, page, pageSize int) (*ConversationMessagesListResponse, error)
	ListConversations(ctx context.Context, opts *ConversationsListOptions) (*ConversationsListResponse, error)
	ListCountries(ctx context.Context) (*CountriesListResponse, error)
	ListCoupons(ctx context.Context, opts *CouponsListOptions) (*CouponsListResponse, error)
	ListCurrencies(ctx context.Context) (*CurrenciesListResponse, error)
	ListCustomerAddresses(ctx context.Context, customerID string, opts *CustomerAddressesListOptions) (*CustomerAddressesListResponse, error)
	ListCustomerBlacklist(ctx context.Context, opts *CustomerBlacklistListOptions) (*CustomerBlacklistListResponse, error)
	ListCustomerGroups(ctx context.Context, opts *CustomerGroupsListOptions) (*CustomerGroupsListResponse, error)
	ListCustomers(ctx context.Context, opts *CustomersListOptions) (*CustomersListResponse, error)
	SearchAddonProducts(ctx context.Context, opts *AddonProductSearchOptions) (*AddonProductsListResponse, error)
	SearchCustomerGroups(ctx context.Context, opts *CustomerGroupSearchOptions) (*CustomerGroupsListResponse, error)
	SearchCustomers(ctx context.Context, opts *CustomerSearchOptions) (*CustomersListResponse, error)
	SearchGifts(ctx context.Context, opts *GiftSearchOptions) (*GiftsListResponse, error)
	SearchOrders(ctx context.Context, opts *OrderSearchOptions) (*OrdersListResponse, error)
	SearchProducts(ctx context.Context, opts *ProductSearchOptions) (*ProductsListResponse, error)
	SearchProductsPost(ctx context.Context, req *ProductSearchRequest) (*ProductsListResponse, error)
	SearchPromotions(ctx context.Context, opts *PromotionSearchOptions) (*PromotionsListResponse, error)
	ListCustomerSavedSearches(ctx context.Context, opts *CustomerSavedSearchesListOptions) (*CustomerSavedSearchesListResponse, error)
	ListCustomFields(ctx context.Context, opts *CustomFieldsListOptions) (*CustomFieldsListResponse, error)
	ListDeliveryOptions(ctx context.Context, opts *DeliveryOptionsListOptions) (*DeliveryOptionsListResponse, error)
	ListDeliveryTimeSlots(ctx context.Context, id string, opts *DeliveryTimeSlotsListOptions) (*DeliveryTimeSlotsListResponse, error)
	ListDiscountCodes(ctx context.Context, opts *DiscountCodesListOptions) (*DiscountCodesListResponse, error)
	ListDisputes(ctx context.Context, opts *DisputesListOptions) (*DisputesListResponse, error)
	ListDomains(ctx context.Context, opts *DomainsListOptions) (*DomainsListResponse, error)
	ListDraftOrders(ctx context.Context, opts *DraftOrdersListOptions) (*DraftOrdersListResponse, error)
	ListFiles(ctx context.Context, opts *FilesListOptions) (*FilesListResponse, error)
	ListFlashPrices(ctx context.Context, opts *FlashPriceListOptions) (*FlashPriceListResponse, error)
	// Flash price campaigns (documented endpoints)
	ListFlashPriceCampaigns(ctx context.Context, opts *FlashPriceCampaignsListOptions) (json.RawMessage, error)
	ListFulfillmentOrders(ctx context.Context, opts *FulfillmentOrdersListOptions) (*FulfillmentOrdersListResponse, error)
	ListFulfillments(ctx context.Context, opts *FulfillmentsListOptions) (*FulfillmentsListResponse, error)
	ListFulfillmentServices(ctx context.Context, opts *FulfillmentServicesListOptions) (*FulfillmentServicesListResponse, error)
	ListGiftCards(ctx context.Context, opts *GiftCardsListOptions) (*GiftCardsListResponse, error)
	ListGifts(ctx context.Context, opts *GiftsListOptions) (*GiftsListResponse, error)
	ListInventoryLevels(ctx context.Context, opts *InventoryListOptions) (*InventoryListResponse, error)
	ListLabels(ctx context.Context, opts *LabelsListOptions) (*LabelsListResponse, error)
	ListLocalDeliveryOptions(ctx context.Context, opts *LocalDeliveryListOptions) (*LocalDeliveryListResponse, error)
	ListLocations(ctx context.Context, opts *LocationsListOptions) (*LocationsListResponse, error)
	ListMarketingEvents(ctx context.Context, opts *MarketingEventsListOptions) (*MarketingEventsListResponse, error)
	ListMarkets(ctx context.Context, opts *MarketsListOptions) (*MarketsListResponse, error)
	ListMedias(ctx context.Context, opts *MediasListOptions) (*MediasListResponse, error)
	ListMembershipTiers(ctx context.Context, opts *MembershipTiersListOptions) (*MembershipTiersListResponse, error)
	ListMemberPointRules(ctx context.Context) (json.RawMessage, error)
	ListMerchants(ctx context.Context) ([]Merchant, error)
	ListMerchantStaff(ctx context.Context, opts *MerchantStaffListOptions) (*MerchantStaffListResponse, error)
	ListMetafieldDefinitions(ctx context.Context, opts *MetafieldDefinitionsListOptions) (*MetafieldDefinitionsListResponse, error)
	ListMetafields(ctx context.Context, opts *MetafieldsListOptions) (*MetafieldsListResponse, error)
	ListOperationLogs(ctx context.Context, opts *OperationLogsListOptions) (*OperationLogsListResponse, error)
	ListOrderAttributions(ctx context.Context, opts *OrderAttributionListOptions) (*OrderAttributionListResponse, error)
	ListOrderFulfillmentOrders(ctx context.Context, orderID string) (*FulfillmentOrdersListResponse, error)
	ListOrderPayments(ctx context.Context, orderID string) (*PaymentsListResponse, error)
	ListOrderRefunds(ctx context.Context, orderID string) (*RefundsListResponse, error)
	ListOrderRisks(ctx context.Context, orderID string, opts *OrderRisksListOptions) (*OrderRisksListResponse, error)
	ListOrderTags(ctx context.Context) (json.RawMessage, error)
	ListOrders(ctx context.Context, opts *OrdersListOptions) (*OrdersListResponse, error)
	ListOrderTransactions(ctx context.Context, orderID string) (*TransactionsListResponse, error)
	ListPages(ctx context.Context, opts *PagesListOptions) (*PagesListResponse, error)
	ListPayments(ctx context.Context, opts *PaymentsListOptions) (*PaymentsListResponse, error)
	ListPayouts(ctx context.Context, opts *PayoutsListOptions) (*PayoutsListResponse, error)
	ListPickupLocations(ctx context.Context, opts *PickupListOptions) (*PickupListResponse, error)
	ListPointsTransactions(ctx context.Context, customerID string, opts *PointsTransactionsListOptions) (*PointsTransactionsListResponse, error)
	ListPriceRules(ctx context.Context, opts *PriceRulesListOptions) (*PriceRulesListResponse, error)
	ListProductListings(ctx context.Context, opts *ProductListingsListOptions) (*ProductListingsListResponse, error)
	ListProductReviews(ctx context.Context, opts *ProductReviewsListOptions) (*ProductReviewsListResponse, error)
	ListProductReviewComments(ctx context.Context, opts *ProductReviewCommentsListOptions) (json.RawMessage, error)
	ListProducts(ctx context.Context, opts *ProductsListOptions) (*ProductsListResponse, error)
	ListProductSubscriptions(ctx context.Context, opts *ProductSubscriptionsListOptions) (*ProductSubscriptionsListResponse, error)
	ListPromotions(ctx context.Context, opts *PromotionsListOptions) (*PromotionsListResponse, error)
	// Promotions (documented endpoint)
	GetPromotionsCouponCenter(ctx context.Context) (json.RawMessage, error)
	ListPurchaseOrders(ctx context.Context, opts *PurchaseOrdersListOptions) (*PurchaseOrdersListResponse, error)
	ListPOSPurchaseOrders(ctx context.Context, opts *POSPurchaseOrdersListOptions) (json.RawMessage, error)
	ListRedirects(ctx context.Context, opts *RedirectsListOptions) (*RedirectsListResponse, error)
	ListRefunds(ctx context.Context, opts *RefundsListOptions) (*RefundsListResponse, error)
	ListReturnOrders(ctx context.Context, opts *ReturnOrdersListOptions) (*ReturnOrdersListResponse, error)
	ListSales(ctx context.Context, opts *SalesListOptions) (*SalesListResponse, error)
	ListScriptTags(ctx context.Context, opts *ScriptTagsListOptions) (*ScriptTagsListResponse, error)
	ListSellingPlans(ctx context.Context, opts *SellingPlansListOptions) (*SellingPlansListResponse, error)
	ListShipments(ctx context.Context, opts *ShipmentsListOptions) (*ShipmentsListResponse, error)
	ListShippingZones(ctx context.Context, opts *ShippingZonesListOptions) (*ShippingZonesListResponse, error)
	ListSizeCharts(ctx context.Context, opts *SizeChartsListOptions) (*SizeChartsListResponse, error)
	ListSmartCollections(ctx context.Context, opts *SmartCollectionsListOptions) (*SmartCollectionsListResponse, error)
	ListStaffs(ctx context.Context, opts *StaffsListOptions) (*StaffsListResponse, error)
	ListCustomerStoreCredits(ctx context.Context, customerID string, page, perPage int) (json.RawMessage, error)
	ListUserCredits(ctx context.Context, opts *UserCreditsListOptions) (json.RawMessage, error)
	ListStorefrontCarts(ctx context.Context, opts *StorefrontCartsListOptions) (*StorefrontCartsListResponse, error)
	ListStorefrontOAuthClients(ctx context.Context, opts *StorefrontOAuthClientsListOptions) (*StorefrontOAuthClientsListResponse, error)
	ListStorefrontOAuthApplications(ctx context.Context, opts *StorefrontOAuthApplicationsListOptions) (*StorefrontOAuthApplicationsListResponse, error)
	ListStorefrontProducts(ctx context.Context, opts *StorefrontProductsListOptions) (*StorefrontProductsListResponse, error)
	ListStorefrontPromotions(ctx context.Context, opts *StorefrontPromotionsListOptions) (*StorefrontPromotionsListResponse, error)
	ListStorefrontTokens(ctx context.Context, opts *StorefrontTokensListOptions) (*StorefrontTokensListResponse, error)
	ListSubscriptions(ctx context.Context, opts *SubscriptionsListOptions) (*SubscriptionsListResponse, error)
	ListTags(ctx context.Context, opts *TagsListOptions) (*TagsListResponse, error)
	ListTaxes(ctx context.Context, opts *TaxesListOptions) (*TaxesListResponse, error)
	ListTaxonomies(ctx context.Context, opts *TaxonomiesListOptions) (*TaxonomiesListResponse, error)
	ListTaxServices(ctx context.Context, opts *TaxServicesListOptions) (*TaxServicesListResponse, error)
	ListThemes(ctx context.Context, opts *ThemesListOptions) (*ThemesListResponse, error)
	ListTokens(ctx context.Context, opts *TokensListOptions) (*TokensListResponse, error)
	ListTransactions(ctx context.Context, opts *TransactionsListOptions) (*TransactionsListResponse, error)
	ListUserCoupons(ctx context.Context, opts *UserCouponsListOptions) (*UserCouponsListResponse, error)
	ListUserCouponsListEndpoint(ctx context.Context, opts *UserCouponsListEndpointOptions) (json.RawMessage, error)
	ListWarehouses(ctx context.Context, opts *WarehousesListOptions) (*WarehousesListResponse, error)
	ListWebhooks(ctx context.Context, opts *WebhooksListOptions) (*WebhooksListResponse, error)
	ListWishLists(ctx context.Context, opts *WishListsListOptions) (*WishListsListResponse, error)
	ListWishListItems(ctx context.Context, opts *WishListItemsListOptions) (json.RawMessage, error)
	MoveFulfillmentOrder(ctx context.Context, id string, newLocationID string) (*FulfillmentOrder, error)
	PostOrderMessage(ctx context.Context, id string, body any) (json.RawMessage, error)
	Post(ctx context.Context, path string, body, result interface{}) error
	PublishProductToChannel(ctx context.Context, channelID string, req *ChannelPublishProductRequest) error
	PublishProductToChannelListing(ctx context.Context, channelID string, req *ChannelProductPublishRequest) (*ChannelProductListing, error)
	Put(ctx context.Context, path string, body, result interface{}) error
	ReceivePurchaseOrder(ctx context.Context, id string) (*PurchaseOrder, error)
	ReceiveReturnOrder(ctx context.Context, id string) (*ReturnOrder, error)
	RefundPayment(ctx context.Context, id string, amount string, reason string) (*Payment, error)
	RemoveProductFromCollection(ctx context.Context, id, productID string) error
	RemoveWishListItem(ctx context.Context, wishListID, itemID string) error
	ReplaceProductTags(ctx context.Context, id string, tags []string) (*Product, error)
	RevokeUserCoupon(ctx context.Context, id string) error
	RedeemUserCoupon(ctx context.Context, couponCode string, body any) (json.RawMessage, error)
	RotateMultipassSecret(ctx context.Context) (*Multipass, error)
	RotateStorefrontOAuthClientSecret(ctx context.Context, id string) (*StorefrontOAuthClient, error)
	SendAbandonedCheckoutRecoveryEmail(ctx context.Context, id string) error
	SendConversationMessage(ctx context.Context, conversationID string, req *ConversationMessageCreateRequest) (*ConversationMessage, error)
	SendDraftOrderInvoice(ctx context.Context, id string) error
	SetDefaultCustomerAddress(ctx context.Context, customerID, addressID string) (*CustomerAddress, error)
	SetCustomerTags(ctx context.Context, id string, tags []string) (*Customer, error)
	SetInventoryLevel(ctx context.Context, req *InventoryLevelSetRequest) (*InventoryLevel, error)
	SplitOrder(ctx context.Context, id string, lineItemIDs []string) (*OrderSplitResponse, error)
	SubmitDispute(ctx context.Context, id string) (*Dispute, error)
	UnpublishProductFromChannel(ctx context.Context, channelID, productID string) error
	UnpublishProductFromChannelListing(ctx context.Context, channelID, productID string) error
	UpdateCustomer(ctx context.Context, id string, req *CustomerUpdateRequest) (*Customer, error)
	UpdateOrder(ctx context.Context, id string, req *OrderUpdateRequest) (*Order, error)
	UpdateOrderDelivery(ctx context.Context, orderID string, req *OrderDeliveryUpdateRequest) (*OrderDelivery, error)
	UpdateOrderDeliveryStatus(ctx context.Context, id string, status string) (*Order, error)
	UpdateOrderPaymentStatus(ctx context.Context, id string, status string) (*Order, error)
	UpdateOrderStatus(ctx context.Context, id string, status string) (*Order, error)
	UpdateOrderTags(ctx context.Context, id string, tags []string) (*Order, error)
	UpdateAddonProduct(ctx context.Context, id string, req *AddonProductUpdateRequest) (*AddonProduct, error)
	UpdateAddonProductQuantity(ctx context.Context, id string, req *AddonProductQuantityRequest) (*AddonProduct, error)
	UpdateAddonProductsQuantityBySKU(ctx context.Context, req *AddonProductQuantityBySKURequest) (*AddonProduct, error)
	UpdateAddonProductStocks(ctx context.Context, id string, req *AddonProductStocksUpdateRequest) (*AddonProductStocksResponse, error)
	UpdateAffiliateCampaign(ctx context.Context, id string, req *AffiliateCampaignUpdateRequest) (*AffiliateCampaign, error)
	UpdateArticle(ctx context.Context, id string, req *ArticleUpdateRequest) (*Article, error)
	UpdateAsset(ctx context.Context, themeID string, req *AssetUpdateRequest) (*Asset, error)
	UpdateBlog(ctx context.Context, id string, req *BlogUpdateRequest) (*Blog, error)
	UpdateCarrierService(ctx context.Context, id string, req *CarrierServiceUpdateRequest) (*CarrierService, error)
	UpdateCatalogPricing(ctx context.Context, id string, req *CatalogPricingUpdateRequest) (*CatalogPricing, error)
	UpdateCategory(ctx context.Context, id string, req *CategoryUpdateRequest) (*Category, error)
	BulkUpdateCategoryProductSorting(ctx context.Context, id string, body any) (json.RawMessage, error)
	UpdateChannel(ctx context.Context, id string, req *ChannelUpdateRequest) (*Channel, error)
	UpdateChannelProductPrice(ctx context.Context, channelID, productID, priceID string, body any) (json.RawMessage, error)
	UpdateChannelProductListing(ctx context.Context, channelID, productID string, req *ChannelProductUpdateRequest) (*ChannelProductListing, error)
	UpdateCheckoutSettings(ctx context.Context, req *CheckoutSettingsUpdateRequest) (*CheckoutSettings, error)
	UpdateCollection(ctx context.Context, id string, req *CollectionUpdateRequest) (*Collection, error)
	UpdateCompanyCatalog(ctx context.Context, id string, req *CompanyCatalogUpdateRequest) (*CompanyCatalog, error)
	UpdateConversation(ctx context.Context, id string, req *ConversationUpdateRequest) (*Conversation, error)
	UpdateCoupon(ctx context.Context, id string, req *CouponUpdateRequest) (*Coupon, error)
	UpdateCurrency(ctx context.Context, code string, req *CurrencyUpdateRequest) (*Currency, error)
	UpdateCustomerGroup(ctx context.Context, id string, req *CustomerGroupUpdateRequest) (*CustomerGroup, error)
	UpdateCustomerTags(ctx context.Context, id string, req *CustomerTagsUpdateRequest) (*Customer, error)
	UpdateCustomerSubscriptions(ctx context.Context, customerID string, body any) (json.RawMessage, error)
	UpdateCustomerStoreCredits(ctx context.Context, customerID string, req *StoreCreditUpdateRequest) (json.RawMessage, error)
	BulkUpdateUserCredits(ctx context.Context, body any) (json.RawMessage, error)
	BulkUpdateMemberPoints(ctx context.Context, body any) (json.RawMessage, error)
	UpdateProductReviewComment(ctx context.Context, id string, body any) (json.RawMessage, error)
	BulkUpdateProductReviewComments(ctx context.Context, body any) (json.RawMessage, error)
	BulkUpdateProductStocks(ctx context.Context, body any) error
	BulkDeleteProducts(ctx context.Context, productIDs []string) error
	BulkDeleteProductReviewComments(ctx context.Context, body any) (json.RawMessage, error)
	UpdateCustomField(ctx context.Context, id string, req *CustomFieldUpdateRequest) (*CustomField, error)
	UpdateDeliveryOptionPickupStore(ctx context.Context, id string, req *PickupStoreUpdateRequest) (*DeliveryOption, error)
	UpdateDisputeEvidence(ctx context.Context, id string, req *DisputeUpdateEvidenceRequest) (*Dispute, error)
	UpdateDomain(ctx context.Context, id string, req *DomainUpdateRequest) (*Domain, error)
	UpdateFile(ctx context.Context, id string, req *FileUpdateRequest) (*File, error)
	UpdateFlashPrice(ctx context.Context, id string, req *FlashPriceUpdateRequest) (*FlashPrice, error)
	UpdateFlashPriceCampaign(ctx context.Context, id string, body any) (json.RawMessage, error)
	UpdateFulfillmentService(ctx context.Context, id string, req *FulfillmentServiceUpdateRequest) (*FulfillmentService, error)
	UpdateGift(ctx context.Context, id string, req *GiftUpdateRequest) (*Gift, error)
	UpdateGiftQuantity(ctx context.Context, id string, quantity int) (*Gift, error)
	UpdateGiftStocks(ctx context.Context, id string, body any) (json.RawMessage, error)
	UpdateGiftsQuantityBySKU(ctx context.Context, sku string, quantity int) error
	UpdatePOSPurchaseOrder(ctx context.Context, id string, body any) (json.RawMessage, error)
	BulkDeletePOSPurchaseOrders(ctx context.Context, body any) (json.RawMessage, error)
	UpdateLabel(ctx context.Context, id string, req *LabelUpdateRequest) (*Label, error)
	UpdateLocalDeliveryOption(ctx context.Context, id string, req *LocalDeliveryUpdateRequest) (*LocalDeliveryOption, error)
	UpdateLocation(ctx context.Context, id string, req *LocationUpdateRequest) (*Location, error)
	UpdateMarketingEvent(ctx context.Context, id string, req *MarketingEventUpdateRequest) (*MarketingEvent, error)
	UpdateMedia(ctx context.Context, id string, req *MediaUpdateRequest) (*Media, error)
	UpdateMetafield(ctx context.Context, id string, req *MetafieldUpdateRequest) (*Metafield, error)
	UpdateMetafieldDefinition(ctx context.Context, id string, req *MetafieldDefinitionUpdateRequest) (*MetafieldDefinition, error)
	UpdateOrderRisk(ctx context.Context, orderID, riskID string, req *OrderRiskUpdateRequest) (*OrderRisk, error)
	UpdatePage(ctx context.Context, id string, req *PageUpdateRequest) (*Page, error)
	UpdatePickupLocation(ctx context.Context, id string, req *PickupUpdateRequest) (*PickupLocation, error)
	UpdatePriceRule(ctx context.Context, id string, req *PriceRuleUpdateRequest) (*PriceRule, error)
	UpdateProduct(ctx context.Context, id string, req *ProductUpdateRequest) (*Product, error)
	UpdateProductPrice(ctx context.Context, id string, price float64) (*Product, error)
	UpdateProductQuantity(ctx context.Context, id string, quantity int) (*Product, error)
	UpdateProductQuantityBySKU(ctx context.Context, sku string, quantity int) error
	UpdateProductStocks(ctx context.Context, productID string, body any) (json.RawMessage, error)
	UpdateProductTags(ctx context.Context, id string, req *ProductTagsUpdateRequest) (*Product, error)
	UpdateProductVariation(ctx context.Context, productID string, variationID string, req *ProductVariationUpdateRequest) (*ProductVariation, error)
	UpdateProductVariationPrice(ctx context.Context, productID string, variationID string, price float64) (*Product, error)
	UpdateProductVariationQuantity(ctx context.Context, productID string, variationID string, quantity int) (*Product, error)
	UpdateProductsLabelsBulk(ctx context.Context, body any) error
	UpdateProductsRetailStatusBulk(ctx context.Context, body any) error
	UpdateProductsStatusBulk(ctx context.Context, body any) error
	UpdatePromotion(ctx context.Context, id string, req *PromotionUpdateRequest) (*Promotion, error)
	UpdateRedirect(ctx context.Context, id string, req *RedirectUpdateRequest) (*Redirect, error)
	UpdateReturnOrder(ctx context.Context, id string, req *ReturnOrderUpdateRequest) (*ReturnOrder, error)
	UpdateScriptTag(ctx context.Context, id string, req *ScriptTagUpdateRequest) (*ScriptTag, error)
	UpdateSettings(ctx context.Context, req *UserSettingsUpdateRequest) (*SettingsResponse, error)
	UpdateSettingsDomains(ctx context.Context, body any) (json.RawMessage, error)
	UpdateSettingsLayoutsDraft(ctx context.Context, body any) (json.RawMessage, error)
	UpdateSettingsThemeDraft(ctx context.Context, body any) (json.RawMessage, error)
	PublishSettingsLayouts(ctx context.Context, body any) (json.RawMessage, error)
	PublishSettingsTheme(ctx context.Context, body any) (json.RawMessage, error)
	UpdateShipment(ctx context.Context, id string, req *ShipmentUpdateRequest) (*Shipment, error)
	UpdateShopSettings(ctx context.Context, req *ShopSettingsUpdateRequest) (*ShopSettings, error)
	UpdateSizeChart(ctx context.Context, id string, req *SizeChartUpdateRequest) (*SizeChart, error)
	UpdateSmartCollection(ctx context.Context, id string, req *SmartCollectionUpdateRequest) (*SmartCollection, error)
	UpdateStaff(ctx context.Context, id string, req *StaffUpdateRequest) (*Staff, error)
	UpdateStorefrontOAuthClient(ctx context.Context, id string, req *StorefrontOAuthClientUpdateRequest) (*StorefrontOAuthClient, error)
	UpdateTax(ctx context.Context, id string, req *TaxUpdateRequest) (*Tax, error)
	UpdateTaxonomy(ctx context.Context, id string, req *TaxonomyUpdateRequest) (*Taxonomy, error)
	UpdateTaxService(ctx context.Context, id string, req *TaxServiceUpdateRequest) (*TaxService, error)
	UpdateTheme(ctx context.Context, id string, req *ThemeUpdateRequest) (*Theme, error)
	UpdateWarehouse(ctx context.Context, id string, req *WarehouseUpdateRequest) (*Warehouse, error)
	UpdateWebhook(ctx context.Context, id string, req *WebhookUpdateRequest) (*Webhook, error)
	VerifyDomain(ctx context.Context, id string) (*Domain, error)
	VoidPayment(ctx context.Context, id string) (*Payment, error)
}

// AdminAPIClient provides access to undocumented Shopline admin endpoints
// via a proxy API service with bearer token auth.
type AdminAPIClient interface {
	// Orders - admin actions
	CommentOrder(ctx context.Context, orderID string, req *AdminCommentRequest) (json.RawMessage, error)
	ListOrderComments(ctx context.Context, orderID string) ([]AdminOrderComment, error)
	AdminRefundOrder(ctx context.Context, orderID string, req *AdminRefundRequest) (json.RawMessage, error)
	ReissueReceipt(ctx context.Context, orderID string) (json.RawMessage, error)

	// Products - visibility
	HideProduct(ctx context.Context, productID string) (json.RawMessage, error)
	PublishProduct(ctx context.Context, productID string) (json.RawMessage, error)
	UnpublishProduct(ctx context.Context, productID string) (json.RawMessage, error)

	// Shipping
	GetShipmentStatus(ctx context.Context, orderID string) (*AdminShipmentStatus, error)
	GetTrackingNumber(ctx context.Context, orderID string) (*AdminTrackingResponse, error)
	ExecuteShipment(ctx context.Context, orderID string, req *AdminExecuteShipmentRequest) (*AdminTrackingResponse, error)
	PrintPackingLabel(ctx context.Context, orderID string, req *AdminPrintLabelRequest) (*AdminPackingLabel, error)

	// Livestreams
	ListLivestreams(ctx context.Context, opts *AdminListStreamsOptions) (json.RawMessage, error)
	GetLivestream(ctx context.Context, streamID string) (json.RawMessage, error)
	CreateLivestream(ctx context.Context, req *AdminCreateStreamRequest) (json.RawMessage, error)
	UpdateLivestream(ctx context.Context, streamID string, req *AdminUpdateStreamRequest) (json.RawMessage, error)
	DeleteLivestream(ctx context.Context, streamID string) error
	AddStreamProducts(ctx context.Context, streamID string, req *AdminAddStreamProductsRequest) (json.RawMessage, error)
	RemoveStreamProducts(ctx context.Context, streamID string, req *AdminRemoveStreamProductsRequest) error
	StartLivestream(ctx context.Context, streamID string, req *AdminStartStreamRequest) (json.RawMessage, error)
	EndLivestream(ctx context.Context, streamID string) error
	GetStreamComments(ctx context.Context, streamID string, pageNo int) (json.RawMessage, error)
	GetStreamActiveVideos(ctx context.Context, streamID, platform string) (json.RawMessage, error)
	ToggleStreamProductDisplay(ctx context.Context, streamID, productID string, req *AdminToggleStreamProductRequest) (json.RawMessage, error)

	// Message Center
	ListConversations(ctx context.Context, opts *AdminListConversationsOptions) (json.RawMessage, error)
	SendMessage(ctx context.Context, conversationID string, req *AdminSendMessageRequest) (json.RawMessage, error)
	ListInstantMessages(ctx context.Context, opts *AdminListInstantMessagesOptions) (json.RawMessage, error)
	GetInstantMessages(ctx context.Context, conversationID string, qv *AdminInstantMessagesQuery) (json.RawMessage, error)
	SendInstantMessage(ctx context.Context, req *AdminSendInstantMessageRequest) (json.RawMessage, error)
	GetMessageCenterChannels(ctx context.Context) (json.RawMessage, error)
	GetMessageCenterStaffInfo(ctx context.Context) (json.RawMessage, error)
	GetMessageCenterProfile(ctx context.Context, scopeID string) (json.RawMessage, error)

	// Payments / Shoplytics / Express Links (admin public endpoints)
	CreateExpressLink(ctx context.Context, req *AdminCreateExpressLinkRequest) (json.RawMessage, error)
	GetPaymentsPayouts(ctx context.Context, opts *AdminPaymentsPayoutsOptions) (json.RawMessage, error)
	GetPaymentsAccountSummary(ctx context.Context) (json.RawMessage, error)
	GetShoplyticsCustomersNewAndReturning(ctx context.Context, opts *AdminShoplyticsNewReturningOptions) (json.RawMessage, error)
	GetShoplyticsCustomersFirstOrderChannels(ctx context.Context) (json.RawMessage, error)
	GetShoplyticsPaymentsMethodsGrid(ctx context.Context) (json.RawMessage, error)

	// Social Posts
	GetSocialChannels(ctx context.Context) (json.RawMessage, error)
	GetChannelPosts(ctx context.Context, opts *SocialChannelPostsOptions) (json.RawMessage, error)
	GetSocialCategories(ctx context.Context) (json.RawMessage, error)
	SearchSocialProducts(ctx context.Context, opts *SocialProductSearchOptions) (json.RawMessage, error)
	ListSalesEvents(ctx context.Context, opts *SalesEventListOptions) (json.RawMessage, error)
	GetSalesEvent(ctx context.Context, eventID string, fieldScopes string) (json.RawMessage, error)
	CreateSalesEvent(ctx context.Context, req *CreateSalesEventRequest) (json.RawMessage, error)
	ScheduleSalesEvent(ctx context.Context, eventID string, req *ScheduleSalesEventRequest) error
	DeleteSalesEvent(ctx context.Context, eventID string) error
	PublishSalesEvent(ctx context.Context, eventID string) error
	AddSalesEventProducts(ctx context.Context, eventID string, req *AddSalesEventProductsRequest) (json.RawMessage, error)
	UpdateSalesEventProductKeys(ctx context.Context, eventID string, req *UpdateProductKeysRequest) (json.RawMessage, error)
	LinkFacebookPost(ctx context.Context, eventID string, req *LinkFacebookPostRequest) (json.RawMessage, error)
	LinkInstagramPost(ctx context.Context, eventID string, req *LinkInstagramPostRequest) (json.RawMessage, error)
	LinkFBGroupPost(ctx context.Context, eventID string, req *LinkFBGroupPostRequest) (json.RawMessage, error)
}
