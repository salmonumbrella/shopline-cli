package cmd

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"
)

var verbAliases = map[string][]string{
	"list":          {"ls", "l"},
	"get":           {"show", "g"},
	"show":          {"g"},
	"create":        {"new", "add", "mk"},
	"update":        {"edit", "up"},
	"edit":          {"up"},
	"delete":        {"del", "rm"},
	"remove":        {"rm"},
	"cancel":        {"void"},
	"search":        {"find", "q"},
	"query":         {"q"},
	"find":          {"q"},
	"count":         {"cnt"},
	"login":         {"signin", "sign-in"},
	"logout":        {"signout", "sign-out"},
	"close":         {"cl"},
	"reopen":        {"re"},
	"complete":      {"cpl"},
	"publish":       {"pub"},
	"adjust":        {"adj"},
	"send-recovery": {"sr"},
	"info":          {"i"},
	"settings":      {"cfg"},
	"status":        {"stat"},

	// Subcommand nouns
	"action-logs":   {"al"},
	"activate":      {"on"},
	"assign":        {"asg"},
	"capture":       {"cap"},
	"children":      {"chi"},
	"comments":      {"cmt"},
	"deactivate":    {"off"},
	"domains":       {"dom"},
	"enable":        {"ena"},
	"events":        {"ev"},
	"exchange":      {"xch"},
	"execute":       {"ex"},
	"end":           {"fin"},
	"hide":          {"hid"},
	"history":       {"his"},
	"images":        {"img"},
	"items":         {"itm"},
	"metafields":    {"mf"},
	"order":         {"ord"},
	"payments":      {"pay"},
	"profiles":      {"prf"},
	"receive":       {"rcv"},
	"segments":      {"seg"},
	"send":          {"snd"},
	"send-invoice":  {"si"},
	"start":         {"go"},
	"stocks":        {"stk"},
	"summary":       {"sum"},
	"tracking":      {"trk"},
	"unpublish":     {"unp"},
	"variations":    {"var"},
}

var resourceAliases = map[string][]string{
	// High frequency
	"orders":       {"ord", "o"},
	"products":     {"prod", "p"},
	"customers":    {"cust", "cu", "contacts", "contact"},
	"refunds":      {"ref", "rf"},
	"collections":  {"col"},
	"draft-orders": {"drafts", "do"},
	"fulfillments": {"ful", "ff"},
	"payments":     {"pay"},
	"transactions": {"tx"},
	"shipments":    {"shp"},
	"inventory":    {"inv"},

	// Medium frequency
	"abandoned-checkouts":           {"ac"},
	"addon-products":                {"ap"},
	"affiliate-campaigns":           {"afc"},
	"auth":                          {"au"},
	"articles":                      {"art"},
	"assets":                        {"as"},
	"balance":                       {"bal"},
	"blogs":                         {"bl"},
	"bulk-operations":               {"bo"},
	"carrier-services":              {"cs"},
	"carts":                         {"ct"},
	"catalog-pricing":               {"cp"},
	"categories":                    {"cat"},
	"channels":                      {"ch"},
	"channel-products":              {"chp"},
	"checkout-settings":             {"cos"},
	"company-catalogs":              {"cc"},
	"company-credits":               {"ccr"},
	"conversations":                 {"conv", "cv"},
	"countries":                     {"cnt"},
	"coupons":                       {"cpn"},
	"currencies":                    {"cur"},
	"custom-fields":                 {"cf"},
	"customer-addresses":            {"ca"},
	"customer-blacklist":            {"cbl"},
	"customer-groups":               {"cg"},
	"customer-saved-searches":       {"css"},
	"delivery-options":              {"dop"},
	"discount-codes":                {"discounts", "dc"},
	"disputes":                      {"dis"},
	"domains":                       {"dm"},
	"express-links":                 {"el"},
	"fields":                        {"fld"},
	"fields-presets":                {"fp"},
	"files":                         {"fi"},
	"flash-price":                   {"flp"},
	"flash-price-campaigns":         {"fpc"},
	"fulfillment-orders":            {"fo"},
	"fulfillment-services":          {"fs"},
	"gift-cards":                    {"giftcard", "gc"},
	"gifts":                         {"gi"},
	"inventory-levels":              {"il"},
	"labels":                        {"lb"},
	"livestreams":                   {"lv"},
	"local-delivery":                {"ld"},
	"locations":                     {"loc"},
	"marketing-events":              {"me"},
	"markets":                       {"mk"},
	"media":                         {"md"},
	"medias":                        {"mds"},
	"member-points":                 {"mp"},
	"membership":                    {"mem"},
	"merchants":                     {"mr"},
	"message-center":                {"mc"},
	"metafield-definitions":         {"mfd"},
	"metafields":                    {"mf"},
	"multipass":                     {"mup"},
	"operation-logs":                {"ol"},
	"order-attribution":             {"oa"},
	"order-risks":                   {"ork"},
	"orders-metafields":             {"omf"},
	"pages":                         {"pg"},
	"payouts":                       {"po"},
	"pickup":                        {"pu"},
	"pos-purchase-orders":           {"ppo"},
	"price-rules":                   {"pr"},
	"product-listings":              {"pl"},
	"product-review-comments":       {"prc"},
	"product-reviews":               {"prv"},
	"product-subscriptions":         {"psu"},
	"products-metafields":           {"pmf"},
	"promotions":                    {"promo", "pm"},
	"purchase-orders":               {"pur"},
	"redirects":                     {"rd"},
	"return-orders":                 {"ro"},
	"sales":                         {"sl"},
	"schema":                        {"sch"},
	"script-tags":                   {"st"},
	"selling-plans":                 {"sp"},
	"settings":                      {"set"},
	"settings-endpoints":            {"se"},
	"shipping":                      {"ship", "sh"},
	"shipping-zones":                {"sz"},
	"shoplytics":                    {"sly"},
	"social-posts":                  {"sop"},
	"size-charts":                   {"sc"},
	"smart-collections":             {"smc"},
	"staffs":                        {"sf"},
	"staffs-permissions":            {"sfp"},
	"store-credits":                 {"scr"},
	"storefront-carts":              {"sfc"},
	"storefront-oauth":              {"sfo"},
	"storefront-oauth-applications": {"sfoa"},
	"storefront-products":           {"sfpr"},
	"storefront-promotions":         {"sfpm"},
	"storefront-tokens":             {"sft"},
	"subscriptions":                 {"sub"},
	"tags":                          {"tg"},
	"tax-services":                  {"ts"},
	"taxes":                         {"tax"},
	"taxonomies":                    {"txn"},
	"themes":                        {"th"},
	"token":                         {"tk"},
	"tokens":                        {"tok"},
	"user-coupons":                  {"uc"},
	"user-credits":                  {"ucr"},
	"warehouses":                    {"wh"},
	"webhooks":                      {"hooks", "hk"},
	"wish-list-items":               {"wli"},
	"wish-lists":                    {"wl"},
}

func applyDesirePathAliases(root *cobra.Command) {
	if root == nil {
		return
	}
	applyAliasesRecursive(root, root)
}

func applyAliasesRecursive(cmd *cobra.Command, root *cobra.Command) {
	addDesireAliases(cmd, root)
	for _, sub := range cmd.Commands() {
		applyAliasesRecursive(sub, root)
	}
}

func addDesireAliases(cmd *cobra.Command, root *cobra.Command) {
	name := cmd.Name()
	if name == "" {
		return
	}

	if aliases, ok := verbAliases[name]; ok {
		for _, a := range aliases {
			addAliasIfSafe(cmd, a)
		}
	}

	if cmd.Parent() == root {
		if singular := singularize(name); singular != "" && singular != name {
			addAliasIfSafe(cmd, singular)
		}
		if aliases, ok := resourceAliases[name]; ok {
			for _, a := range aliases {
				addAliasIfSafe(cmd, a)
			}
		}
	}
}

func addAliasIfSafe(cmd *cobra.Command, alias string) {
	if alias == "" || alias == cmd.Name() || slices.Contains(cmd.Aliases, alias) {
		return
	}
	parent := cmd.Parent()
	if parent != nil {
		for _, sibling := range parent.Commands() {
			if sibling == cmd {
				continue
			}
			if sibling.Name() == alias || slices.Contains(sibling.Aliases, alias) {
				return
			}
		}
	}
	cmd.Aliases = append(cmd.Aliases, alias)
}

func singularize(name string) string {
	if strings.HasSuffix(name, "ies") && len(name) > 3 {
		return name[:len(name)-3] + "y"
	}
	if strings.HasSuffix(name, "xes") || strings.HasSuffix(name, "ses") ||
		strings.HasSuffix(name, "ches") || strings.HasSuffix(name, "shes") {
		return name[:len(name)-2]
	}
	// Don't strip trailing "s" from words ending in "us", "ss", or "is"
	if strings.HasSuffix(name, "us") || strings.HasSuffix(name, "ss") || strings.HasSuffix(name, "is") {
		return name
	}
	if strings.HasSuffix(name, "s") && len(name) > 1 {
		return name[:len(name)-1]
	}
	return name
}
