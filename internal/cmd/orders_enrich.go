package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/salmonumbrella/shopline-cli/internal/api"
)

type orderSummaryOutput struct {
	ID             string    `json:"id"`
	OrderNumber    string    `json:"order_number"`
	Status         string    `json:"status"`
	OrderStatus    string    `json:"order_status"`
	PaymentStatus  string    `json:"payment_status"`
	FulfillStatus  string    `json:"fulfill_status"`
	DeliveryStatus string    `json:"delivery_status"`
	TotalPrice     string    `json:"total_price"`
	TotalAmount    string    `json:"total_amount"`
	Currency       string    `json:"currency"`
	CustomerEmail  string    `json:"customer_email"`
	CustomerName   string    `json:"customer_name"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

func buildOrderSummaryOutputResponse(
	ctx context.Context,
	client api.APIClient,
	resp *api.OrdersListResponse,
	jobs int,
) api.ListResponse[orderSummaryOutput] {
	out := api.ListResponse[orderSummaryOutput]{}
	if resp == nil {
		out.Items = []orderSummaryOutput{}
		return out
	}

	out.Pagination = resp.Pagination
	out.Page = resp.Page
	out.PageSize = resp.PageSize
	out.TotalCount = resp.TotalCount
	out.HasMore = resp.HasMore
	if len(resp.Items) == 0 {
		out.Items = []orderSummaryOutput{}
		return out
	}

	if jobs < 1 {
		jobs = 1
	}

	out.Items = make([]orderSummaryOutput, len(resp.Items))
	sem := make(chan struct{}, jobs)
	var wg sync.WaitGroup
	for i := range resp.Items {
		i := i
		sem <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			out.Items[i] = enrichOrderSummary(ctx, client, resp.Items[i])
		}()
	}
	wg.Wait()
	return out
}

func enrichOrderSummary(ctx context.Context, client api.APIClient, summary api.OrderSummary) orderSummaryOutput {
	out := orderSummaryOutput{
		ID:             summary.ID,
		OrderNumber:    summary.OrderNumber,
		Status:         summary.Status,
		OrderStatus:    summary.Status,
		PaymentStatus:  summary.PaymentStatus,
		FulfillStatus:  summary.FulfillStatus,
		DeliveryStatus: summary.FulfillStatus,
		TotalPrice:     summary.TotalPrice,
		TotalAmount:    summary.TotalPrice,
		Currency:       summary.Currency,
		CustomerEmail:  summary.CustomerEmail,
		CustomerName:   summary.CustomerName,
		CreatedAt:      summary.CreatedAt,
		UpdatedAt:      summary.UpdatedAt,
	}

	if client == nil {
		normalizeOrderSummaryOutput(&out)
		return out
	}

	if needsOrderDetailEnrichment(out) {
		if detail, err := client.GetOrder(ctx, summary.ID); err == nil && detail != nil {
			applyOrderDetailEnrichment(&out, detail)
		}
	}

	if needsOrderActionLogEnrichment(out) {
		if raw, err := client.GetOrderActionLogs(ctx, summary.ID); err == nil && len(raw) > 0 {
			if logs, parseErr := parseOrderActionLogs(raw); parseErr == nil {
				applyOrderActionLogEnrichment(&out, logs)
			}
		}
	}

	normalizeOrderSummaryOutput(&out)
	return out
}

func needsOrderDetailEnrichment(out orderSummaryOutput) bool {
	return out.TotalAmount == "" || out.TotalPrice == "" || out.Currency == ""
}

func needsOrderActionLogEnrichment(out orderSummaryOutput) bool {
	return out.PaymentStatus == "" || out.FulfillStatus == "" || out.OrderStatus == ""
}

func applyOrderDetailEnrichment(out *orderSummaryOutput, detail *api.Order) {
	if out == nil || detail == nil {
		return
	}

	out.Status = firstNonEmpty(out.Status, detail.Status)
	out.OrderStatus = firstNonEmpty(out.OrderStatus, detail.Status)
	out.PaymentStatus = firstNonEmpty(out.PaymentStatus, detail.PaymentStatus)
	out.FulfillStatus = firstNonEmpty(out.FulfillStatus, detail.FulfillStatus)
	out.DeliveryStatus = firstNonEmpty(out.DeliveryStatus, detail.FulfillStatus)
	out.CustomerEmail = firstNonEmpty(out.CustomerEmail, detail.CustomerEmail)
	out.CustomerName = firstNonEmpty(out.CustomerName, detail.CustomerName)

	totalAmount, currency := deriveOrderTotal(detail)
	out.TotalAmount = firstNonEmpty(out.TotalAmount, detail.TotalPrice, totalAmount)
	out.TotalPrice = firstNonEmpty(out.TotalPrice, detail.TotalPrice, totalAmount)
	out.Currency = firstNonEmpty(out.Currency, detail.Currency, currency)
}

func applyOrderActionLogEnrichment(out *orderSummaryOutput, logs []orderActionLog) {
	if out == nil || len(logs) == 0 {
		return
	}

	out.OrderStatus = firstNonEmpty(
		out.OrderStatus,
		latestOrderActionLogValue(logs, "updated_status", "updated_order_status"),
	)
	out.Status = firstNonEmpty(out.Status, out.OrderStatus)

	delivery := latestOrderActionLogValue(logs, "updated_delivery_status", "updated_order_delivery_status")
	out.DeliveryStatus = firstNonEmpty(out.DeliveryStatus, delivery)
	out.FulfillStatus = firstNonEmpty(out.FulfillStatus, delivery)

	out.PaymentStatus = firstNonEmpty(
		out.PaymentStatus,
		latestOrderActionLogValue(logs, "updated_payment_status", "updated_order_payment_status"),
	)
}

func latestOrderActionLogValue(logs []orderActionLog, wantKeys ...string) string {
	if len(logs) == 0 || len(wantKeys) == 0 {
		return ""
	}

	targets := make(map[string]struct{}, len(wantKeys))
	for _, key := range wantKeys {
		key = strings.ToLower(strings.TrimSpace(key))
		if key != "" {
			targets[key] = struct{}{}
		}
	}
	if len(targets) == 0 {
		return ""
	}

	var (
		bestValue string
		bestTime  string
		fallback  string
	)

	for _, log := range logs {
		var value string
		for key := range targets {
			if value == "" {
				value = firstNonEmpty(
					stringFromMap(log.Data, key),
					stringFromMap(log.PerformData, key),
				)
			}
		}
		if value == "" {
			continue
		}
		fallback = value
		if bestTime == "" || isLaterActionLogTime(log.CreatedAt, bestTime) {
			bestTime = strings.TrimSpace(log.CreatedAt)
			bestValue = value
		}
	}

	return firstNonEmpty(bestValue, fallback)
}

func isLaterActionLogTime(left, right string) bool {
	leftTime, leftOK := parseActionLogTime(left)
	rightTime, rightOK := parseActionLogTime(right)
	switch {
	case leftOK && rightOK:
		return leftTime.After(rightTime)
	case leftOK:
		return true
	case rightOK:
		return false
	default:
		return strings.TrimSpace(left) > strings.TrimSpace(right)
	}
}

func deriveOrderTotal(detail *api.Order) (string, string) {
	if detail == nil {
		return "", ""
	}

	if amount := strings.TrimSpace(detail.TotalPrice); amount != "" {
		return amount, strings.TrimSpace(detail.Currency)
	}

	totalMinor := 0
	currency := strings.TrimSpace(detail.Currency)
	found := false

	for _, item := range detail.SubtotalItems {
		if item.TotalPrice != nil {
			totalMinor += item.TotalPrice.Cents
			currency = firstNonEmpty(currency, item.TotalPrice.CurrencyISO)
			found = true
			continue
		}

		price := firstNonNilPrice(item.ItemPrice, item.DiscountedPrice, item.PriceSale, item.Price)
		if price != nil {
			qty := item.Quantity
			if qty <= 0 {
				qty = 1
			}
			totalMinor += price.Cents * qty
			currency = firstNonEmpty(currency, price.CurrencyISO)
			found = true
		}
	}

	if !found {
		for _, item := range detail.LineItems {
			lineCurrency := firstNonEmpty(item.Currency, currency)
			price, ok := parseRawPrice(item.Total, lineCurrency)
			if !ok {
				price, ok = parseRawPrice(item.Subtotal, lineCurrency)
			}
			if !ok {
				price, ok = parseRawPrice(item.Price, lineCurrency)
				if ok {
					qty := item.Quantity
					if qty <= 0 {
						qty = 1
					}
					price.Cents *= qty
				}
			}
			if !ok {
				continue
			}
			totalMinor += price.Cents
			currency = firstNonEmpty(currency, item.Currency, price.CurrencyISO)
			found = true
		}
	}

	if !found {
		return "", currency
	}

	return formatMinorUnits(totalMinor, currency), currency
}

func firstNonNilPrice(prices ...*api.Price) *api.Price {
	for _, price := range prices {
		if price != nil {
			return price
		}
	}
	return nil
}

func parseRawPrice(raw json.RawMessage, currency string) (api.Price, bool) {
	if len(raw) == 0 {
		return api.Price{}, false
	}

	var price api.Price
	if err := json.Unmarshal(raw, &price); err == nil && (price.Cents != 0 || price.CurrencyISO != "" || price.Label != "" || price.Dollars != 0) {
		price.CurrencyISO = firstNonEmpty(price.CurrencyISO, strings.ToUpper(strings.TrimSpace(currency)))
		if price.Cents == 0 && price.Dollars != 0 {
			price.Cents = minorUnitsFromMajor(price.Dollars, price.CurrencyISO)
		}
		return price, true
	}

	var number float64
	if err := json.Unmarshal(raw, &number); err == nil {
		return api.Price{
			Cents:   minorUnitsFromMajor(number, currency),
			Dollars: number,
		}, true
	}

	var text string
	if err := json.Unmarshal(raw, &text); err == nil {
		text = strings.TrimSpace(text)
		if !looksLikePlainNumber(text) {
			return api.Price{}, false
		}
		number, err := strconv.ParseFloat(text, 64)
		if err != nil {
			return api.Price{}, false
		}
		return api.Price{
			Cents:   minorUnitsFromMajor(number, currency),
			Dollars: number,
		}, true
	}

	return api.Price{}, false
}

func looksLikePlainNumber(s string) bool {
	if s == "" {
		return false
	}
	start := 0
	if s[0] == '-' {
		if len(s) == 1 {
			return false
		}
		start = 1
	}

	seenDigit := false
	seenDot := false
	for i := start; i < len(s); i++ {
		switch ch := s[i]; {
		case ch >= '0' && ch <= '9':
			seenDigit = true
		case ch == '.' && !seenDot:
			seenDot = true
		default:
			return false
		}
	}
	return seenDigit
}

func minorUnitsFromMajor(value float64, currency string) int {
	decimals := 2
	cur := strings.ToUpper(strings.TrimSpace(currency))
	if _, ok := zeroDecimalCurrencies[cur]; ok {
		decimals = 0
	} else if _, ok := threeDecimalCurrencies[cur]; ok {
		decimals = 3
	}

	scale := math.Pow10(decimals)
	return int(math.Round(value * scale))
}

func formatMinorUnits(minor int, currency string) string {
	cur := strings.ToUpper(strings.TrimSpace(currency))
	decimals := 2
	if _, ok := zeroDecimalCurrencies[cur]; ok {
		decimals = 0
	} else if _, ok := threeDecimalCurrencies[cur]; ok {
		decimals = 3
	}

	sign := ""
	if minor < 0 {
		sign = "-"
		minor = -minor
	}

	if decimals == 0 {
		return fmt.Sprintf("%s%d", sign, minor)
	}

	divisor := 1
	for i := 0; i < decimals; i++ {
		divisor *= 10
	}
	whole := minor / divisor
	fraction := minor % divisor
	return fmt.Sprintf("%s%d.%0*d", sign, whole, decimals, fraction)
}

func normalizeOrderSummaryOutput(out *orderSummaryOutput) {
	if out == nil {
		return
	}
	out.OrderStatus = firstNonEmpty(out.OrderStatus, out.Status)
	out.Status = firstNonEmpty(out.Status, out.OrderStatus)

	out.DeliveryStatus = firstNonEmpty(out.DeliveryStatus, out.FulfillStatus)
	out.FulfillStatus = firstNonEmpty(out.FulfillStatus, out.DeliveryStatus)

	out.TotalAmount = firstNonEmpty(out.TotalAmount, out.TotalPrice)
	out.TotalPrice = firstNonEmpty(out.TotalPrice, out.TotalAmount)
}
