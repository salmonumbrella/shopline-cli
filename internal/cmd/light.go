package cmd

import (
	"time"
	"unicode/utf8"

	"github.com/salmonumbrella/shopline-cli/internal/api"
)

// truncateRunes shortens s to at most max runes, appending "..." if truncated.
func truncateRunes(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	if max <= 3 {
		return "..."[:max]
	}
	runes := []rune(s)
	return string(runes[:max-3]) + "..."
}

// Light structs provide minimal payloads for agent-friendly output (--light / --li).
// Each keeps 3-7 essential fields, targeting <10 JSON lines per item.

type lightOrder struct {
	ID            string `json:"id"`
	OrderNumber   string `json:"order_number"`
	Status        string `json:"status"`
	PaymentStatus string `json:"payment_status"`
	FulfillStatus string `json:"fulfill_status"`
	TotalPrice    string `json:"total_price"`
	Currency      string `json:"currency"`
	CustomerName  string `json:"customer_name"`
}

func toLightOrder(o *api.Order) lightOrder {
	return lightOrder{
		ID:            o.ID,
		OrderNumber:   o.OrderNumber,
		Status:        o.Status,
		PaymentStatus: o.PaymentStatus,
		FulfillStatus: o.FulfillStatus,
		TotalPrice:    o.TotalPrice,
		Currency:      o.Currency,
		CustomerName:  truncateRunes(o.CustomerName, 50),
	}
}

type lightOrderSummary struct {
	ID            string `json:"id"`
	OrderNumber   string `json:"order_number"`
	Status        string `json:"status"`
	PaymentStatus string `json:"payment_status"`
	TotalPrice    string `json:"total_price"`
	Currency      string `json:"currency"`
	CustomerName  string `json:"customer_name"`
}

func toLightOrderSummary(o *api.OrderSummary) lightOrderSummary {
	return lightOrderSummary{
		ID:            o.ID,
		OrderNumber:   o.OrderNumber,
		Status:        o.Status,
		PaymentStatus: o.PaymentStatus,
		TotalPrice:    o.TotalPrice,
		Currency:      o.Currency,
		CustomerName:  truncateRunes(o.CustomerName, 50),
	}
}

type lightProduct struct {
	ID     string `json:"id"`
	Title  string `json:"title"`
	Status string `json:"status"`
	Vendor string `json:"vendor"`
	Handle string `json:"handle"`
}

func toLightProduct(p *api.Product) lightProduct {
	return lightProduct{
		ID:     p.ID,
		Title:  truncateRunes(p.Title, 80),
		Status: p.Status,
		Vendor: truncateRunes(p.Vendor, 50),
		Handle: p.Handle,
	}
}

type lightCustomer struct {
	ID          string `json:"id"`
	Email       string `json:"email"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Phone       string `json:"phone"`
	OrdersCount int    `json:"orders_count"`
	TotalSpent  string `json:"total_spent"`
}

func toLightCustomer(c *api.Customer) lightCustomer {
	return lightCustomer{
		ID:          c.ID,
		Email:       c.Email,
		FirstName:   c.FirstName,
		LastName:    c.LastName,
		Phone:       c.Phone,
		OrdersCount: c.OrdersCount,
		TotalSpent:  c.TotalSpent,
	}
}

type lightCoupon struct {
	ID            string    `json:"id"`
	Code          string    `json:"code"`
	Title         string    `json:"title"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	Status        string    `json:"status"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
}

func toLightCoupon(c *api.Coupon) lightCoupon {
	return lightCoupon{
		ID:            c.ID,
		Code:          truncateRunes(c.Code, 40),
		Title:         truncateRunes(c.Title, 80),
		DiscountType:  c.DiscountType,
		DiscountValue: c.DiscountValue,
		Status:        c.Status,
		StartsAt:      c.StartsAt,
		EndsAt:        c.EndsAt,
	}
}

type lightPromotion struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	Type          string    `json:"type"`
	Status        string    `json:"status"`
	DiscountType  string    `json:"discount_type"`
	DiscountValue float64   `json:"discount_value"`
	StartsAt      time.Time `json:"starts_at"`
	EndsAt        time.Time `json:"ends_at"`
}

func toLightPromotion(p *api.Promotion) lightPromotion {
	return lightPromotion{
		ID:            p.ID,
		Title:         truncateRunes(p.Title, 80),
		Type:          p.Type,
		Status:        p.Status,
		DiscountType:  p.DiscountType,
		DiscountValue: p.DiscountValue,
		StartsAt:      p.StartsAt,
		EndsAt:        p.EndsAt,
	}
}

type lightDraftOrder struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Status        string `json:"status"`
	CustomerEmail string `json:"customer_email"`
	TotalPrice    string `json:"total_price"`
	Currency      string `json:"currency"`
}

func toLightDraftOrder(d *api.DraftOrder) lightDraftOrder {
	return lightDraftOrder{
		ID:            d.ID,
		Name:          truncateRunes(d.Name, 80),
		Status:        d.Status,
		CustomerEmail: d.CustomerEmail,
		TotalPrice:    d.TotalPrice,
		Currency:      d.Currency,
	}
}

type lightAbandonedCheckout struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	TotalPrice string `json:"total_price"`
	Currency   string `json:"currency"`
	CustomerID string `json:"customer_id"`
}

func toLightAbandonedCheckout(a *api.AbandonedCheckout) lightAbandonedCheckout {
	return lightAbandonedCheckout{
		ID:         a.ID,
		Email:      a.Email,
		Phone:      a.Phone,
		TotalPrice: a.TotalPrice,
		Currency:   a.Currency,
		CustomerID: a.CustomerID,
	}
}

type lightReturnOrder struct {
	ID          string `json:"id"`
	OrderID     string `json:"order_id"`
	OrderNumber string `json:"order_number"`
	Status      string `json:"status"`
	ReturnType  string `json:"return_type"`
	TotalAmount string `json:"total_amount"`
	Currency    string `json:"currency"`
}

func toLightReturnOrder(r *api.ReturnOrder) lightReturnOrder {
	return lightReturnOrder{
		ID:          r.ID,
		OrderID:     r.OrderID,
		OrderNumber: r.OrderNumber,
		Status:      r.Status,
		ReturnType:  r.ReturnType,
		TotalAmount: r.TotalAmount,
		Currency:    r.Currency,
	}
}

type lightGiftCard struct {
	ID       string             `json:"id"`
	Code     string             `json:"code"`
	Balance  string             `json:"balance"`
	Currency string             `json:"currency"`
	Status   api.GiftCardStatus `json:"status"`
}

func toLightGiftCard(g *api.GiftCard) lightGiftCard {
	return lightGiftCard{
		ID:       g.ID,
		Code:     g.Code,
		Balance:  g.Balance,
		Currency: g.Currency,
		Status:   g.Status,
	}
}

// toLightSlice converts a slice of items using the given function.
func toLightSlice[T any, L any](items []T, fn func(*T) L) []L {
	result := make([]L, len(items))
	for i := range items {
		result[i] = fn(&items[i])
	}
	return result
}
