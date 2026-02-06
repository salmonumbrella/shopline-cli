package api

import (
	"context"
	"encoding/json"
	"fmt"
)

// MockClient is a mock implementation of APIClient for testing.
// Use the Set* methods to configure expected returns for each method.
type MockClient struct {
	// Handler functions for each API method
	handlers map[string]interface{}
}

// NewMockClient creates a new MockClient.
func NewMockClient() *MockClient {
	return &MockClient{
		handlers: make(map[string]interface{}),
	}
}

// SetHandler sets a handler function for a specific method.
// The handler should match the signature of the method being mocked.
func (m *MockClient) SetHandler(method string, handler interface{}) {
	m.handlers[method] = handler
}

// notImplemented returns an error for methods without handlers.
func (m *MockClient) notImplemented(method string) error {
	return fmt.Errorf("mock: %s not implemented", method)
}

// Auto-generated mock method implementations

func (m *MockClient) AcceptDispute(ctx context.Context, id string) (*Dispute, error) {
	return nil, m.notImplemented("AcceptDispute")
}

func (m *MockClient) ActivateCoupon(ctx context.Context, id string) (*Coupon, error) {
	return nil, m.notImplemented("ActivateCoupon")
}

func (m *MockClient) ActivateFlashPrice(ctx context.Context, id string) (*FlashPrice, error) {
	return nil, m.notImplemented("ActivateFlashPrice")
}

func (m *MockClient) ActivateGift(ctx context.Context, id string) (*Gift, error) {
	return nil, m.notImplemented("ActivateGift")
}

func (m *MockClient) ActivatePromotion(ctx context.Context, id string) (*Promotion, error) {
	return nil, m.notImplemented("ActivatePromotion")
}

func (m *MockClient) ActivateSale(ctx context.Context, id string) (*Sale, error) {
	return nil, m.notImplemented("ActivateSale")
}

func (m *MockClient) AddProductsToCollection(ctx context.Context, id string, productIDs []string) error {
	return m.notImplemented("AddProductsToCollection")
}

func (m *MockClient) AddWishListItem(ctx context.Context, wishListID string, req *WishListItemCreateRequest) (*WishListItem, error) {
	return nil, m.notImplemented("AddWishListItem")
}

func (m *MockClient) AdjustCompanyCredit(ctx context.Context, id string, req *CompanyCreditAdjustRequest) (*CompanyCredit, error) {
	return nil, m.notImplemented("AdjustCompanyCredit")
}

func (m *MockClient) AdjustInventory(ctx context.Context, id string, delta int) (*InventoryLevel, error) {
	return nil, m.notImplemented("AdjustInventory")
}

func (m *MockClient) AdjustInventoryLevel(ctx context.Context, req *InventoryLevelAdjustRequest) (*InventoryLevel, error) {
	return nil, m.notImplemented("AdjustInventoryLevel")
}

func (m *MockClient) AdjustMemberPoints(ctx context.Context, customerID string, points int, description string) (*MemberPoints, error) {
	return nil, m.notImplemented("AdjustMemberPoints")
}

func (m *MockClient) AssignUserCoupon(ctx context.Context, req *UserCouponAssignRequest) (*UserCoupon, error) {
	return nil, m.notImplemented("AssignUserCoupon")
}

func (m *MockClient) CancelBulkOperation(ctx context.Context, id string) (*BulkOperation, error) {
	return nil, m.notImplemented("CancelBulkOperation")
}

func (m *MockClient) CancelFulfillmentOrder(ctx context.Context, id string) (*FulfillmentOrder, error) {
	return nil, m.notImplemented("CancelFulfillmentOrder")
}

func (m *MockClient) CancelOrder(ctx context.Context, id string) error {
	return m.notImplemented("CancelOrder")
}

func (m *MockClient) CancelPurchaseOrder(ctx context.Context, id string) (*PurchaseOrder, error) {
	return nil, m.notImplemented("CancelPurchaseOrder")
}

func (m *MockClient) CancelReturnOrder(ctx context.Context, id string) error {
	return m.notImplemented("CancelReturnOrder")
}

func (m *MockClient) CapturePayment(ctx context.Context, id string, amount string) (*Payment, error) {
	return nil, m.notImplemented("CapturePayment")
}

func (m *MockClient) CloseFulfillmentOrder(ctx context.Context, id string) (*FulfillmentOrder, error) {
	return nil, m.notImplemented("CloseFulfillmentOrder")
}

func (m *MockClient) CompleteDraftOrder(ctx context.Context, id string) (*DraftOrder, error) {
	return nil, m.notImplemented("CompleteDraftOrder")
}

func (m *MockClient) CompleteReturnOrder(ctx context.Context, id string) (*ReturnOrder, error) {
	return nil, m.notImplemented("CompleteReturnOrder")
}

func (m *MockClient) CreateAddonProduct(ctx context.Context, req *AddonProductCreateRequest) (*AddonProduct, error) {
	return nil, m.notImplemented("CreateAddonProduct")
}

func (m *MockClient) CreateAffiliateCampaign(ctx context.Context, req *AffiliateCampaignCreateRequest) (*AffiliateCampaign, error) {
	return nil, m.notImplemented("CreateAffiliateCampaign")
}

func (m *MockClient) CreateArticle(ctx context.Context, req *ArticleCreateRequest) (*Article, error) {
	return nil, m.notImplemented("CreateArticle")
}

func (m *MockClient) CreateBlog(ctx context.Context, req *BlogCreateRequest) (*Blog, error) {
	return nil, m.notImplemented("CreateBlog")
}

func (m *MockClient) CreateBulkMutation(ctx context.Context, req *BulkOperationMutationRequest) (*BulkOperation, error) {
	return nil, m.notImplemented("CreateBulkMutation")
}

func (m *MockClient) CreateBulkQuery(ctx context.Context, req *BulkOperationCreateRequest) (*BulkOperation, error) {
	return nil, m.notImplemented("CreateBulkQuery")
}

func (m *MockClient) CreateCarrierService(ctx context.Context, req *CarrierServiceCreateRequest) (*CarrierService, error) {
	return nil, m.notImplemented("CreateCarrierService")
}

func (m *MockClient) CreateCatalogPricing(ctx context.Context, req *CatalogPricingCreateRequest) (*CatalogPricing, error) {
	return nil, m.notImplemented("CreateCatalogPricing")
}

func (m *MockClient) CreateCategory(ctx context.Context, req *CategoryCreateRequest) (*Category, error) {
	return nil, m.notImplemented("CreateCategory")
}

func (m *MockClient) CreateChannel(ctx context.Context, req *ChannelCreateRequest) (*Channel, error) {
	return nil, m.notImplemented("CreateChannel")
}

func (m *MockClient) CreateChannelProductPrice(ctx context.Context, channelID, productID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("CreateChannelProductPrice")
}

func (m *MockClient) CreateCollection(ctx context.Context, req *CollectionCreateRequest) (*Collection, error) {
	return nil, m.notImplemented("CreateCollection")
}

func (m *MockClient) CreateCompanyCatalog(ctx context.Context, req *CompanyCatalogCreateRequest) (*CompanyCatalog, error) {
	return nil, m.notImplemented("CreateCompanyCatalog")
}

func (m *MockClient) CreateCompanyCredit(ctx context.Context, req *CompanyCreditCreateRequest) (*CompanyCredit, error) {
	return nil, m.notImplemented("CreateCompanyCredit")
}

func (m *MockClient) CreateConversation(ctx context.Context, req *ConversationCreateRequest) (*Conversation, error) {
	return nil, m.notImplemented("CreateConversation")
}

func (m *MockClient) CreateCoupon(ctx context.Context, req *CouponCreateRequest) (*Coupon, error) {
	return nil, m.notImplemented("CreateCoupon")
}

func (m *MockClient) CreateCustomer(ctx context.Context, req *CustomerCreateRequest) (*Customer, error) {
	return nil, m.notImplemented("CreateCustomer")
}

func (m *MockClient) CreateCustomerAddress(ctx context.Context, customerID string, req *CustomerAddressCreateRequest) (*CustomerAddress, error) {
	return nil, m.notImplemented("CreateCustomerAddress")
}

func (m *MockClient) CreateCustomerBlacklist(ctx context.Context, req *CustomerBlacklistCreateRequest) (*CustomerBlacklist, error) {
	return nil, m.notImplemented("CreateCustomerBlacklist")
}

func (m *MockClient) CreateCustomerGroup(ctx context.Context, req *CustomerGroupCreateRequest) (*CustomerGroup, error) {
	return nil, m.notImplemented("CreateCustomerGroup")
}

func (m *MockClient) CreateCustomerSavedSearch(ctx context.Context, req *CustomerSavedSearchCreateRequest) (*CustomerSavedSearch, error) {
	return nil, m.notImplemented("CreateCustomerSavedSearch")
}

func (m *MockClient) CreateCustomField(ctx context.Context, req *CustomFieldCreateRequest) (*CustomField, error) {
	return nil, m.notImplemented("CreateCustomField")
}

func (m *MockClient) CreateDiscountCode(ctx context.Context, req *DiscountCodeCreateRequest) (*DiscountCode, error) {
	return nil, m.notImplemented("CreateDiscountCode")
}

func (m *MockClient) CreateDomain(ctx context.Context, req *DomainCreateRequest) (*Domain, error) {
	return nil, m.notImplemented("CreateDomain")
}

func (m *MockClient) CreateDraftOrder(ctx context.Context, req *DraftOrderCreateRequest) (*DraftOrder, error) {
	return nil, m.notImplemented("CreateDraftOrder")
}

func (m *MockClient) CreateFile(ctx context.Context, req *FileCreateRequest) (*File, error) {
	return nil, m.notImplemented("CreateFile")
}

func (m *MockClient) CreateFlashPrice(ctx context.Context, req *FlashPriceCreateRequest) (*FlashPrice, error) {
	return nil, m.notImplemented("CreateFlashPrice")
}

func (m *MockClient) CreateFulfillmentService(ctx context.Context, req *FulfillmentServiceCreateRequest) (*FulfillmentService, error) {
	return nil, m.notImplemented("CreateFulfillmentService")
}

func (m *MockClient) CreateGift(ctx context.Context, req *GiftCreateRequest) (*Gift, error) {
	return nil, m.notImplemented("CreateGift")
}

func (m *MockClient) CreateGiftCard(ctx context.Context, req *GiftCardCreateRequest) (*GiftCard, error) {
	return nil, m.notImplemented("CreateGiftCard")
}

func (m *MockClient) CreateLabel(ctx context.Context, req *LabelCreateRequest) (*Label, error) {
	return nil, m.notImplemented("CreateLabel")
}

func (m *MockClient) CreateLocalDeliveryOption(ctx context.Context, req *LocalDeliveryCreateRequest) (*LocalDeliveryOption, error) {
	return nil, m.notImplemented("CreateLocalDeliveryOption")
}

func (m *MockClient) CreateLocation(ctx context.Context, req *LocationCreateRequest) (*Location, error) {
	return nil, m.notImplemented("CreateLocation")
}

func (m *MockClient) CreateMarket(ctx context.Context, req *MarketCreateRequest) (*Market, error) {
	return nil, m.notImplemented("CreateMarket")
}

func (m *MockClient) CreateMarketingEvent(ctx context.Context, req *MarketingEventCreateRequest) (*MarketingEvent, error) {
	return nil, m.notImplemented("CreateMarketingEvent")
}

func (m *MockClient) CreateMedia(ctx context.Context, req *MediaCreateRequest) (*Media, error) {
	return nil, m.notImplemented("CreateMedia")
}

func (m *MockClient) CreateMembershipTier(ctx context.Context, req *MembershipTierCreateRequest) (*MembershipTier, error) {
	return nil, m.notImplemented("CreateMembershipTier")
}

func (m *MockClient) CreateMetafield(ctx context.Context, req *MetafieldCreateRequest) (*Metafield, error) {
	return nil, m.notImplemented("CreateMetafield")
}

func (m *MockClient) CreateMetafieldDefinition(ctx context.Context, req *MetafieldDefinitionCreateRequest) (*MetafieldDefinition, error) {
	return nil, m.notImplemented("CreateMetafieldDefinition")
}

func (m *MockClient) CreateOrderRisk(ctx context.Context, orderID string, req *OrderRiskCreateRequest) (*OrderRisk, error) {
	return nil, m.notImplemented("CreateOrderRisk")
}

func (m *MockClient) CreatePage(ctx context.Context, req *PageCreateRequest) (*Page, error) {
	return nil, m.notImplemented("CreatePage")
}

func (m *MockClient) CreatePickupLocation(ctx context.Context, req *PickupCreateRequest) (*PickupLocation, error) {
	return nil, m.notImplemented("CreatePickupLocation")
}

func (m *MockClient) CreatePriceRule(ctx context.Context, req *PriceRuleCreateRequest) (*PriceRule, error) {
	return nil, m.notImplemented("CreatePriceRule")
}

func (m *MockClient) CreateProductListing(ctx context.Context, productID string) (*ProductListing, error) {
	return nil, m.notImplemented("CreateProductListing")
}

func (m *MockClient) CreateProductReview(ctx context.Context, req *ProductReviewCreateRequest) (*ProductReview, error) {
	return nil, m.notImplemented("CreateProductReview")
}

func (m *MockClient) CreateProductSubscription(ctx context.Context, req *ProductSubscriptionCreateRequest) (*ProductSubscription, error) {
	return nil, m.notImplemented("CreateProductSubscription")
}

func (m *MockClient) CreatePromotion(ctx context.Context, req *PromotionCreateRequest) (*Promotion, error) {
	return nil, m.notImplemented("CreatePromotion")
}

func (m *MockClient) CreatePurchaseOrder(ctx context.Context, req *PurchaseOrderCreateRequest) (*PurchaseOrder, error) {
	return nil, m.notImplemented("CreatePurchaseOrder")
}

func (m *MockClient) CreateRedirect(ctx context.Context, req *RedirectCreateRequest) (*Redirect, error) {
	return nil, m.notImplemented("CreateRedirect")
}

func (m *MockClient) CreateRefund(ctx context.Context, req *RefundCreateRequest) (*Refund, error) {
	return nil, m.notImplemented("CreateRefund")
}

func (m *MockClient) CreateReturnOrder(ctx context.Context, req *ReturnOrderCreateRequest) (*ReturnOrder, error) {
	return nil, m.notImplemented("CreateReturnOrder")
}

func (m *MockClient) CreateSale(ctx context.Context, req *SaleCreateRequest) (*Sale, error) {
	return nil, m.notImplemented("CreateSale")
}

func (m *MockClient) CreateScriptTag(ctx context.Context, req *ScriptTagCreateRequest) (*ScriptTag, error) {
	return nil, m.notImplemented("CreateScriptTag")
}

func (m *MockClient) CreateSellingPlan(ctx context.Context, req *SellingPlanCreateRequest) (*SellingPlan, error) {
	return nil, m.notImplemented("CreateSellingPlan")
}

func (m *MockClient) CreateShipment(ctx context.Context, req *ShipmentCreateRequest) (*Shipment, error) {
	return nil, m.notImplemented("CreateShipment")
}

func (m *MockClient) CreateShippingZone(ctx context.Context, req *ShippingZoneCreateRequest) (*ShippingZone, error) {
	return nil, m.notImplemented("CreateShippingZone")
}

func (m *MockClient) CreateSizeChart(ctx context.Context, req *SizeChartCreateRequest) (*SizeChart, error) {
	return nil, m.notImplemented("CreateSizeChart")
}

func (m *MockClient) CreateSmartCollection(ctx context.Context, req *SmartCollectionCreateRequest) (*SmartCollection, error) {
	return nil, m.notImplemented("CreateSmartCollection")
}

func (m *MockClient) CreateStoreCredit(ctx context.Context, req *StoreCreditCreateRequest) (*StoreCredit, error) {
	return nil, m.notImplemented("CreateStoreCredit")
}

func (m *MockClient) CreateStorefrontCart(ctx context.Context, req *StorefrontCartCreateRequest) (*StorefrontCart, error) {
	return nil, m.notImplemented("CreateStorefrontCart")
}

func (m *MockClient) CreateStorefrontOAuthClient(ctx context.Context, req *StorefrontOAuthClientCreateRequest) (*StorefrontOAuthClient, error) {
	return nil, m.notImplemented("CreateStorefrontOAuthClient")
}

func (m *MockClient) CreateStorefrontToken(ctx context.Context, req *StorefrontTokenCreateRequest) (*StorefrontToken, error) {
	return nil, m.notImplemented("CreateStorefrontToken")
}

func (m *MockClient) CreateSubscription(ctx context.Context, req *SubscriptionCreateRequest) (*Subscription, error) {
	return nil, m.notImplemented("CreateSubscription")
}

func (m *MockClient) CreateTag(ctx context.Context, req *TagCreateRequest) (*Tag, error) {
	return nil, m.notImplemented("CreateTag")
}

func (m *MockClient) CreateTax(ctx context.Context, req *TaxCreateRequest) (*Tax, error) {
	return nil, m.notImplemented("CreateTax")
}

func (m *MockClient) CreateTaxonomy(ctx context.Context, req *TaxonomyCreateRequest) (*Taxonomy, error) {
	return nil, m.notImplemented("CreateTaxonomy")
}

func (m *MockClient) CreateTaxService(ctx context.Context, req *TaxServiceCreateRequest) (*TaxService, error) {
	return nil, m.notImplemented("CreateTaxService")
}

func (m *MockClient) CreateTheme(ctx context.Context, req *ThemeCreateRequest) (*Theme, error) {
	return nil, m.notImplemented("CreateTheme")
}

func (m *MockClient) CreateToken(ctx context.Context, req *TokenCreateRequest) (*Token, error) {
	return nil, m.notImplemented("CreateToken")
}

func (m *MockClient) CreateWarehouse(ctx context.Context, req *WarehouseCreateRequest) (*Warehouse, error) {
	return nil, m.notImplemented("CreateWarehouse")
}

func (m *MockClient) CreateWebhook(ctx context.Context, req *WebhookCreateRequest) (*Webhook, error) {
	return nil, m.notImplemented("CreateWebhook")
}

func (m *MockClient) CreateWishList(ctx context.Context, req *WishListCreateRequest) (*WishList, error) {
	return nil, m.notImplemented("CreateWishList")
}

func (m *MockClient) DeactivateCoupon(ctx context.Context, id string) (*Coupon, error) {
	return nil, m.notImplemented("DeactivateCoupon")
}

func (m *MockClient) DeactivateFlashPrice(ctx context.Context, id string) (*FlashPrice, error) {
	return nil, m.notImplemented("DeactivateFlashPrice")
}

func (m *MockClient) DeactivateGift(ctx context.Context, id string) (*Gift, error) {
	return nil, m.notImplemented("DeactivateGift")
}

func (m *MockClient) DeactivatePromotion(ctx context.Context, id string) (*Promotion, error) {
	return nil, m.notImplemented("DeactivatePromotion")
}

func (m *MockClient) DeactivateSale(ctx context.Context, id string) (*Sale, error) {
	return nil, m.notImplemented("DeactivateSale")
}

func (m *MockClient) Delete(ctx context.Context, path string) error {
	return m.notImplemented("Delete")
}

func (m *MockClient) DeleteAddonProduct(ctx context.Context, id string) error {
	return m.notImplemented("DeleteAddonProduct")
}

func (m *MockClient) DeleteAffiliateCampaign(ctx context.Context, id string) error {
	return m.notImplemented("DeleteAffiliateCampaign")
}

func (m *MockClient) DeleteArticle(ctx context.Context, id string) error {
	return m.notImplemented("DeleteArticle")
}

func (m *MockClient) DeleteAsset(ctx context.Context, themeID, key string) error {
	return m.notImplemented("DeleteAsset")
}

func (m *MockClient) DeleteBlog(ctx context.Context, id string) error {
	return m.notImplemented("DeleteBlog")
}

func (m *MockClient) DeleteCarrierService(ctx context.Context, id string) error {
	return m.notImplemented("DeleteCarrierService")
}

func (m *MockClient) DeleteCatalogPricing(ctx context.Context, id string) error {
	return m.notImplemented("DeleteCatalogPricing")
}

func (m *MockClient) DeleteCategory(ctx context.Context, id string) error {
	return m.notImplemented("DeleteCategory")
}

func (m *MockClient) DeleteChannel(ctx context.Context, id string) error {
	return m.notImplemented("DeleteChannel")
}

func (m *MockClient) DeleteCollection(ctx context.Context, id string) error {
	return m.notImplemented("DeleteCollection")
}

func (m *MockClient) DeleteCompanyCatalog(ctx context.Context, id string) error {
	return m.notImplemented("DeleteCompanyCatalog")
}

func (m *MockClient) DeleteCompanyCredit(ctx context.Context, id string) error {
	return m.notImplemented("DeleteCompanyCredit")
}

func (m *MockClient) DeleteConversation(ctx context.Context, id string) error {
	return m.notImplemented("DeleteConversation")
}

func (m *MockClient) DeleteCoupon(ctx context.Context, id string) error {
	return m.notImplemented("DeleteCoupon")
}

func (m *MockClient) DeleteCustomer(ctx context.Context, id string) error {
	return m.notImplemented("DeleteCustomer")
}

func (m *MockClient) DeleteCustomerAddress(ctx context.Context, customerID, addressID string) error {
	return m.notImplemented("DeleteCustomerAddress")
}

func (m *MockClient) DeleteCustomerBlacklist(ctx context.Context, id string) error {
	return m.notImplemented("DeleteCustomerBlacklist")
}

func (m *MockClient) DeleteCustomerGroup(ctx context.Context, id string) error {
	return m.notImplemented("DeleteCustomerGroup")
}

func (m *MockClient) DeleteCustomerSavedSearch(ctx context.Context, id string) error {
	return m.notImplemented("DeleteCustomerSavedSearch")
}

func (m *MockClient) DeleteCustomField(ctx context.Context, id string) error {
	return m.notImplemented("DeleteCustomField")
}

func (m *MockClient) DeleteDiscountCode(ctx context.Context, id string) error {
	return m.notImplemented("DeleteDiscountCode")
}

func (m *MockClient) DeleteDomain(ctx context.Context, id string) error {
	return m.notImplemented("DeleteDomain")
}

func (m *MockClient) DeleteDraftOrder(ctx context.Context, id string) error {
	return m.notImplemented("DeleteDraftOrder")
}

func (m *MockClient) DeleteFile(ctx context.Context, id string) error {
	return m.notImplemented("DeleteFile")
}

func (m *MockClient) DeleteFlashPrice(ctx context.Context, id string) error {
	return m.notImplemented("DeleteFlashPrice")
}

func (m *MockClient) DeleteFulfillmentService(ctx context.Context, id string) error {
	return m.notImplemented("DeleteFulfillmentService")
}

func (m *MockClient) DeleteGift(ctx context.Context, id string) error {
	return m.notImplemented("DeleteGift")
}

func (m *MockClient) DeleteGiftCard(ctx context.Context, id string) error {
	return m.notImplemented("DeleteGiftCard")
}

func (m *MockClient) DeleteLabel(ctx context.Context, id string) error {
	return m.notImplemented("DeleteLabel")
}

func (m *MockClient) DeleteLocalDeliveryOption(ctx context.Context, id string) error {
	return m.notImplemented("DeleteLocalDeliveryOption")
}

func (m *MockClient) DeleteLocation(ctx context.Context, id string) error {
	return m.notImplemented("DeleteLocation")
}

func (m *MockClient) DeleteMarket(ctx context.Context, id string) error {
	return m.notImplemented("DeleteMarket")
}

func (m *MockClient) DeleteMarketingEvent(ctx context.Context, id string) error {
	return m.notImplemented("DeleteMarketingEvent")
}

func (m *MockClient) DeleteMedia(ctx context.Context, id string) error {
	return m.notImplemented("DeleteMedia")
}

func (m *MockClient) DeleteMembershipTier(ctx context.Context, id string) error {
	return m.notImplemented("DeleteMembershipTier")
}

func (m *MockClient) DeleteMetafield(ctx context.Context, id string) error {
	return m.notImplemented("DeleteMetafield")
}

func (m *MockClient) DeleteMetafieldDefinition(ctx context.Context, id string) error {
	return m.notImplemented("DeleteMetafieldDefinition")
}

func (m *MockClient) DeleteOrderRisk(ctx context.Context, orderID, riskID string) error {
	return m.notImplemented("DeleteOrderRisk")
}

func (m *MockClient) DeletePage(ctx context.Context, id string) error {
	return m.notImplemented("DeletePage")
}

func (m *MockClient) DeletePickupLocation(ctx context.Context, id string) error {
	return m.notImplemented("DeletePickupLocation")
}

func (m *MockClient) DeletePriceRule(ctx context.Context, id string) error {
	return m.notImplemented("DeletePriceRule")
}

func (m *MockClient) DeleteProductListing(ctx context.Context, id string) error {
	return m.notImplemented("DeleteProductListing")
}

func (m *MockClient) DeleteProductReview(ctx context.Context, id string) error {
	return m.notImplemented("DeleteProductReview")
}

func (m *MockClient) DeleteProductSubscription(ctx context.Context, id string) error {
	return m.notImplemented("DeleteProductSubscription")
}

func (m *MockClient) DeletePromotion(ctx context.Context, id string) error {
	return m.notImplemented("DeletePromotion")
}

func (m *MockClient) DeletePurchaseOrder(ctx context.Context, id string) error {
	return m.notImplemented("DeletePurchaseOrder")
}

func (m *MockClient) DeleteRedirect(ctx context.Context, id string) error {
	return m.notImplemented("DeleteRedirect")
}

func (m *MockClient) DeleteSale(ctx context.Context, id string) error {
	return m.notImplemented("DeleteSale")
}

func (m *MockClient) DeleteScriptTag(ctx context.Context, id string) error {
	return m.notImplemented("DeleteScriptTag")
}

func (m *MockClient) DeleteSellingPlan(ctx context.Context, id string) error {
	return m.notImplemented("DeleteSellingPlan")
}

func (m *MockClient) DeleteShipment(ctx context.Context, id string) error {
	return m.notImplemented("DeleteShipment")
}

func (m *MockClient) DeleteShippingZone(ctx context.Context, id string) error {
	return m.notImplemented("DeleteShippingZone")
}

func (m *MockClient) DeleteSizeChart(ctx context.Context, id string) error {
	return m.notImplemented("DeleteSizeChart")
}

func (m *MockClient) DeleteSmartCollection(ctx context.Context, id string) error {
	return m.notImplemented("DeleteSmartCollection")
}

func (m *MockClient) DeleteStaff(ctx context.Context, id string) error {
	return m.notImplemented("DeleteStaff")
}

func (m *MockClient) DeleteStoreCredit(ctx context.Context, id string) error {
	return m.notImplemented("DeleteStoreCredit")
}

func (m *MockClient) DeleteStorefrontCart(ctx context.Context, id string) error {
	return m.notImplemented("DeleteStorefrontCart")
}

func (m *MockClient) DeleteStorefrontOAuthClient(ctx context.Context, id string) error {
	return m.notImplemented("DeleteStorefrontOAuthClient")
}

func (m *MockClient) DeleteStorefrontToken(ctx context.Context, id string) error {
	return m.notImplemented("DeleteStorefrontToken")
}

func (m *MockClient) DeleteSubscription(ctx context.Context, id string) error {
	return m.notImplemented("DeleteSubscription")
}

func (m *MockClient) DeleteTag(ctx context.Context, id string) error {
	return m.notImplemented("DeleteTag")
}

func (m *MockClient) DeleteTax(ctx context.Context, id string) error {
	return m.notImplemented("DeleteTax")
}

func (m *MockClient) DeleteTaxonomy(ctx context.Context, id string) error {
	return m.notImplemented("DeleteTaxonomy")
}

func (m *MockClient) DeleteTaxService(ctx context.Context, id string) error {
	return m.notImplemented("DeleteTaxService")
}

func (m *MockClient) DeleteTheme(ctx context.Context, id string) error {
	return m.notImplemented("DeleteTheme")
}

func (m *MockClient) DeleteToken(ctx context.Context, id string) error {
	return m.notImplemented("DeleteToken")
}

func (m *MockClient) DeleteWarehouse(ctx context.Context, id string) error {
	return m.notImplemented("DeleteWarehouse")
}

func (m *MockClient) DeleteWebhook(ctx context.Context, id string) error {
	return m.notImplemented("DeleteWebhook")
}

func (m *MockClient) DeleteWishList(ctx context.Context, id string) error {
	return m.notImplemented("DeleteWishList")
}

func (m *MockClient) DisableMultipass(ctx context.Context) error {
	return m.notImplemented("DisableMultipass")
}

func (m *MockClient) EnableMultipass(ctx context.Context) (*Multipass, error) {
	return nil, m.notImplemented("EnableMultipass")
}

func (m *MockClient) GenerateMultipassToken(ctx context.Context, req *MultipassTokenRequest) (*MultipassToken, error) {
	return nil, m.notImplemented("GenerateMultipassToken")
}

func (m *MockClient) Get(ctx context.Context, path string, result interface{}) error {
	return m.notImplemented("Get")
}

func (m *MockClient) GetAbandonedCheckout(ctx context.Context, id string) (*AbandonedCheckout, error) {
	return nil, m.notImplemented("GetAbandonedCheckout")
}

func (m *MockClient) GetAddonProduct(ctx context.Context, id string) (*AddonProduct, error) {
	return nil, m.notImplemented("GetAddonProduct")
}

func (m *MockClient) GetAffiliateCampaign(ctx context.Context, id string) (*AffiliateCampaign, error) {
	return nil, m.notImplemented("GetAffiliateCampaign")
}

func (m *MockClient) GetArticle(ctx context.Context, id string) (*Article, error) {
	return nil, m.notImplemented("GetArticle")
}

func (m *MockClient) GetAsset(ctx context.Context, themeID, key string) (*Asset, error) {
	return nil, m.notImplemented("GetAsset")
}

func (m *MockClient) GetBalance(ctx context.Context) (*Balance, error) {
	return nil, m.notImplemented("GetBalance")
}

func (m *MockClient) GetBalanceTransaction(ctx context.Context, id string) (*BalanceTransaction, error) {
	return nil, m.notImplemented("GetBalanceTransaction")
}

func (m *MockClient) GetBlog(ctx context.Context, id string) (*Blog, error) {
	return nil, m.notImplemented("GetBlog")
}

func (m *MockClient) GetBulkOperation(ctx context.Context, id string) (*BulkOperation, error) {
	return nil, m.notImplemented("GetBulkOperation")
}

func (m *MockClient) GetCarrierService(ctx context.Context, id string) (*CarrierService, error) {
	return nil, m.notImplemented("GetCarrierService")
}

func (m *MockClient) GetCatalogPricing(ctx context.Context, id string) (*CatalogPricing, error) {
	return nil, m.notImplemented("GetCatalogPricing")
}

func (m *MockClient) GetCategory(ctx context.Context, id string) (*Category, error) {
	return nil, m.notImplemented("GetCategory")
}

func (m *MockClient) GetCDPEvent(ctx context.Context, id string) (*CDPEvent, error) {
	return nil, m.notImplemented("GetCDPEvent")
}

func (m *MockClient) GetCDPProfile(ctx context.Context, id string) (*CDPCustomerProfile, error) {
	return nil, m.notImplemented("GetCDPProfile")
}

func (m *MockClient) GetCDPSegment(ctx context.Context, id string) (*CDPSegment, error) {
	return nil, m.notImplemented("GetCDPSegment")
}

func (m *MockClient) GetChannel(ctx context.Context, id string) (*Channel, error) {
	return nil, m.notImplemented("GetChannel")
}

func (m *MockClient) GetChannelPrices(ctx context.Context, channelID string) (json.RawMessage, error) {
	return nil, m.notImplemented("GetChannelPrices")
}

func (m *MockClient) GetChannelProductListing(ctx context.Context, channelID, productID string) (*ChannelProductListing, error) {
	return nil, m.notImplemented("GetChannelProductListing")
}

func (m *MockClient) GetCheckoutSettings(ctx context.Context) (*CheckoutSettings, error) {
	return nil, m.notImplemented("GetCheckoutSettings")
}

func (m *MockClient) GetCollection(ctx context.Context, id string) (*Collection, error) {
	return nil, m.notImplemented("GetCollection")
}

func (m *MockClient) GetCompanyCatalog(ctx context.Context, id string) (*CompanyCatalog, error) {
	return nil, m.notImplemented("GetCompanyCatalog")
}

func (m *MockClient) GetCompanyCredit(ctx context.Context, id string) (*CompanyCredit, error) {
	return nil, m.notImplemented("GetCompanyCredit")
}

func (m *MockClient) GetConversation(ctx context.Context, id string) (*Conversation, error) {
	return nil, m.notImplemented("GetConversation")
}

func (m *MockClient) GetCountry(ctx context.Context, code string) (*Country, error) {
	return nil, m.notImplemented("GetCountry")
}

func (m *MockClient) GetCoupon(ctx context.Context, id string) (*Coupon, error) {
	return nil, m.notImplemented("GetCoupon")
}

func (m *MockClient) GetCouponByCode(ctx context.Context, code string) (*Coupon, error) {
	return nil, m.notImplemented("GetCouponByCode")
}

func (m *MockClient) GetCurrency(ctx context.Context, code string) (*Currency, error) {
	return nil, m.notImplemented("GetCurrency")
}

func (m *MockClient) GetCurrentBulkOperation(ctx context.Context) (*BulkOperation, error) {
	return nil, m.notImplemented("GetCurrentBulkOperation")
}

func (m *MockClient) GetCustomer(ctx context.Context, id string) (*Customer, error) {
	return nil, m.notImplemented("GetCustomer")
}

func (m *MockClient) GetLineCustomer(ctx context.Context, lineID string) (*Customer, error) {
	return nil, m.notImplemented("GetLineCustomer")
}

func (m *MockClient) GetCustomerAddress(ctx context.Context, customerID, addressID string) (*CustomerAddress, error) {
	return nil, m.notImplemented("GetCustomerAddress")
}

func (m *MockClient) GetCustomerBlacklist(ctx context.Context, id string) (*CustomerBlacklist, error) {
	return nil, m.notImplemented("GetCustomerBlacklist")
}

func (m *MockClient) GetCustomerGroup(ctx context.Context, id string) (*CustomerGroup, error) {
	return nil, m.notImplemented("GetCustomerGroup")
}

func (m *MockClient) GetCustomerSavedSearch(ctx context.Context, id string) (*CustomerSavedSearch, error) {
	return nil, m.notImplemented("GetCustomerSavedSearch")
}

func (m *MockClient) GetCustomField(ctx context.Context, id string) (*CustomField, error) {
	return nil, m.notImplemented("GetCustomField")
}

func (m *MockClient) GetDeliveryOption(ctx context.Context, id string) (*DeliveryOption, error) {
	return nil, m.notImplemented("GetDeliveryOption")
}

func (m *MockClient) GetDiscountCode(ctx context.Context, id string) (*DiscountCode, error) {
	return nil, m.notImplemented("GetDiscountCode")
}

func (m *MockClient) GetDiscountCodeByCode(ctx context.Context, code string) (*DiscountCode, error) {
	return nil, m.notImplemented("GetDiscountCodeByCode")
}

func (m *MockClient) GetDispute(ctx context.Context, id string) (*Dispute, error) {
	return nil, m.notImplemented("GetDispute")
}

func (m *MockClient) GetDomain(ctx context.Context, id string) (*Domain, error) {
	return nil, m.notImplemented("GetDomain")
}

func (m *MockClient) GetDraftOrder(ctx context.Context, id string) (*DraftOrder, error) {
	return nil, m.notImplemented("GetDraftOrder")
}

func (m *MockClient) GetFile(ctx context.Context, id string) (*File, error) {
	return nil, m.notImplemented("GetFile")
}

func (m *MockClient) GetFlashPrice(ctx context.Context, id string) (*FlashPrice, error) {
	return nil, m.notImplemented("GetFlashPrice")
}

func (m *MockClient) GetFulfillment(ctx context.Context, id string) (*Fulfillment, error) {
	return nil, m.notImplemented("GetFulfillment")
}

func (m *MockClient) GetFulfillmentOrder(ctx context.Context, id string) (*FulfillmentOrder, error) {
	return nil, m.notImplemented("GetFulfillmentOrder")
}

func (m *MockClient) GetFulfillmentService(ctx context.Context, id string) (*FulfillmentService, error) {
	return nil, m.notImplemented("GetFulfillmentService")
}

func (m *MockClient) GetGift(ctx context.Context, id string) (*Gift, error) {
	return nil, m.notImplemented("GetGift")
}

func (m *MockClient) GetGiftCard(ctx context.Context, id string) (*GiftCard, error) {
	return nil, m.notImplemented("GetGiftCard")
}

func (m *MockClient) GetInventoryLevel(ctx context.Context, id string) (*InventoryLevel, error) {
	return nil, m.notImplemented("GetInventoryLevel")
}

func (m *MockClient) GetLabel(ctx context.Context, id string) (*Label, error) {
	return nil, m.notImplemented("GetLabel")
}

func (m *MockClient) GetLocalDeliveryOption(ctx context.Context, id string) (*LocalDeliveryOption, error) {
	return nil, m.notImplemented("GetLocalDeliveryOption")
}

func (m *MockClient) GetLocation(ctx context.Context, id string) (*Location, error) {
	return nil, m.notImplemented("GetLocation")
}

func (m *MockClient) GetMarket(ctx context.Context, id string) (*Market, error) {
	return nil, m.notImplemented("GetMarket")
}

func (m *MockClient) GetMarketingEvent(ctx context.Context, id string) (*MarketingEvent, error) {
	return nil, m.notImplemented("GetMarketingEvent")
}

func (m *MockClient) GetMedia(ctx context.Context, id string) (*Media, error) {
	return nil, m.notImplemented("GetMedia")
}

func (m *MockClient) GetMemberPoints(ctx context.Context, customerID string) (*MemberPoints, error) {
	return nil, m.notImplemented("GetMemberPoints")
}

func (m *MockClient) GetMembershipTier(ctx context.Context, id string) (*MembershipTier, error) {
	return nil, m.notImplemented("GetMembershipTier")
}

func (m *MockClient) GetMerchant(ctx context.Context) (*Merchant, error) {
	return nil, m.notImplemented("GetMerchant")
}

func (m *MockClient) GetMerchantStaff(ctx context.Context, id string) (*MerchantStaff, error) {
	return nil, m.notImplemented("GetMerchantStaff")
}

func (m *MockClient) GetMetafield(ctx context.Context, id string) (*Metafield, error) {
	return nil, m.notImplemented("GetMetafield")
}

func (m *MockClient) GetMetafieldDefinition(ctx context.Context, id string) (*MetafieldDefinition, error) {
	return nil, m.notImplemented("GetMetafieldDefinition")
}

func (m *MockClient) GetMultipass(ctx context.Context) (*Multipass, error) {
	return nil, m.notImplemented("GetMultipass")
}

func (m *MockClient) GetOperationLog(ctx context.Context, id string) (*OperationLog, error) {
	return nil, m.notImplemented("GetOperationLog")
}

func (m *MockClient) GetOrder(ctx context.Context, id string) (*Order, error) {
	return nil, m.notImplemented("GetOrder")
}

func (m *MockClient) GetOrderAttribution(ctx context.Context, orderID string) (*OrderAttribution, error) {
	return nil, m.notImplemented("GetOrderAttribution")
}

func (m *MockClient) GetOrderRisk(ctx context.Context, orderID, riskID string) (*OrderRisk, error) {
	return nil, m.notImplemented("GetOrderRisk")
}

func (m *MockClient) GetPage(ctx context.Context, id string) (*Page, error) {
	return nil, m.notImplemented("GetPage")
}

func (m *MockClient) GetPayment(ctx context.Context, id string) (*Payment, error) {
	return nil, m.notImplemented("GetPayment")
}

func (m *MockClient) GetPayout(ctx context.Context, id string) (*Payout, error) {
	return nil, m.notImplemented("GetPayout")
}

func (m *MockClient) GetPickupLocation(ctx context.Context, id string) (*PickupLocation, error) {
	return nil, m.notImplemented("GetPickupLocation")
}

func (m *MockClient) GetPriceRule(ctx context.Context, id string) (*PriceRule, error) {
	return nil, m.notImplemented("GetPriceRule")
}

func (m *MockClient) GetProduct(ctx context.Context, id string) (*Product, error) {
	return nil, m.notImplemented("GetProduct")
}

func (m *MockClient) GetProductListing(ctx context.Context, id string) (*ProductListing, error) {
	return nil, m.notImplemented("GetProductListing")
}

func (m *MockClient) GetProductReview(ctx context.Context, id string) (*ProductReview, error) {
	return nil, m.notImplemented("GetProductReview")
}

func (m *MockClient) GetProductSubscription(ctx context.Context, id string) (*ProductSubscription, error) {
	return nil, m.notImplemented("GetProductSubscription")
}

func (m *MockClient) GetPromotion(ctx context.Context, id string) (*Promotion, error) {
	return nil, m.notImplemented("GetPromotion")
}

func (m *MockClient) GetPurchaseOrder(ctx context.Context, id string) (*PurchaseOrder, error) {
	return nil, m.notImplemented("GetPurchaseOrder")
}

func (m *MockClient) GetRedirect(ctx context.Context, id string) (*Redirect, error) {
	return nil, m.notImplemented("GetRedirect")
}

func (m *MockClient) GetRefund(ctx context.Context, id string) (*Refund, error) {
	return nil, m.notImplemented("GetRefund")
}

func (m *MockClient) GetReturnOrder(ctx context.Context, id string) (*ReturnOrder, error) {
	return nil, m.notImplemented("GetReturnOrder")
}

func (m *MockClient) GetSale(ctx context.Context, id string) (*Sale, error) {
	return nil, m.notImplemented("GetSale")
}

func (m *MockClient) GetScriptTag(ctx context.Context, id string) (*ScriptTag, error) {
	return nil, m.notImplemented("GetScriptTag")
}

func (m *MockClient) GetSellingPlan(ctx context.Context, id string) (*SellingPlan, error) {
	return nil, m.notImplemented("GetSellingPlan")
}

func (m *MockClient) GetSettings(ctx context.Context) (*SettingsResponse, error) {
	return nil, m.notImplemented("GetSettings")
}

func (m *MockClient) GetShipment(ctx context.Context, id string) (*Shipment, error) {
	return nil, m.notImplemented("GetShipment")
}

func (m *MockClient) GetShippingZone(ctx context.Context, id string) (*ShippingZone, error) {
	return nil, m.notImplemented("GetShippingZone")
}

func (m *MockClient) GetShop(ctx context.Context) (*Shop, error) {
	return nil, m.notImplemented("GetShop")
}

func (m *MockClient) GetShopSettings(ctx context.Context) (*ShopSettings, error) {
	return nil, m.notImplemented("GetShopSettings")
}

func (m *MockClient) GetSizeChart(ctx context.Context, id string) (*SizeChart, error) {
	return nil, m.notImplemented("GetSizeChart")
}

func (m *MockClient) GetSmartCollection(ctx context.Context, id string) (*SmartCollection, error) {
	return nil, m.notImplemented("GetSmartCollection")
}

func (m *MockClient) GetStaff(ctx context.Context, id string) (*Staff, error) {
	return nil, m.notImplemented("GetStaff")
}

func (m *MockClient) GetStoreCredit(ctx context.Context, id string) (*StoreCredit, error) {
	return nil, m.notImplemented("GetStoreCredit")
}

func (m *MockClient) GetStorefrontCart(ctx context.Context, id string) (*StorefrontCart, error) {
	return nil, m.notImplemented("GetStorefrontCart")
}

func (m *MockClient) GetStorefrontOAuthClient(ctx context.Context, id string) (*StorefrontOAuthClient, error) {
	return nil, m.notImplemented("GetStorefrontOAuthClient")
}

func (m *MockClient) GetStorefrontProduct(ctx context.Context, id string) (*StorefrontProduct, error) {
	return nil, m.notImplemented("GetStorefrontProduct")
}

func (m *MockClient) GetStorefrontProductByHandle(ctx context.Context, handle string) (*StorefrontProduct, error) {
	return nil, m.notImplemented("GetStorefrontProductByHandle")
}

func (m *MockClient) GetStorefrontPromotion(ctx context.Context, id string) (*StorefrontPromotion, error) {
	return nil, m.notImplemented("GetStorefrontPromotion")
}

func (m *MockClient) GetStorefrontPromotionByCode(ctx context.Context, code string) (*StorefrontPromotion, error) {
	return nil, m.notImplemented("GetStorefrontPromotionByCode")
}

func (m *MockClient) GetStorefrontToken(ctx context.Context, id string) (*StorefrontToken, error) {
	return nil, m.notImplemented("GetStorefrontToken")
}

func (m *MockClient) GetSubscription(ctx context.Context, id string) (*Subscription, error) {
	return nil, m.notImplemented("GetSubscription")
}

func (m *MockClient) GetTag(ctx context.Context, id string) (*Tag, error) {
	return nil, m.notImplemented("GetTag")
}

func (m *MockClient) GetTax(ctx context.Context, id string) (*Tax, error) {
	return nil, m.notImplemented("GetTax")
}

func (m *MockClient) GetTaxonomy(ctx context.Context, id string) (*Taxonomy, error) {
	return nil, m.notImplemented("GetTaxonomy")
}

func (m *MockClient) GetTaxService(ctx context.Context, id string) (*TaxService, error) {
	return nil, m.notImplemented("GetTaxService")
}

func (m *MockClient) GetTheme(ctx context.Context, id string) (*Theme, error) {
	return nil, m.notImplemented("GetTheme")
}

func (m *MockClient) GetToken(ctx context.Context, id string) (*Token, error) {
	return nil, m.notImplemented("GetToken")
}

func (m *MockClient) GetTransaction(ctx context.Context, id string) (*Transaction, error) {
	return nil, m.notImplemented("GetTransaction")
}

func (m *MockClient) GetUserCoupon(ctx context.Context, id string) (*UserCoupon, error) {
	return nil, m.notImplemented("GetUserCoupon")
}

func (m *MockClient) GetWarehouse(ctx context.Context, id string) (*Warehouse, error) {
	return nil, m.notImplemented("GetWarehouse")
}

func (m *MockClient) GetWebhook(ctx context.Context, id string) (*Webhook, error) {
	return nil, m.notImplemented("GetWebhook")
}

func (m *MockClient) GetWishList(ctx context.Context, id string) (*WishList, error) {
	return nil, m.notImplemented("GetWishList")
}

func (m *MockClient) ListAbandonedCheckouts(ctx context.Context, opts *AbandonedCheckoutsListOptions) (*AbandonedCheckoutsListResponse, error) {
	return nil, m.notImplemented("ListAbandonedCheckouts")
}

func (m *MockClient) ListAddonProducts(ctx context.Context, opts *AddonProductsListOptions) (*AddonProductsListResponse, error) {
	return nil, m.notImplemented("ListAddonProducts")
}

func (m *MockClient) ListAffiliateCampaigns(ctx context.Context, opts *AffiliateCampaignsListOptions) (*AffiliateCampaignsListResponse, error) {
	return nil, m.notImplemented("ListAffiliateCampaigns")
}

func (m *MockClient) ListArticles(ctx context.Context, opts *ArticlesListOptions) (*ArticlesListResponse, error) {
	return nil, m.notImplemented("ListArticles")
}

func (m *MockClient) ListAssets(ctx context.Context, themeID string) (*AssetsListResponse, error) {
	return nil, m.notImplemented("ListAssets")
}

func (m *MockClient) ListBalanceTransactions(ctx context.Context, opts *BalanceTransactionsListOptions) (*BalanceTransactionsListResponse, error) {
	return nil, m.notImplemented("ListBalanceTransactions")
}

func (m *MockClient) ListBlogs(ctx context.Context, opts *BlogsListOptions) (*BlogsListResponse, error) {
	return nil, m.notImplemented("ListBlogs")
}

func (m *MockClient) ListBulkOperations(ctx context.Context, opts *BulkOperationsListOptions) (*BulkOperationsListResponse, error) {
	return nil, m.notImplemented("ListBulkOperations")
}

func (m *MockClient) ListCarrierServices(ctx context.Context, opts *CarrierServicesListOptions) (*CarrierServicesListResponse, error) {
	return nil, m.notImplemented("ListCarrierServices")
}

func (m *MockClient) ListCatalogPricing(ctx context.Context, opts *CatalogPricingListOptions) (*CatalogPricingListResponse, error) {
	return nil, m.notImplemented("ListCatalogPricing")
}

func (m *MockClient) ListCategories(ctx context.Context, opts *CategoriesListOptions) (*CategoriesListResponse, error) {
	return nil, m.notImplemented("ListCategories")
}

func (m *MockClient) ListCDPEvents(ctx context.Context, opts *CDPEventsListOptions) (*CDPEventsListResponse, error) {
	return nil, m.notImplemented("ListCDPEvents")
}

func (m *MockClient) ListCDPProfiles(ctx context.Context, opts *CDPProfilesListOptions) (*CDPProfilesListResponse, error) {
	return nil, m.notImplemented("ListCDPProfiles")
}

func (m *MockClient) ListCDPSegments(ctx context.Context, opts *CDPSegmentsListOptions) (*CDPSegmentsListResponse, error) {
	return nil, m.notImplemented("ListCDPSegments")
}

func (m *MockClient) ListChannelProductListings(ctx context.Context, channelID string, opts *ChannelProductsListOptions) (*ChannelProductsListResponse, error) {
	return nil, m.notImplemented("ListChannelProductListings")
}

func (m *MockClient) ListChannelProducts(ctx context.Context, channelID string, page, pageSize int) (*ChannelProductsResponse, error) {
	return nil, m.notImplemented("ListChannelProducts")
}

func (m *MockClient) ListChannels(ctx context.Context, opts *ChannelsListOptions) (*ChannelsListResponse, error) {
	return nil, m.notImplemented("ListChannels")
}

func (m *MockClient) ListCollections(ctx context.Context, opts *CollectionsListOptions) (*CollectionsListResponse, error) {
	return nil, m.notImplemented("ListCollections")
}

func (m *MockClient) ListCompanyCatalogs(ctx context.Context, opts *CompanyCatalogsListOptions) (*CompanyCatalogsListResponse, error) {
	return nil, m.notImplemented("ListCompanyCatalogs")
}

func (m *MockClient) ListCompanyCredits(ctx context.Context, opts *CompanyCreditsListOptions) (*CompanyCreditsListResponse, error) {
	return nil, m.notImplemented("ListCompanyCredits")
}

func (m *MockClient) ListCompanyCreditTransactions(ctx context.Context, creditID string, page, pageSize int) (*CompanyCreditTransactionsListResponse, error) {
	return nil, m.notImplemented("ListCompanyCreditTransactions")
}

func (m *MockClient) ListConversationMessages(ctx context.Context, conversationID string, page, pageSize int) (*ConversationMessagesListResponse, error) {
	return nil, m.notImplemented("ListConversationMessages")
}

func (m *MockClient) ListConversations(ctx context.Context, opts *ConversationsListOptions) (*ConversationsListResponse, error) {
	return nil, m.notImplemented("ListConversations")
}

func (m *MockClient) ListCountries(ctx context.Context) (*CountriesListResponse, error) {
	return nil, m.notImplemented("ListCountries")
}

func (m *MockClient) ListCoupons(ctx context.Context, opts *CouponsListOptions) (*CouponsListResponse, error) {
	return nil, m.notImplemented("ListCoupons")
}

func (m *MockClient) ListCurrencies(ctx context.Context) (*CurrenciesListResponse, error) {
	return nil, m.notImplemented("ListCurrencies")
}

func (m *MockClient) ListCustomerAddresses(ctx context.Context, customerID string, opts *CustomerAddressesListOptions) (*CustomerAddressesListResponse, error) {
	return nil, m.notImplemented("ListCustomerAddresses")
}

func (m *MockClient) ListCustomerBlacklist(ctx context.Context, opts *CustomerBlacklistListOptions) (*CustomerBlacklistListResponse, error) {
	return nil, m.notImplemented("ListCustomerBlacklist")
}

func (m *MockClient) ListCustomerGroups(ctx context.Context, opts *CustomerGroupsListOptions) (*CustomerGroupsListResponse, error) {
	return nil, m.notImplemented("ListCustomerGroups")
}

func (m *MockClient) ListCustomers(ctx context.Context, opts *CustomersListOptions) (*CustomersListResponse, error) {
	return nil, m.notImplemented("ListCustomers")
}

func (m *MockClient) SearchCustomers(ctx context.Context, opts *CustomerSearchOptions) (*CustomersListResponse, error) {
	return nil, m.notImplemented("SearchCustomers")
}

func (m *MockClient) ListCustomerSavedSearches(ctx context.Context, opts *CustomerSavedSearchesListOptions) (*CustomerSavedSearchesListResponse, error) {
	return nil, m.notImplemented("ListCustomerSavedSearches")
}

func (m *MockClient) ListCustomFields(ctx context.Context, opts *CustomFieldsListOptions) (*CustomFieldsListResponse, error) {
	return nil, m.notImplemented("ListCustomFields")
}

func (m *MockClient) ListDeliveryOptions(ctx context.Context, opts *DeliveryOptionsListOptions) (*DeliveryOptionsListResponse, error) {
	return nil, m.notImplemented("ListDeliveryOptions")
}

func (m *MockClient) ListDeliveryTimeSlots(ctx context.Context, id string, opts *DeliveryTimeSlotsListOptions) (*DeliveryTimeSlotsListResponse, error) {
	return nil, m.notImplemented("ListDeliveryTimeSlots")
}

func (m *MockClient) ListDiscountCodes(ctx context.Context, opts *DiscountCodesListOptions) (*DiscountCodesListResponse, error) {
	return nil, m.notImplemented("ListDiscountCodes")
}

func (m *MockClient) ListDisputes(ctx context.Context, opts *DisputesListOptions) (*DisputesListResponse, error) {
	return nil, m.notImplemented("ListDisputes")
}

func (m *MockClient) ListDomains(ctx context.Context, opts *DomainsListOptions) (*DomainsListResponse, error) {
	return nil, m.notImplemented("ListDomains")
}

func (m *MockClient) ListDraftOrders(ctx context.Context, opts *DraftOrdersListOptions) (*DraftOrdersListResponse, error) {
	return nil, m.notImplemented("ListDraftOrders")
}

func (m *MockClient) ListFiles(ctx context.Context, opts *FilesListOptions) (*FilesListResponse, error) {
	return nil, m.notImplemented("ListFiles")
}

func (m *MockClient) ListFlashPrices(ctx context.Context, opts *FlashPriceListOptions) (*FlashPriceListResponse, error) {
	return nil, m.notImplemented("ListFlashPrices")
}

func (m *MockClient) ListFulfillmentOrders(ctx context.Context, opts *FulfillmentOrdersListOptions) (*FulfillmentOrdersListResponse, error) {
	return nil, m.notImplemented("ListFulfillmentOrders")
}

func (m *MockClient) ListFulfillments(ctx context.Context, opts *FulfillmentsListOptions) (*FulfillmentsListResponse, error) {
	return nil, m.notImplemented("ListFulfillments")
}

func (m *MockClient) ListFulfillmentServices(ctx context.Context, opts *FulfillmentServicesListOptions) (*FulfillmentServicesListResponse, error) {
	return nil, m.notImplemented("ListFulfillmentServices")
}

func (m *MockClient) ListGiftCards(ctx context.Context, opts *GiftCardsListOptions) (*GiftCardsListResponse, error) {
	return nil, m.notImplemented("ListGiftCards")
}

func (m *MockClient) ListGifts(ctx context.Context, opts *GiftsListOptions) (*GiftsListResponse, error) {
	return nil, m.notImplemented("ListGifts")
}

func (m *MockClient) ListInventoryLevels(ctx context.Context, opts *InventoryListOptions) (*InventoryListResponse, error) {
	return nil, m.notImplemented("ListInventoryLevels")
}

func (m *MockClient) ListLabels(ctx context.Context, opts *LabelsListOptions) (*LabelsListResponse, error) {
	return nil, m.notImplemented("ListLabels")
}

func (m *MockClient) ListLocalDeliveryOptions(ctx context.Context, opts *LocalDeliveryListOptions) (*LocalDeliveryListResponse, error) {
	return nil, m.notImplemented("ListLocalDeliveryOptions")
}

func (m *MockClient) ListLocations(ctx context.Context, opts *LocationsListOptions) (*LocationsListResponse, error) {
	return nil, m.notImplemented("ListLocations")
}

func (m *MockClient) ListMarketingEvents(ctx context.Context, opts *MarketingEventsListOptions) (*MarketingEventsListResponse, error) {
	return nil, m.notImplemented("ListMarketingEvents")
}

func (m *MockClient) ListMarkets(ctx context.Context, opts *MarketsListOptions) (*MarketsListResponse, error) {
	return nil, m.notImplemented("ListMarkets")
}

func (m *MockClient) ListMedias(ctx context.Context, opts *MediasListOptions) (*MediasListResponse, error) {
	return nil, m.notImplemented("ListMedias")
}

func (m *MockClient) ListMembershipTiers(ctx context.Context, opts *MembershipTiersListOptions) (*MembershipTiersListResponse, error) {
	return nil, m.notImplemented("ListMembershipTiers")
}

func (m *MockClient) ListMerchants(ctx context.Context) ([]Merchant, error) {
	return nil, m.notImplemented("ListMerchants")
}

func (m *MockClient) ListMerchantStaff(ctx context.Context, opts *MerchantStaffListOptions) (*MerchantStaffListResponse, error) {
	return nil, m.notImplemented("ListMerchantStaff")
}

func (m *MockClient) ListMetafieldDefinitions(ctx context.Context, opts *MetafieldDefinitionsListOptions) (*MetafieldDefinitionsListResponse, error) {
	return nil, m.notImplemented("ListMetafieldDefinitions")
}

func (m *MockClient) ListMetafields(ctx context.Context, opts *MetafieldsListOptions) (*MetafieldsListResponse, error) {
	return nil, m.notImplemented("ListMetafields")
}

func (m *MockClient) ListOperationLogs(ctx context.Context, opts *OperationLogsListOptions) (*OperationLogsListResponse, error) {
	return nil, m.notImplemented("ListOperationLogs")
}

func (m *MockClient) ListOrderAttributions(ctx context.Context, opts *OrderAttributionListOptions) (*OrderAttributionListResponse, error) {
	return nil, m.notImplemented("ListOrderAttributions")
}

func (m *MockClient) ListOrderFulfillmentOrders(ctx context.Context, orderID string) (*FulfillmentOrdersListResponse, error) {
	return nil, m.notImplemented("ListOrderFulfillmentOrders")
}

func (m *MockClient) ListOrderPayments(ctx context.Context, orderID string) (*PaymentsListResponse, error) {
	return nil, m.notImplemented("ListOrderPayments")
}

func (m *MockClient) ListOrderRefunds(ctx context.Context, orderID string) (*RefundsListResponse, error) {
	return nil, m.notImplemented("ListOrderRefunds")
}

func (m *MockClient) ListOrderRisks(ctx context.Context, orderID string, opts *OrderRisksListOptions) (*OrderRisksListResponse, error) {
	return nil, m.notImplemented("ListOrderRisks")
}

func (m *MockClient) ListOrders(ctx context.Context, opts *OrdersListOptions) (*OrdersListResponse, error) {
	return nil, m.notImplemented("ListOrders")
}

func (m *MockClient) ListOrderTransactions(ctx context.Context, orderID string) (*TransactionsListResponse, error) {
	return nil, m.notImplemented("ListOrderTransactions")
}

func (m *MockClient) ListPages(ctx context.Context, opts *PagesListOptions) (*PagesListResponse, error) {
	return nil, m.notImplemented("ListPages")
}

func (m *MockClient) ListPayments(ctx context.Context, opts *PaymentsListOptions) (*PaymentsListResponse, error) {
	return nil, m.notImplemented("ListPayments")
}

func (m *MockClient) ListPayouts(ctx context.Context, opts *PayoutsListOptions) (*PayoutsListResponse, error) {
	return nil, m.notImplemented("ListPayouts")
}

func (m *MockClient) ListPickupLocations(ctx context.Context, opts *PickupListOptions) (*PickupListResponse, error) {
	return nil, m.notImplemented("ListPickupLocations")
}

func (m *MockClient) ListPointsTransactions(ctx context.Context, customerID string, opts *PointsTransactionsListOptions) (*PointsTransactionsListResponse, error) {
	return nil, m.notImplemented("ListPointsTransactions")
}

func (m *MockClient) ListPriceRules(ctx context.Context, opts *PriceRulesListOptions) (*PriceRulesListResponse, error) {
	return nil, m.notImplemented("ListPriceRules")
}

func (m *MockClient) ListProductListings(ctx context.Context, opts *ProductListingsListOptions) (*ProductListingsListResponse, error) {
	return nil, m.notImplemented("ListProductListings")
}

func (m *MockClient) ListProductReviews(ctx context.Context, opts *ProductReviewsListOptions) (*ProductReviewsListResponse, error) {
	return nil, m.notImplemented("ListProductReviews")
}

func (m *MockClient) ListProducts(ctx context.Context, opts *ProductsListOptions) (*ProductsListResponse, error) {
	return nil, m.notImplemented("ListProducts")
}

func (m *MockClient) ListProductSubscriptions(ctx context.Context, opts *ProductSubscriptionsListOptions) (*ProductSubscriptionsListResponse, error) {
	return nil, m.notImplemented("ListProductSubscriptions")
}

func (m *MockClient) ListPromotions(ctx context.Context, opts *PromotionsListOptions) (*PromotionsListResponse, error) {
	return nil, m.notImplemented("ListPromotions")
}

func (m *MockClient) ListPurchaseOrders(ctx context.Context, opts *PurchaseOrdersListOptions) (*PurchaseOrdersListResponse, error) {
	return nil, m.notImplemented("ListPurchaseOrders")
}

func (m *MockClient) ListRedirects(ctx context.Context, opts *RedirectsListOptions) (*RedirectsListResponse, error) {
	return nil, m.notImplemented("ListRedirects")
}

func (m *MockClient) ListRefunds(ctx context.Context, opts *RefundsListOptions) (*RefundsListResponse, error) {
	return nil, m.notImplemented("ListRefunds")
}

func (m *MockClient) ListReturnOrders(ctx context.Context, opts *ReturnOrdersListOptions) (*ReturnOrdersListResponse, error) {
	return nil, m.notImplemented("ListReturnOrders")
}

func (m *MockClient) ListSales(ctx context.Context, opts *SalesListOptions) (*SalesListResponse, error) {
	return nil, m.notImplemented("ListSales")
}

func (m *MockClient) ListScriptTags(ctx context.Context, opts *ScriptTagsListOptions) (*ScriptTagsListResponse, error) {
	return nil, m.notImplemented("ListScriptTags")
}

func (m *MockClient) ListSellingPlans(ctx context.Context, opts *SellingPlansListOptions) (*SellingPlansListResponse, error) {
	return nil, m.notImplemented("ListSellingPlans")
}

func (m *MockClient) ListShipments(ctx context.Context, opts *ShipmentsListOptions) (*ShipmentsListResponse, error) {
	return nil, m.notImplemented("ListShipments")
}

func (m *MockClient) ListShippingZones(ctx context.Context, opts *ShippingZonesListOptions) (*ShippingZonesListResponse, error) {
	return nil, m.notImplemented("ListShippingZones")
}

func (m *MockClient) ListSizeCharts(ctx context.Context, opts *SizeChartsListOptions) (*SizeChartsListResponse, error) {
	return nil, m.notImplemented("ListSizeCharts")
}

func (m *MockClient) ListSmartCollections(ctx context.Context, opts *SmartCollectionsListOptions) (*SmartCollectionsListResponse, error) {
	return nil, m.notImplemented("ListSmartCollections")
}

func (m *MockClient) ListStaffs(ctx context.Context, opts *StaffsListOptions) (*StaffsListResponse, error) {
	return nil, m.notImplemented("ListStaffs")
}

func (m *MockClient) ListStoreCredits(ctx context.Context, opts *StoreCreditsListOptions) (*StoreCreditsListResponse, error) {
	return nil, m.notImplemented("ListStoreCredits")
}

func (m *MockClient) ListStorefrontCarts(ctx context.Context, opts *StorefrontCartsListOptions) (*StorefrontCartsListResponse, error) {
	return nil, m.notImplemented("ListStorefrontCarts")
}

func (m *MockClient) ListStorefrontOAuthClients(ctx context.Context, opts *StorefrontOAuthClientsListOptions) (*StorefrontOAuthClientsListResponse, error) {
	return nil, m.notImplemented("ListStorefrontOAuthClients")
}

func (m *MockClient) ListStorefrontProducts(ctx context.Context, opts *StorefrontProductsListOptions) (*StorefrontProductsListResponse, error) {
	return nil, m.notImplemented("ListStorefrontProducts")
}

func (m *MockClient) ListStorefrontPromotions(ctx context.Context, opts *StorefrontPromotionsListOptions) (*StorefrontPromotionsListResponse, error) {
	return nil, m.notImplemented("ListStorefrontPromotions")
}

func (m *MockClient) ListStorefrontTokens(ctx context.Context, opts *StorefrontTokensListOptions) (*StorefrontTokensListResponse, error) {
	return nil, m.notImplemented("ListStorefrontTokens")
}

func (m *MockClient) ListSubscriptions(ctx context.Context, opts *SubscriptionsListOptions) (*SubscriptionsListResponse, error) {
	return nil, m.notImplemented("ListSubscriptions")
}

func (m *MockClient) ListTags(ctx context.Context, opts *TagsListOptions) (*TagsListResponse, error) {
	return nil, m.notImplemented("ListTags")
}

func (m *MockClient) ListTaxes(ctx context.Context, opts *TaxesListOptions) (*TaxesListResponse, error) {
	return nil, m.notImplemented("ListTaxes")
}

func (m *MockClient) ListTaxonomies(ctx context.Context, opts *TaxonomiesListOptions) (*TaxonomiesListResponse, error) {
	return nil, m.notImplemented("ListTaxonomies")
}

func (m *MockClient) ListTaxServices(ctx context.Context, opts *TaxServicesListOptions) (*TaxServicesListResponse, error) {
	return nil, m.notImplemented("ListTaxServices")
}

func (m *MockClient) ListThemes(ctx context.Context, opts *ThemesListOptions) (*ThemesListResponse, error) {
	return nil, m.notImplemented("ListThemes")
}

func (m *MockClient) ListTokens(ctx context.Context, opts *TokensListOptions) (*TokensListResponse, error) {
	return nil, m.notImplemented("ListTokens")
}

func (m *MockClient) ListTransactions(ctx context.Context, opts *TransactionsListOptions) (*TransactionsListResponse, error) {
	return nil, m.notImplemented("ListTransactions")
}

func (m *MockClient) ListUserCoupons(ctx context.Context, opts *UserCouponsListOptions) (*UserCouponsListResponse, error) {
	return nil, m.notImplemented("ListUserCoupons")
}

func (m *MockClient) ListWarehouses(ctx context.Context, opts *WarehousesListOptions) (*WarehousesListResponse, error) {
	return nil, m.notImplemented("ListWarehouses")
}

func (m *MockClient) ListWebhooks(ctx context.Context, opts *WebhooksListOptions) (*WebhooksListResponse, error) {
	return nil, m.notImplemented("ListWebhooks")
}

func (m *MockClient) ListWishLists(ctx context.Context, opts *WishListsListOptions) (*WishListsListResponse, error) {
	return nil, m.notImplemented("ListWishLists")
}

func (m *MockClient) MoveFulfillmentOrder(ctx context.Context, id string, newLocationID string) (*FulfillmentOrder, error) {
	return nil, m.notImplemented("MoveFulfillmentOrder")
}

func (m *MockClient) Post(ctx context.Context, path string, body, result interface{}) error {
	return m.notImplemented("Post")
}

func (m *MockClient) PublishProductToChannel(ctx context.Context, channelID string, req *ChannelPublishProductRequest) error {
	return m.notImplemented("PublishProductToChannel")
}

func (m *MockClient) PublishProductToChannelListing(ctx context.Context, channelID string, req *ChannelProductPublishRequest) (*ChannelProductListing, error) {
	return nil, m.notImplemented("PublishProductToChannelListing")
}

func (m *MockClient) Put(ctx context.Context, path string, body, result interface{}) error {
	return m.notImplemented("Put")
}

func (m *MockClient) ReceivePurchaseOrder(ctx context.Context, id string) (*PurchaseOrder, error) {
	return nil, m.notImplemented("ReceivePurchaseOrder")
}

func (m *MockClient) ReceiveReturnOrder(ctx context.Context, id string) (*ReturnOrder, error) {
	return nil, m.notImplemented("ReceiveReturnOrder")
}

func (m *MockClient) RefundPayment(ctx context.Context, id string, amount string, reason string) (*Payment, error) {
	return nil, m.notImplemented("RefundPayment")
}

func (m *MockClient) RemoveProductFromCollection(ctx context.Context, id, productID string) error {
	return m.notImplemented("RemoveProductFromCollection")
}

func (m *MockClient) RemoveWishListItem(ctx context.Context, wishListID, itemID string) error {
	return m.notImplemented("RemoveWishListItem")
}

func (m *MockClient) RevokeUserCoupon(ctx context.Context, id string) error {
	return m.notImplemented("RevokeUserCoupon")
}

func (m *MockClient) RotateMultipassSecret(ctx context.Context) (*Multipass, error) {
	return nil, m.notImplemented("RotateMultipassSecret")
}

func (m *MockClient) RotateStorefrontOAuthClientSecret(ctx context.Context, id string) (*StorefrontOAuthClient, error) {
	return nil, m.notImplemented("RotateStorefrontOAuthClientSecret")
}

func (m *MockClient) SendAbandonedCheckoutRecoveryEmail(ctx context.Context, id string) error {
	return m.notImplemented("SendAbandonedCheckoutRecoveryEmail")
}

func (m *MockClient) SendConversationMessage(ctx context.Context, conversationID string, req *ConversationMessageCreateRequest) (*ConversationMessage, error) {
	return nil, m.notImplemented("SendConversationMessage")
}

func (m *MockClient) SendDraftOrderInvoice(ctx context.Context, id string) error {
	return m.notImplemented("SendDraftOrderInvoice")
}

func (m *MockClient) SetDefaultCustomerAddress(ctx context.Context, customerID, addressID string) (*CustomerAddress, error) {
	return nil, m.notImplemented("SetDefaultCustomerAddress")
}

func (m *MockClient) SetCustomerTags(ctx context.Context, id string, tags []string) (*Customer, error) {
	return nil, m.notImplemented("SetCustomerTags")
}

func (m *MockClient) SetInventoryLevel(ctx context.Context, req *InventoryLevelSetRequest) (*InventoryLevel, error) {
	return nil, m.notImplemented("SetInventoryLevel")
}

func (m *MockClient) SubmitDispute(ctx context.Context, id string) (*Dispute, error) {
	return nil, m.notImplemented("SubmitDispute")
}

func (m *MockClient) UnpublishProductFromChannel(ctx context.Context, channelID, productID string) error {
	return m.notImplemented("UnpublishProductFromChannel")
}

func (m *MockClient) UnpublishProductFromChannelListing(ctx context.Context, channelID, productID string) error {
	return m.notImplemented("UnpublishProductFromChannelListing")
}

func (m *MockClient) UpdateAffiliateCampaign(ctx context.Context, id string, req *AffiliateCampaignUpdateRequest) (*AffiliateCampaign, error) {
	return nil, m.notImplemented("UpdateAffiliateCampaign")
}

func (m *MockClient) UpdateArticle(ctx context.Context, id string, req *ArticleUpdateRequest) (*Article, error) {
	return nil, m.notImplemented("UpdateArticle")
}

func (m *MockClient) UpdateAsset(ctx context.Context, themeID string, req *AssetUpdateRequest) (*Asset, error) {
	return nil, m.notImplemented("UpdateAsset")
}

func (m *MockClient) UpdateBlog(ctx context.Context, id string, req *BlogUpdateRequest) (*Blog, error) {
	return nil, m.notImplemented("UpdateBlog")
}

func (m *MockClient) UpdateCarrierService(ctx context.Context, id string, req *CarrierServiceUpdateRequest) (*CarrierService, error) {
	return nil, m.notImplemented("UpdateCarrierService")
}

func (m *MockClient) UpdateCatalogPricing(ctx context.Context, id string, req *CatalogPricingUpdateRequest) (*CatalogPricing, error) {
	return nil, m.notImplemented("UpdateCatalogPricing")
}

func (m *MockClient) UpdateCategory(ctx context.Context, id string, req *CategoryUpdateRequest) (*Category, error) {
	return nil, m.notImplemented("UpdateCategory")
}

func (m *MockClient) BulkUpdateCategoryProductSorting(ctx context.Context, id string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("BulkUpdateCategoryProductSorting")
}

func (m *MockClient) UpdateChannel(ctx context.Context, id string, req *ChannelUpdateRequest) (*Channel, error) {
	return nil, m.notImplemented("UpdateChannel")
}

func (m *MockClient) UpdateChannelProductPrice(ctx context.Context, channelID, productID, priceID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("UpdateChannelProductPrice")
}

func (m *MockClient) UpdateChannelProductListing(ctx context.Context, channelID, productID string, req *ChannelProductUpdateRequest) (*ChannelProductListing, error) {
	return nil, m.notImplemented("UpdateChannelProductListing")
}

func (m *MockClient) UpdateCheckoutSettings(ctx context.Context, req *CheckoutSettingsUpdateRequest) (*CheckoutSettings, error) {
	return nil, m.notImplemented("UpdateCheckoutSettings")
}

func (m *MockClient) UpdateCollection(ctx context.Context, id string, req *CollectionUpdateRequest) (*Collection, error) {
	return nil, m.notImplemented("UpdateCollection")
}

func (m *MockClient) UpdateCompanyCatalog(ctx context.Context, id string, req *CompanyCatalogUpdateRequest) (*CompanyCatalog, error) {
	return nil, m.notImplemented("UpdateCompanyCatalog")
}

func (m *MockClient) UpdateConversation(ctx context.Context, id string, req *ConversationUpdateRequest) (*Conversation, error) {
	return nil, m.notImplemented("UpdateConversation")
}

func (m *MockClient) UpdateCoupon(ctx context.Context, id string, req *CouponUpdateRequest) (*Coupon, error) {
	return nil, m.notImplemented("UpdateCoupon")
}

func (m *MockClient) UpdateCurrency(ctx context.Context, code string, req *CurrencyUpdateRequest) (*Currency, error) {
	return nil, m.notImplemented("UpdateCurrency")
}

func (m *MockClient) UpdateCustomerGroup(ctx context.Context, id string, req *CustomerGroupUpdateRequest) (*CustomerGroup, error) {
	return nil, m.notImplemented("UpdateCustomerGroup")
}

func (m *MockClient) UpdateCustomer(ctx context.Context, id string, req *CustomerUpdateRequest) (*Customer, error) {
	return nil, m.notImplemented("UpdateCustomer")
}

func (m *MockClient) UpdateCustomerTags(ctx context.Context, id string, req *CustomerTagsUpdateRequest) (*Customer, error) {
	return nil, m.notImplemented("UpdateCustomerTags")
}

func (m *MockClient) UpdateCustomerSubscriptions(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("UpdateCustomerSubscriptions")
}

func (m *MockClient) UpdateCustomField(ctx context.Context, id string, req *CustomFieldUpdateRequest) (*CustomField, error) {
	return nil, m.notImplemented("UpdateCustomField")
}

func (m *MockClient) UpdateDeliveryOptionPickupStore(ctx context.Context, id string, req *PickupStoreUpdateRequest) (*DeliveryOption, error) {
	return nil, m.notImplemented("UpdateDeliveryOptionPickupStore")
}

func (m *MockClient) UpdateDisputeEvidence(ctx context.Context, id string, req *DisputeUpdateEvidenceRequest) (*Dispute, error) {
	return nil, m.notImplemented("UpdateDisputeEvidence")
}

func (m *MockClient) UpdateDomain(ctx context.Context, id string, req *DomainUpdateRequest) (*Domain, error) {
	return nil, m.notImplemented("UpdateDomain")
}

func (m *MockClient) UpdateFile(ctx context.Context, id string, req *FileUpdateRequest) (*File, error) {
	return nil, m.notImplemented("UpdateFile")
}

func (m *MockClient) UpdateFlashPrice(ctx context.Context, id string, req *FlashPriceUpdateRequest) (*FlashPrice, error) {
	return nil, m.notImplemented("UpdateFlashPrice")
}

func (m *MockClient) UpdateFulfillmentService(ctx context.Context, id string, req *FulfillmentServiceUpdateRequest) (*FulfillmentService, error) {
	return nil, m.notImplemented("UpdateFulfillmentService")
}

func (m *MockClient) UpdateLabel(ctx context.Context, id string, req *LabelUpdateRequest) (*Label, error) {
	return nil, m.notImplemented("UpdateLabel")
}

func (m *MockClient) UpdateLocalDeliveryOption(ctx context.Context, id string, req *LocalDeliveryUpdateRequest) (*LocalDeliveryOption, error) {
	return nil, m.notImplemented("UpdateLocalDeliveryOption")
}

func (m *MockClient) UpdateLocation(ctx context.Context, id string, req *LocationUpdateRequest) (*Location, error) {
	return nil, m.notImplemented("UpdateLocation")
}

func (m *MockClient) UpdateMarketingEvent(ctx context.Context, id string, req *MarketingEventUpdateRequest) (*MarketingEvent, error) {
	return nil, m.notImplemented("UpdateMarketingEvent")
}

func (m *MockClient) UpdateMedia(ctx context.Context, id string, req *MediaUpdateRequest) (*Media, error) {
	return nil, m.notImplemented("UpdateMedia")
}

func (m *MockClient) UpdateMetafield(ctx context.Context, id string, req *MetafieldUpdateRequest) (*Metafield, error) {
	return nil, m.notImplemented("UpdateMetafield")
}

func (m *MockClient) UpdateMetafieldDefinition(ctx context.Context, id string, req *MetafieldDefinitionUpdateRequest) (*MetafieldDefinition, error) {
	return nil, m.notImplemented("UpdateMetafieldDefinition")
}

func (m *MockClient) UpdateOrderRisk(ctx context.Context, orderID, riskID string, req *OrderRiskUpdateRequest) (*OrderRisk, error) {
	return nil, m.notImplemented("UpdateOrderRisk")
}

func (m *MockClient) UpdatePage(ctx context.Context, id string, req *PageUpdateRequest) (*Page, error) {
	return nil, m.notImplemented("UpdatePage")
}

func (m *MockClient) UpdatePickupLocation(ctx context.Context, id string, req *PickupUpdateRequest) (*PickupLocation, error) {
	return nil, m.notImplemented("UpdatePickupLocation")
}

func (m *MockClient) UpdatePriceRule(ctx context.Context, id string, req *PriceRuleUpdateRequest) (*PriceRule, error) {
	return nil, m.notImplemented("UpdatePriceRule")
}

func (m *MockClient) UpdateRedirect(ctx context.Context, id string, req *RedirectUpdateRequest) (*Redirect, error) {
	return nil, m.notImplemented("UpdateRedirect")
}

func (m *MockClient) UpdateReturnOrder(ctx context.Context, id string, req *ReturnOrderUpdateRequest) (*ReturnOrder, error) {
	return nil, m.notImplemented("UpdateReturnOrder")
}

func (m *MockClient) UpdateScriptTag(ctx context.Context, id string, req *ScriptTagUpdateRequest) (*ScriptTag, error) {
	return nil, m.notImplemented("UpdateScriptTag")
}

func (m *MockClient) UpdateSettings(ctx context.Context, req *UserSettingsUpdateRequest) (*SettingsResponse, error) {
	return nil, m.notImplemented("UpdateSettings")
}

func (m *MockClient) UpdateShipment(ctx context.Context, id string, req *ShipmentUpdateRequest) (*Shipment, error) {
	return nil, m.notImplemented("UpdateShipment")
}

func (m *MockClient) UpdateShopSettings(ctx context.Context, req *ShopSettingsUpdateRequest) (*ShopSettings, error) {
	return nil, m.notImplemented("UpdateShopSettings")
}

func (m *MockClient) UpdateSizeChart(ctx context.Context, id string, req *SizeChartUpdateRequest) (*SizeChart, error) {
	return nil, m.notImplemented("UpdateSizeChart")
}

func (m *MockClient) UpdateSmartCollection(ctx context.Context, id string, req *SmartCollectionUpdateRequest) (*SmartCollection, error) {
	return nil, m.notImplemented("UpdateSmartCollection")
}

func (m *MockClient) UpdateStaff(ctx context.Context, id string, req *StaffUpdateRequest) (*Staff, error) {
	return nil, m.notImplemented("UpdateStaff")
}

func (m *MockClient) UpdateStorefrontOAuthClient(ctx context.Context, id string, req *StorefrontOAuthClientUpdateRequest) (*StorefrontOAuthClient, error) {
	return nil, m.notImplemented("UpdateStorefrontOAuthClient")
}

func (m *MockClient) UpdateTax(ctx context.Context, id string, req *TaxUpdateRequest) (*Tax, error) {
	return nil, m.notImplemented("UpdateTax")
}

func (m *MockClient) UpdateTaxonomy(ctx context.Context, id string, req *TaxonomyUpdateRequest) (*Taxonomy, error) {
	return nil, m.notImplemented("UpdateTaxonomy")
}

func (m *MockClient) UpdateTaxService(ctx context.Context, id string, req *TaxServiceUpdateRequest) (*TaxService, error) {
	return nil, m.notImplemented("UpdateTaxService")
}

func (m *MockClient) UpdateTheme(ctx context.Context, id string, req *ThemeUpdateRequest) (*Theme, error) {
	return nil, m.notImplemented("UpdateTheme")
}

func (m *MockClient) UpdateWarehouse(ctx context.Context, id string, req *WarehouseUpdateRequest) (*Warehouse, error) {
	return nil, m.notImplemented("UpdateWarehouse")
}

func (m *MockClient) VerifyDomain(ctx context.Context, id string) (*Domain, error) {
	return nil, m.notImplemented("VerifyDomain")
}

func (m *MockClient) VoidPayment(ctx context.Context, id string) (*Payment, error) {
	return nil, m.notImplemented("VoidPayment")
}

func (m *MockClient) ListOrderMetafields(ctx context.Context, orderID string) (json.RawMessage, error) {
	return nil, m.notImplemented("ListOrderMetafields")
}

func (m *MockClient) GetOrderMetafield(ctx context.Context, orderID, metafieldID string) (json.RawMessage, error) {
	return nil, m.notImplemented("GetOrderMetafield")
}

func (m *MockClient) CreateOrderMetafield(ctx context.Context, orderID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("CreateOrderMetafield")
}

func (m *MockClient) UpdateOrderMetafield(ctx context.Context, orderID, metafieldID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("UpdateOrderMetafield")
}

func (m *MockClient) DeleteOrderMetafield(ctx context.Context, orderID, metafieldID string) error {
	return m.notImplemented("DeleteOrderMetafield")
}

func (m *MockClient) BulkCreateOrderMetafields(ctx context.Context, orderID string, body any) error {
	return m.notImplemented("BulkCreateOrderMetafields")
}

func (m *MockClient) BulkUpdateOrderMetafields(ctx context.Context, orderID string, body any) error {
	return m.notImplemented("BulkUpdateOrderMetafields")
}

func (m *MockClient) BulkDeleteOrderMetafields(ctx context.Context, orderID string, body any) error {
	return m.notImplemented("BulkDeleteOrderMetafields")
}

func (m *MockClient) ListOrderAppMetafields(ctx context.Context, orderID string) (json.RawMessage, error) {
	return nil, m.notImplemented("ListOrderAppMetafields")
}

func (m *MockClient) GetOrderAppMetafield(ctx context.Context, orderID, metafieldID string) (json.RawMessage, error) {
	return nil, m.notImplemented("GetOrderAppMetafield")
}

func (m *MockClient) CreateOrderAppMetafield(ctx context.Context, orderID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("CreateOrderAppMetafield")
}

func (m *MockClient) UpdateOrderAppMetafield(ctx context.Context, orderID, metafieldID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("UpdateOrderAppMetafield")
}

func (m *MockClient) DeleteOrderAppMetafield(ctx context.Context, orderID, metafieldID string) error {
	return m.notImplemented("DeleteOrderAppMetafield")
}

func (m *MockClient) BulkCreateOrderAppMetafields(ctx context.Context, orderID string, body any) error {
	return m.notImplemented("BulkCreateOrderAppMetafields")
}

func (m *MockClient) BulkUpdateOrderAppMetafields(ctx context.Context, orderID string, body any) error {
	return m.notImplemented("BulkUpdateOrderAppMetafields")
}

func (m *MockClient) BulkDeleteOrderAppMetafields(ctx context.Context, orderID string, body any) error {
	return m.notImplemented("BulkDeleteOrderAppMetafields")
}

func (m *MockClient) ListOrderItemMetafields(ctx context.Context, orderID string) (json.RawMessage, error) {
	return nil, m.notImplemented("ListOrderItemMetafields")
}

func (m *MockClient) BulkCreateOrderItemMetafields(ctx context.Context, orderID string, body any) error {
	return m.notImplemented("BulkCreateOrderItemMetafields")
}

func (m *MockClient) BulkUpdateOrderItemMetafields(ctx context.Context, orderID string, body any) error {
	return m.notImplemented("BulkUpdateOrderItemMetafields")
}

func (m *MockClient) BulkDeleteOrderItemMetafields(ctx context.Context, orderID string, body any) error {
	return m.notImplemented("BulkDeleteOrderItemMetafields")
}

func (m *MockClient) ListOrderItemAppMetafields(ctx context.Context, orderID string) (json.RawMessage, error) {
	return nil, m.notImplemented("ListOrderItemAppMetafields")
}

func (m *MockClient) BulkCreateOrderItemAppMetafields(ctx context.Context, orderID string, body any) error {
	return m.notImplemented("BulkCreateOrderItemAppMetafields")
}

func (m *MockClient) BulkUpdateOrderItemAppMetafields(ctx context.Context, orderID string, body any) error {
	return m.notImplemented("BulkUpdateOrderItemAppMetafields")
}

func (m *MockClient) BulkDeleteOrderItemAppMetafields(ctx context.Context, orderID string, body any) error {
	return m.notImplemented("BulkDeleteOrderItemAppMetafields")
}

func (m *MockClient) ListCustomerMetafields(ctx context.Context, customerID string) (json.RawMessage, error) {
	return nil, m.notImplemented("ListCustomerMetafields")
}

func (m *MockClient) GetCustomerMetafield(ctx context.Context, customerID, metafieldID string) (json.RawMessage, error) {
	return nil, m.notImplemented("GetCustomerMetafield")
}

func (m *MockClient) CreateCustomerMetafield(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("CreateCustomerMetafield")
}

func (m *MockClient) UpdateCustomerMetafield(ctx context.Context, customerID, metafieldID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("UpdateCustomerMetafield")
}

func (m *MockClient) DeleteCustomerMetafield(ctx context.Context, customerID, metafieldID string) error {
	return m.notImplemented("DeleteCustomerMetafield")
}

func (m *MockClient) BulkCreateCustomerMetafields(ctx context.Context, customerID string, body any) error {
	return m.notImplemented("BulkCreateCustomerMetafields")
}

func (m *MockClient) BulkUpdateCustomerMetafields(ctx context.Context, customerID string, body any) error {
	return m.notImplemented("BulkUpdateCustomerMetafields")
}

func (m *MockClient) BulkDeleteCustomerMetafields(ctx context.Context, customerID string, body any) error {
	return m.notImplemented("BulkDeleteCustomerMetafields")
}

func (m *MockClient) ListCustomerAppMetafields(ctx context.Context, customerID string) (json.RawMessage, error) {
	return nil, m.notImplemented("ListCustomerAppMetafields")
}

func (m *MockClient) GetCustomerAppMetafield(ctx context.Context, customerID, metafieldID string) (json.RawMessage, error) {
	return nil, m.notImplemented("GetCustomerAppMetafield")
}

func (m *MockClient) CreateCustomerAppMetafield(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("CreateCustomerAppMetafield")
}

func (m *MockClient) UpdateCustomerAppMetafield(ctx context.Context, customerID, metafieldID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("UpdateCustomerAppMetafield")
}

func (m *MockClient) DeleteCustomerAppMetafield(ctx context.Context, customerID, metafieldID string) error {
	return m.notImplemented("DeleteCustomerAppMetafield")
}

func (m *MockClient) BulkCreateCustomerAppMetafields(ctx context.Context, customerID string, body any) error {
	return m.notImplemented("BulkCreateCustomerAppMetafields")
}

func (m *MockClient) BulkUpdateCustomerAppMetafields(ctx context.Context, customerID string, body any) error {
	return m.notImplemented("BulkUpdateCustomerAppMetafields")
}

func (m *MockClient) BulkDeleteCustomerAppMetafields(ctx context.Context, customerID string, body any) error {
	return m.notImplemented("BulkDeleteCustomerAppMetafields")
}

func (m *MockClient) GetCustomerStoreCredits(ctx context.Context, customerID string) (json.RawMessage, error) {
	return nil, m.notImplemented("GetCustomerStoreCredits")
}

func (m *MockClient) CreateCustomerStoreCredits(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("CreateCustomerStoreCredits")
}

func (m *MockClient) ListUserCredits(ctx context.Context, opts *UserCreditsListOptions) (json.RawMessage, error) {
	return nil, m.notImplemented("ListUserCredits")
}

func (m *MockClient) BulkUpdateUserCredits(ctx context.Context, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("BulkUpdateUserCredits")
}

func (m *MockClient) GetCustomersMembershipInfo(ctx context.Context) (json.RawMessage, error) {
	return nil, m.notImplemented("GetCustomersMembershipInfo")
}

func (m *MockClient) GetCustomerMemberPointsHistory(ctx context.Context, customerID string) (json.RawMessage, error) {
	return nil, m.notImplemented("GetCustomerMemberPointsHistory")
}

func (m *MockClient) UpdateCustomerMemberPoints(ctx context.Context, customerID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("UpdateCustomerMemberPoints")
}

func (m *MockClient) GetCustomerMembershipTierActionLogs(ctx context.Context, customerID string) (json.RawMessage, error) {
	return nil, m.notImplemented("GetCustomerMembershipTierActionLogs")
}

func (m *MockClient) ListMemberPointRules(ctx context.Context) (json.RawMessage, error) {
	return nil, m.notImplemented("ListMemberPointRules")
}

func (m *MockClient) BulkUpdateMemberPoints(ctx context.Context, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("BulkUpdateMemberPoints")
}

func (m *MockClient) ExchangeCart(ctx context.Context, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("ExchangeCart")
}

func (m *MockClient) PrepareCart(ctx context.Context, cartID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("PrepareCart")
}

func (m *MockClient) AddCartItems(ctx context.Context, cartID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("AddCartItems")
}

func (m *MockClient) UpdateCartItems(ctx context.Context, cartID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("UpdateCartItems")
}

func (m *MockClient) DeleteCartItems(ctx context.Context, cartID string, body any) (json.RawMessage, error) {
	return nil, m.notImplemented("DeleteCartItems")
}

func (m *MockClient) ListCartItemMetafields(ctx context.Context, cartID string) (json.RawMessage, error) {
	return nil, m.notImplemented("ListCartItemMetafields")
}

func (m *MockClient) BulkCreateCartItemMetafields(ctx context.Context, cartID string, body any) error {
	return m.notImplemented("BulkCreateCartItemMetafields")
}

func (m *MockClient) BulkUpdateCartItemMetafields(ctx context.Context, cartID string, body any) error {
	return m.notImplemented("BulkUpdateCartItemMetafields")
}

func (m *MockClient) BulkDeleteCartItemMetafields(ctx context.Context, cartID string, body any) error {
	return m.notImplemented("BulkDeleteCartItemMetafields")
}

func (m *MockClient) ListCartItemAppMetafields(ctx context.Context, cartID string) (json.RawMessage, error) {
	return nil, m.notImplemented("ListCartItemAppMetafields")
}

func (m *MockClient) BulkCreateCartItemAppMetafields(ctx context.Context, cartID string, body any) error {
	return m.notImplemented("BulkCreateCartItemAppMetafields")
}

func (m *MockClient) BulkUpdateCartItemAppMetafields(ctx context.Context, cartID string, body any) error {
	return m.notImplemented("BulkUpdateCartItemAppMetafields")
}

func (m *MockClient) BulkDeleteCartItemAppMetafields(ctx context.Context, cartID string, body any) error {
	return m.notImplemented("BulkDeleteCartItemAppMetafields")
}

func (m *MockClient) GetStaffPermissions(ctx context.Context, staffID string) (json.RawMessage, error) {
	return nil, m.notImplemented("GetStaffPermissions")
}
