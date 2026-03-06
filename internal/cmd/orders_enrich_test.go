package cmd

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/salmonumbrella/shopline-cli/internal/api"
)

func TestParseRawPriceRejectsDisplayOnlyString(t *testing.T) {
	raw := json.RawMessage(`"NT$1,999"`)
	price, ok := parseRawPrice(raw, "TWD")
	if ok {
		t.Fatalf("expected display-only string price to be rejected, got %+v", price)
	}
}

func TestParseRawPriceAcceptsPlainNumericString(t *testing.T) {
	raw := json.RawMessage(`"19.99"`)
	price, ok := parseRawPrice(raw, "USD")
	if !ok {
		t.Fatal("expected plain numeric string price to parse")
	}
	if price.Cents != 1999 {
		t.Fatalf("expected 1999 minor units, got %d", price.Cents)
	}
}

func TestParseRawPriceUsesCurrencyScaleForRawNumbers(t *testing.T) {
	raw := json.RawMessage(`1999`)
	price, ok := parseRawPrice(raw, "TWD")
	if !ok {
		t.Fatal("expected numeric price to parse")
	}
	if price.Cents != 1999 {
		t.Fatalf("expected 1999 minor units for TWD, got %d", price.Cents)
	}
}

func TestParseRawPriceRoundsRawNumbers(t *testing.T) {
	raw := json.RawMessage(`19.995`)
	price, ok := parseRawPrice(raw, "USD")
	if !ok {
		t.Fatal("expected numeric price to parse")
	}
	if price.Cents != 2000 {
		t.Fatalf("expected rounded 2000 minor units, got %d", price.Cents)
	}
}

func TestParseRawPriceDerivesMinorUnitsFromObjectDollars(t *testing.T) {
	raw := json.RawMessage(`{"currency_iso":"USD","dollars":19.99}`)
	price, ok := parseRawPrice(raw, "USD")
	if !ok {
		t.Fatal("expected price object with dollars to parse")
	}
	if price.Cents != 1999 {
		t.Fatalf("expected 1999 minor units, got %d", price.Cents)
	}
}

func TestDeriveOrderTotalPrefersSubtotalItemTotals(t *testing.T) {
	order := &api.Order{
		SubtotalItems: []api.OrderSubtotalItem{
			{
				Quantity:   1,
				TotalPrice: &api.Price{Cents: 12345, CurrencyISO: "USD"},
				ItemPrice:  &api.Price{Cents: 9999, CurrencyISO: "USD"},
			},
		},
	}

	amount, currency := deriveOrderTotal(order)
	if amount != "123.45" {
		t.Fatalf("expected amount 123.45, got %q", amount)
	}
	if currency != "USD" {
		t.Fatalf("expected currency USD, got %q", currency)
	}
}

func TestDeriveOrderTotalUsesZeroDecimalCurrencyForLineItemNumbers(t *testing.T) {
	order := &api.Order{
		Currency: "TWD",
		LineItems: []api.OrderLineItem{
			{
				Quantity: 1,
				Currency: "TWD",
				Price:    json.RawMessage(`1999`),
			},
		},
	}

	amount, currency := deriveOrderTotal(order)
	if amount != "1999" {
		t.Fatalf("expected amount 1999, got %q", amount)
	}
	if currency != "TWD" {
		t.Fatalf("expected currency TWD, got %q", currency)
	}
}

func TestEnrichOrderSummaryBackfillsDeliveryStatusWhenFulfillMissing(t *testing.T) {
	client := &mockAPIClient{
		getOrderActionLogsResp: json.RawMessage(`{"items":[
			{"key":"updated_delivery_status","data":{"updated_delivery_status":"collected"},"created_at":"2026-03-01T12:03:00Z"},
			{"key":"updated_payment_status","data":{"updated_payment_status":"completed"},"created_at":"2026-03-01T12:04:00Z"}
		]}`),
	}

	out := enrichOrderSummary(context.Background(), client, api.OrderSummary{
		ID:          "ord_1",
		OrderNumber: "1001",
		Status:      "completed",
	})

	if out.FulfillStatus != "collected" {
		t.Fatalf("expected fulfill_status=collected from action logs, got %q", out.FulfillStatus)
	}
	if out.DeliveryStatus != "collected" {
		t.Fatalf("expected delivery_status=collected from action logs, got %q", out.DeliveryStatus)
	}
	if out.PaymentStatus != "completed" {
		t.Fatalf("expected payment_status=completed from action logs, got %q", out.PaymentStatus)
	}
}
