package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// commonFlagAliases maps long flag names to short hidden aliases.
// These are applied to every command that has the flag.
// Aliases are hidden from help text but work identically.
var commonFlagAliases = map[string]string{
	// Pagination
	"page":      "pg",
	"page-size": "ps",

	// Filtering
	"status":      "S",
	"name":        "n",
	"email":       "e",
	"phone":       "ph",
	"description": "dsc",
	"title":       "ti",
	"tags":        "tg",
	"handle":      "h",
	"type":        "tp",
	"active":      "ac",
	"published":   "pub",

	// IDs
	"customer-id":       "cid",
	"product-id":        "pid",
	"order-id":          "oid",
	"variant-id":        "vid",
	"location-id":       "lid",
	"theme-id":          "tid",
	"parent-id":         "parid",
	"company-id":        "compid",
	"catalog-id":        "catid",
	"blog-id":           "bid",
	"user-id":           "uid",
	"channel-id":        "chid",
	"coupon-id":         "cpnid",
	"price-rule-id":     "prid",
	"inventory-item-id": "iid",

	// Time
	"from":       "f",
	"to":         "t",
	"starts-at":  "sa",
	"ends-at":    "ea",
	"expires-at": "expa",
	"start-date": "sd",
	"end-date":   "ed",

	// Content
	"body":      "b",
	"body-file": "bf",
	"note":      "N",
	"message":   "m",
	"content":   "ct",
	"source":    "src",

	// Amounts & Values
	"amount":         "amt",
	"price":          "pr",
	"quantity":       "qty",
	"value":          "v",
	"initial-value":  "iv",
	"delta":          "dl",
	"available":      "avl",
	"points":         "pts",
	"discount-type":  "dt",
	"discount-value": "dv",
	"currency":       "cur",
	"position":       "pos",

	// Customer
	"customer-email": "ce",
	"customer-name":  "cn",
	"warehouse-id":   "wid",

	// Content modifiers
	"remarks": "rmks",
	"remark":  "rmk",

	// Misc
	"reason":           "rsn",
	"tracking-number":  "tn",
	"tracking-company": "tc",
	"tracking-url":     "tu",
	"callback-url":     "cb",
	"namespace":        "ns",
	"owner-type":       "ot",
	"owner-id":         "owid",
	"product-type":     "pt",
	"code":             "cd",
	"sku":              "sk",
	"vendor":           "vn",
	"address":          "addr",
	"country":          "co",
	"city":             "ci",
	"platform":         "plat",
	"provider":         "prov",
	"restock":          "rs",
	"by":               "B",
	"expand":           "x",
	"graphql":          "gql",
	"first-name":       "fn",
	"last-name":        "ln",

	// IDs (additional)
	"assignee-id":        "aid",
	"delivery-option-id": "doi",
	"fulfillment-id":     "fid",
	"gift-product-id":    "gpi",
	"gift-variant-id":    "gvi",
	"order-ids":          "ois",
	// NOTE: "on" is also a subcommand alias for "activate" (aliases.go).
	// No collision: subcommand aliases resolve before flag parsing.
	"order-number":    "on",
	"page-id":         "pgi",
	"per-page":        "pp",
	"performer-id":    "pfi",
	"product-ids":     "pis",
	"promotion-id":    "pmi",
	"reference-id":    "rfi",
	"resource-id":     "rid",
	"resource-type":   "rt",
	"selling-plan-id": "spi",
	"store-id":        "sid",
	"supplier-id":     "sui",

	// Filtering (additional)
	"action":     "act",
	"archived":   "arc",
	"campaign":   "cmp",
	"category":   "cgy",
	"channel":    "chn",
	"collection": "col",
	"deep":       "dp",
	"default":    "def",
	"discount":   "dst",
	"display":    "dsp",
	"enabled":    "enb",
	"event":      "evt",
	// NOTE: "en" could collide with a future --en (English locale) flag.
	"event-name":   "en",
	"event-type":   "etp",
	"format":       "fmt",
	"gateway":      "gw",
	"image":        "img",
	"kind":         "kd",
	"level":        "lv",
	"locale":       "loc",
	"medium":       "med",
	"owner":        "own",
	"primary":      "pri",
	"private":      "pvt",
	"public":       "pbl",
	"role":         "rl",
	"rules":        "rul",
	"scopes":       "scp",
	"search-query": "sq",
	"search-type":  "sty",
	"searchable":   "sbl",
	"segment":      "seg",
	"sort-order":   "so",
	"source-type":  "srt",
	"state":        "stt",
	"tag":          "ta",
	"target":       "tgt",
	"text":         "txt",
	"topic":        "top",
	"value-type":   "vtp",
	"visible":      "vis",

	// Time (additional)
	"after":       "af",
	"before":      "bfr",
	"create-time": "ctm",
	"end-time":    "etm",
	"since":       "snc",
	"start-time":  "stm",
	"until":       "utl",

	// Content (additional)
	"author":       "atr",
	"body-html":    "bh",
	"filename":     "fil",
	"instructions": "ins",
	"notes":        "nts",
	"subject":      "sbj",

	// Amounts & Limits (additional)
	"budget":         "bgt",
	"credit-limit":   "crl",
	"exchange-rate":  "xr",
	"limit-per-user": "lpu",
	"max-quantity":   "mxq",
	"min-purchase":   "mnp",
	"min-quantity":   "mnq",
	"rate":           "ra",
	"score":          "scr",
	"trial-days":     "td",
	"unit":           "un",
	"usage-limit":    "ul",

	// Misc (additional)
	"batch":          "bat",
	"by-code":        "bc",
	"by-handle":      "bhd",
	"compound":       "cpd",
	"disjunctive":    "dj",
	"host":           "ho",
	"interval":       "itv",
	"options":        "opt",
	"path":           "pa",
	"priority":       "pio",
	"province":       "pv",
	"rating":         "rat",
	"recommendation": "rec",
	"required":       "req",
	"route":          "rte",
	"sandbox":        "sbx",
	"upsert":         "ups",
	"zip-code":       "zc",
}

// applyCommonFlagAliases walks the entire command tree and adds hidden
// short aliases for well-known flag names. Uses flagAlias from helpers.go.
func applyCommonFlagAliases(root *cobra.Command) {
	applyFlagAliasesRecursive(root)
}

func applyFlagAliasesRecursive(cmd *cobra.Command) {
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if alias, ok := commonFlagAliases[f.Name]; ok {
			flagAlias(cmd.Flags(), f.Name, alias)
		}
	})
	// Also alias persistent flags so subcommands inherit short forms
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if alias, ok := commonFlagAliases[f.Name]; ok {
			flagAlias(cmd.PersistentFlags(), f.Name, alias)
		}
	})
	for _, sub := range cmd.Commands() {
		applyFlagAliasesRecursive(sub)
	}
}

// applyRootFlagAliases adds short aliases for root persistent flags.
func applyRootFlagAliases(root *cobra.Command) {
	pf := root.PersistentFlags()
	flagAlias(pf, "json", "j")
	flagAlias(pf, "output", "out")
	flagAlias(pf, "query", "qr")
	flagAlias(pf, "query-file", "qf")
	flagAlias(pf, "dry-run", "dr")
	flagAlias(pf, "sort-by", "sb")
	flagAlias(pf, "items-only", "io")
	flagAlias(pf, "results-only", "ro")
	flagAlias(pf, "admin-token", "at")
	flagAlias(pf, "admin-merchant-id", "amid")
}
