//go:build integration

package api

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

// TestProductionAPI runs integration tests against the production Shopline API.
// Run with: go test -tags=integration -run TestProductionAPI ./internal/api/...
func TestProductionAPI(t *testing.T) {
	token := os.Getenv("SHOPLINE_API_TOKEN")
	if token == "" {
		t.Skip("SHOPLINE_API_TOKEN not set, skipping production tests")
	}

	client := NewOpenAPIClient(token)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// Track results
	type result struct {
		name    string
		success bool
		count   int
		err     error
	}
	results := []result{}

	// Store IDs for later tests
	var productID, orderID, customerID string

	// ============ PRODUCTS API ============
	t.Run("Products", func(t *testing.T) {
		t.Run("ListProducts", func(t *testing.T) {
			resp, err := client.ListProducts(ctx, &ProductsListOptions{PageSize: 5})
			if err != nil {
				results = append(results, result{"ListProducts", false, 0, err})
				t.Errorf("ListProducts failed: %v", err)
				return
			}
			if len(resp.Items) > 0 {
				productID = resp.Items[0].ID
			}
			results = append(results, result{"ListProducts", true, len(resp.Items), nil})
			t.Logf("✓ ListProducts: %d products returned", len(resp.Items))
		})

		t.Run("GetProduct", func(t *testing.T) {
			if productID == "" {
				t.Skip("No product ID available")
			}
			product, err := client.GetProduct(ctx, productID)
			if err != nil {
				results = append(results, result{"GetProduct", false, 0, err})
				t.Errorf("GetProduct failed: %v", err)
				return
			}
			results = append(results, result{"GetProduct", true, 1, nil})
			t.Logf("✓ GetProduct: %s (%s)", product.Title, product.ID)
		})

		t.Run("SearchProducts", func(t *testing.T) {
			resp, err := client.SearchProducts(ctx, &ProductSearchOptions{PageSize: 5})
			if err != nil {
				results = append(results, result{"SearchProducts", false, 0, err})
				t.Errorf("SearchProducts failed: %v", err)
				return
			}
			results = append(results, result{"SearchProducts", true, len(resp.Items), nil})
			t.Logf("✓ SearchProducts: %d products returned", len(resp.Items))
		})
	})

	// ============ ORDERS API ============
	t.Run("Orders", func(t *testing.T) {
		t.Run("ListOrders", func(t *testing.T) {
			resp, err := client.ListOrders(ctx, &OrdersListOptions{PageSize: 5})
			if err != nil {
				results = append(results, result{"ListOrders", false, 0, err})
				t.Errorf("ListOrders failed: %v", err)
				return
			}
			if len(resp.Items) > 0 {
				orderID = resp.Items[0].ID
			}
			results = append(results, result{"ListOrders", true, len(resp.Items), nil})
			t.Logf("✓ ListOrders: %d orders returned", len(resp.Items))
		})

		t.Run("GetOrder", func(t *testing.T) {
			if orderID == "" {
				t.Skip("No order ID available")
			}
			order, err := client.GetOrder(ctx, orderID)
			if err != nil {
				results = append(results, result{"GetOrder", false, 0, err})
				t.Errorf("GetOrder failed: %v", err)
				return
			}
			results = append(results, result{"GetOrder", true, 1, nil})
			t.Logf("✓ GetOrder: %s (%s)", order.OrderNumber, order.ID)
		})

		t.Run("GetOrderTags", func(t *testing.T) {
			if orderID == "" {
				t.Skip("No order ID available")
			}
			tags, err := client.GetOrderTags(ctx, orderID)
			if err != nil {
				results = append(results, result{"GetOrderTags", false, 0, err})
				t.Errorf("GetOrderTags failed: %v", err)
				return
			}
			results = append(results, result{"GetOrderTags", true, len(tags.Tags), nil})
			t.Logf("✓ GetOrderTags: %d tags", len(tags.Tags))
		})

		t.Run("SearchOrders", func(t *testing.T) {
			resp, err := client.SearchOrders(ctx, &OrderSearchOptions{PageSize: 5})
			if err != nil {
				results = append(results, result{"SearchOrders", false, 0, err})
				t.Errorf("SearchOrders failed: %v", err)
				return
			}
			results = append(results, result{"SearchOrders", true, len(resp.Items), nil})
			t.Logf("✓ SearchOrders: %d orders returned", len(resp.Items))
		})

		t.Run("ListArchivedOrders", func(t *testing.T) {
			resp, err := client.ListArchivedOrders(ctx, &ArchivedOrdersListOptions{PageSize: 5})
			if err != nil {
				results = append(results, result{"ListArchivedOrders", false, 0, err})
				// This often returns 410 Gone - mark as expected
				t.Logf("⚠ ListArchivedOrders: %v (may be deprecated)", err)
				return
			}
			results = append(results, result{"ListArchivedOrders", true, len(resp.Items), nil})
			t.Logf("✓ ListArchivedOrders: %d orders returned", len(resp.Items))
		})
	})

	// ============ CUSTOMERS API ============
	t.Run("Customers", func(t *testing.T) {
		t.Run("ListCustomers", func(t *testing.T) {
			resp, err := client.ListCustomers(ctx, &CustomersListOptions{PageSize: 5})
			if err != nil {
				results = append(results, result{"ListCustomers", false, 0, err})
				t.Errorf("ListCustomers failed: %v", err)
				return
			}
			if len(resp.Items) > 0 {
				customerID = resp.Items[0].ID
			}
			results = append(results, result{"ListCustomers", true, len(resp.Items), nil})
			t.Logf("✓ ListCustomers: %d customers returned", len(resp.Items))
		})

		t.Run("GetCustomer", func(t *testing.T) {
			if customerID == "" {
				t.Skip("No customer ID available")
			}
			customer, err := client.GetCustomer(ctx, customerID)
			if err != nil {
				results = append(results, result{"GetCustomer", false, 0, err})
				t.Errorf("GetCustomer failed: %v", err)
				return
			}
			results = append(results, result{"GetCustomer", true, 1, nil})
			t.Logf("✓ GetCustomer: %s %s (%s)", customer.FirstName, customer.LastName, customer.ID)
		})

		t.Run("GetCustomerPromotions", func(t *testing.T) {
			if customerID == "" {
				t.Skip("No customer ID available")
			}
			promos, err := client.GetCustomerPromotions(ctx, customerID)
			if err != nil {
				results = append(results, result{"GetCustomerPromotions", false, 0, err})
				t.Errorf("GetCustomerPromotions failed: %v", err)
				return
			}
			results = append(results, result{"GetCustomerPromotions", true, len(promos.Items), nil})
			t.Logf("✓ GetCustomerPromotions: %d promotions", len(promos.Items))
		})

		t.Run("SearchCustomers", func(t *testing.T) {
			resp, err := client.SearchCustomers(ctx, &CustomerSearchOptions{PageSize: 5})
			if err != nil {
				results = append(results, result{"SearchCustomers", false, 0, err})
				t.Errorf("SearchCustomers failed: %v", err)
				return
			}
			results = append(results, result{"SearchCustomers", true, len(resp.Items), nil})
			t.Logf("✓ SearchCustomers: %d customers returned", len(resp.Items))
		})
	})

	// ============ PROMOTIONS API ============
	var promotionID string
	t.Run("Promotions", func(t *testing.T) {
		t.Run("ListPromotions", func(t *testing.T) {
			resp, err := client.ListPromotions(ctx, &PromotionsListOptions{PageSize: 5})
			if err != nil {
				results = append(results, result{"ListPromotions", false, 0, err})
				t.Errorf("ListPromotions failed: %v", err)
				return
			}
			if len(resp.Items) > 0 {
				promotionID = resp.Items[0].ID
			}
			results = append(results, result{"ListPromotions", true, len(resp.Items), nil})
			t.Logf("✓ ListPromotions: %d promotions returned", len(resp.Items))
		})

		t.Run("GetPromotion", func(t *testing.T) {
			if promotionID == "" {
				t.Skip("No promotion ID available")
			}
			promo, err := client.GetPromotion(ctx, promotionID)
			if err != nil {
				results = append(results, result{"GetPromotion", false, 0, err})
				t.Errorf("GetPromotion failed: %v", err)
				return
			}
			results = append(results, result{"GetPromotion", true, 1, nil})
			t.Logf("✓ GetPromotion: %s (%s)", promo.Title, promo.ID)
		})

		t.Run("SearchPromotions", func(t *testing.T) {
			resp, err := client.SearchPromotions(ctx, &PromotionSearchOptions{PageSize: 5})
			if err != nil {
				results = append(results, result{"SearchPromotions", false, 0, err})
				t.Errorf("SearchPromotions failed: %v", err)
				return
			}
			results = append(results, result{"SearchPromotions", true, len(resp.Items), nil})
			t.Logf("✓ SearchPromotions: %d promotions returned", len(resp.Items))
		})
	})

	// ============ CUSTOMER GROUPS API ============
	t.Run("CustomerGroups", func(t *testing.T) {
		t.Run("ListCustomerGroups", func(t *testing.T) {
			resp, err := client.ListCustomerGroups(ctx, &CustomerGroupsListOptions{PageSize: 5})
			if err != nil {
				results = append(results, result{"ListCustomerGroups", false, 0, err})
				t.Errorf("ListCustomerGroups failed: %v", err)
				return
			}
			results = append(results, result{"ListCustomerGroups", true, len(resp.Items), nil})
			t.Logf("✓ ListCustomerGroups: %d groups returned", len(resp.Items))
		})

		t.Run("SearchCustomerGroups", func(t *testing.T) {
			resp, err := client.SearchCustomerGroups(ctx, &CustomerGroupSearchOptions{PageSize: 5})
			if err != nil {
				results = append(results, result{"SearchCustomerGroups", false, 0, err})
				t.Errorf("SearchCustomerGroups failed: %v", err)
				return
			}
			results = append(results, result{"SearchCustomerGroups", true, len(resp.Items), nil})
			t.Logf("✓ SearchCustomerGroups: %d groups returned", len(resp.Items))
		})
	})

	// ============ AGENTS API ============
	t.Run("Agents", func(t *testing.T) {
		t.Run("ListAgents", func(t *testing.T) {
			resp, err := client.ListAgents(ctx, &AgentsListOptions{PageSize: 5})
			if err != nil {
				results = append(results, result{"ListAgents", false, 0, err})
				t.Errorf("ListAgents failed: %v", err)
				return
			}
			results = append(results, result{"ListAgents", true, len(resp.Items), nil})
			t.Logf("✓ ListAgents: %d agents returned", len(resp.Items))
		})
	})

	// ============ CONVERSATIONS API ============
	t.Run("Conversations", func(t *testing.T) {
		t.Run("ListConversations", func(t *testing.T) {
			resp, err := client.ListConversations(ctx, &ConversationsListOptions{PageSize: 5})
			if err != nil {
				results = append(results, result{"ListConversations", false, 0, err})
				t.Errorf("ListConversations failed: %v", err)
				return
			}
			results = append(results, result{"ListConversations", true, len(resp.Items), nil})
			t.Logf("✓ ListConversations: %d conversations returned", len(resp.Items))
		})
	})

	// Print summary
	fmt.Println("\n========== PRODUCTION API TEST SUMMARY ==========")
	passed := 0
	failed := 0
	for _, r := range results {
		if r.success {
			passed++
			fmt.Printf("✓ %-25s: %d items\n", r.name, r.count)
		} else {
			failed++
			fmt.Printf("✗ %-25s: %v\n", r.name, r.err)
		}
	}
	fmt.Printf("\nTotal: %d passed, %d failed\n", passed, failed)
	fmt.Println("==================================================")
}
