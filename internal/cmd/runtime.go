package cmd

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"sort"
	"strings"

	"github.com/salmonumbrella/shopline-cli/internal/api"
	"github.com/salmonumbrella/shopline-cli/internal/outfmt"
	"github.com/salmonumbrella/shopline-cli/internal/secrets"
	"github.com/spf13/cobra"
)

// CredentialStore defines the interface for credential storage operations.
type CredentialStore interface {
	List() ([]string, error)
	Get(name string) (*secrets.StoreCredentials, error)
}

// StoreFactory is a function that creates a credential store.
type StoreFactory func() (CredentialStore, error)

// ClientFactory is a function that creates an API client.
type ClientFactory func(handle, accessToken string) api.APIClient

const (
	envAccessToken      = "SHOPLINE_ACCESS_TOKEN"
	envAPIToken         = "SHOPLINE_API_TOKEN"
	envGenericToken     = "SHOPLINE_TOKEN"
	envAdminToken       = "SHOPLINE_ADMIN_TOKEN"
	defaultStoreEnvName = "SHOPLINE_STORE"
)

// clientFactory allows overriding the client creation for testing.
var clientFactory ClientFactory = defaultClientFactory

// secretsStoreFactory allows overriding the secrets store creation for testing.
var secretsStoreFactory StoreFactory = defaultSecretsStoreFactory

// formatterWriter is the output writer for formatters (can be overridden in tests).
var formatterWriter io.Writer = os.Stdout

// formatterWriterDefault holds the initial value so outWriter can detect overrides.
var formatterWriterDefault = formatterWriter

// stderrWriter is the writer for stderr messages (testable).
var stderrWriter io.Writer = os.Stderr

func defaultClientFactory(_, accessToken string) api.APIClient {
	return api.NewClient(accessToken)
}

func defaultSecretsStoreFactory() (CredentialStore, error) {
	return secrets.NewStore()
}

func getClient(cmd *cobra.Command) (api.APIClient, error) {
	storeName, _ := cmd.Flags().GetString("store")
	if storeName == "" {
		storeName = os.Getenv(defaultStoreEnvName)
	}
	storeName = resolveStoreAlias(storeName)

	// Allow token-only usage in environments where keyring is unavailable.
	if storeName == "" {
		if token := directAccessTokenFromEnv(); token != "" {
			return clientFactory("", token), nil
		}
	}

	store, err := secretsStoreFactory()
	if err != nil {
		return nil, fmt.Errorf("failed to open credential store: %w", err)
	}

	if storeName == "" {
		names, err := store.List()
		if err != nil {
			return nil, err
		}
		if len(names) == 0 {
			return nil, fmt.Errorf("no store profiles configured, run 'spl auth login'")
		}
		if len(names) == 1 {
			storeName = names[0]
		} else {
			return nil, fmt.Errorf("multiple profiles configured, use --store to select one")
		}
	}

	creds, err := resolveStoreCredentials(store, storeName)
	if err != nil {
		return nil, err
	}

	return clientFactory(creds.Handle, creds.AccessToken), nil
}

func directAccessTokenFromEnv() string {
	for _, name := range []string{envAccessToken, envAPIToken, envGenericToken} {
		if token := strings.TrimSpace(os.Getenv(name)); token != "" {
			return token
		}
	}
	return ""
}

type matchedProfile struct {
	name  string
	creds *secrets.StoreCredentials
}

func resolveStoreCredentials(store CredentialStore, storeName string) (*secrets.StoreCredentials, error) {
	name := strings.TrimSpace(storeName)
	if name == "" {
		return nil, profileNotFoundError(storeName, nil)
	}

	if creds, err := store.Get(name); err == nil {
		return creds, nil
	}

	matches, err := findProfileMatches(store, name)
	if err != nil {
		return nil, err
	}
	if len(matches) == 1 {
		_, _ = fmt.Fprintf(stderrWriter, "Using profile %q (matched from %q)\n", matches[0].name, storeName)
		return matches[0].creds, nil
	}
	if len(matches) > 1 {
		return nil, profileNotFoundError(storeName, matches)
	}

	return nil, profileNotFoundError(storeName, nil)
}

func profileNotFoundError(storeName string, matches []matchedProfile) error {
	base := fmt.Sprintf("profile not found: %s", storeName)
	if len(matches) > 1 {
		return fmt.Errorf(
			"%s (multiple matches: %s); use --store with an exact profile name and run 'spl auth ls' to list profiles or 'spl auth login' to add one",
			base,
			formatProfileNames(matches),
		)
	}
	return fmt.Errorf("%s; run 'spl auth ls' to list profiles or 'spl auth login' to add one", base)
}

func findProfileMatches(store CredentialStore, requested string) ([]matchedProfile, error) {
	names, err := store.List()
	if err != nil {
		return nil, err
	}
	if len(names) == 0 {
		return nil, nil
	}

	requestKeys := lookupKeys(requested)
	var (
		exactMatches  []matchedProfile
		prefixMatches []matchedProfile
	)

	for _, name := range names {
		creds, err := store.Get(name)
		if err != nil {
			continue
		}

		candidateKeys := lookupKeys(name)
		for k := range lookupKeys(creds.Handle) {
			candidateKeys[k] = struct{}{}
		}

		if hasAnyLookupKey(requestKeys, candidateKeys) {
			exactMatches = append(exactMatches, matchedProfile{name: name, creds: creds})
			continue
		}

		if hasPrefixLookupMatch(requestKeys, name, creds.Handle) {
			prefixMatches = append(prefixMatches, matchedProfile{name: name, creds: creds})
		}
	}

	if len(exactMatches) > 0 {
		return uniqueMatches(exactMatches), nil
	}
	return uniqueMatches(prefixMatches), nil
}

func hasAnyLookupKey(a, b map[string]struct{}) bool {
	for key := range a {
		if _, ok := b[key]; ok {
			return true
		}
	}
	return false
}

func hasPrefixLookupMatch(requestKeys map[string]struct{}, profileName, handle string) bool {
	profileName = strings.ToLower(strings.TrimSpace(profileName))
	handle = strings.ToLower(strings.TrimSpace(handle))

	for req := range requestKeys {
		if len(req) < 3 {
			continue
		}
		if strings.HasPrefix(profileName, req) || strings.HasPrefix(handle, req) {
			return true
		}
	}
	return false
}

func uniqueMatches(matches []matchedProfile) []matchedProfile {
	seen := make(map[string]struct{}, len(matches))
	out := make([]matchedProfile, 0, len(matches))
	for _, match := range matches {
		key := strings.TrimSpace(match.name)
		if key == "" {
			continue
		}
		if _, ok := seen[key]; ok {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, match)
	}
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].name) < strings.ToLower(out[j].name)
	})
	return out
}

func formatProfileNames(matches []matchedProfile) string {
	names := make([]string, 0, len(matches))
	for _, match := range matches {
		names = append(names, match.name)
	}
	if len(names) > 5 {
		names = names[:5]
	}
	return strings.Join(names, ", ")
}

func lookupKeys(value string) map[string]struct{} {
	out := make(map[string]struct{})
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return out
	}

	addLookupKey(out, normalized)
	addLookupVariants(out, normalized)
	if strings.Contains(normalized, ".") && !strings.Contains(normalized, "/") && !strings.Contains(normalized, "://") {
		addHostLookupKeys(out, normalized)
	}

	if !strings.Contains(normalized, "://") && strings.Contains(normalized, "/") {
		addLookupVariants(out, "https://"+normalized)
	}

	return out
}

func addLookupVariants(out map[string]struct{}, raw string) {
	parsed, err := url.Parse(raw)
	if err != nil || parsed == nil {
		return
	}

	if host := strings.ToLower(strings.TrimSpace(parsed.Hostname())); host != "" {
		addLookupKey(out, host)
		addHostLookupKeys(out, host)
	}

	path := strings.Trim(parsed.Path, "/")
	if path == "" {
		return
	}

	parts := strings.Split(path, "/")
	for i, part := range parts {
		addLookupKey(out, part)
		if part == "admin" && i+1 < len(parts) {
			addLookupKey(out, parts[i+1])
		}
	}
	addLookupKey(out, parts[len(parts)-1])
}

func addHostLookupKeys(out map[string]struct{}, host string) {
	for _, suffix := range []string{
		".myshopline.com",
		".shoplineapp.com",
		".shoplineapp.cn",
	} {
		if strings.HasSuffix(host, suffix) {
			addLookupKey(out, strings.TrimSuffix(host, suffix))
		}
	}

	if idx := strings.Index(host, "."); idx > 0 {
		label := host[:idx]
		if label != "www" && label != "admin" {
			addLookupKey(out, label)
		}
	}
}

func addLookupKey(out map[string]struct{}, raw string) {
	key := strings.ToLower(strings.TrimSpace(raw))
	key = strings.Trim(key, "/")
	if key == "" {
		return
	}
	out[key] = struct{}{}
}

// resolveStoreAlias expands short store aliases from SHOPLINE_STORE_ALIASES.
// Format: "alias1:fullname1,alias2:fullname2" (e.g. "ds:demoshop,ts:testshop")
func resolveStoreAlias(name string) string {
	if name == "" {
		return name
	}
	raw := os.Getenv("SHOPLINE_STORE_ALIASES")
	if raw == "" {
		return name
	}
	for _, pair := range strings.Split(raw, ",") {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), name) {
			return strings.TrimSpace(parts[1])
		}
	}
	return name
}

func getFormatter(cmd *cobra.Command) *outfmt.Formatter {
	outputFormat, _ := cmd.Flags().GetString("output")
	colorMode, _ := cmd.Flags().GetString("color")
	query, _ := cmd.Flags().GetString("output-query")
	if query == "" {
		// Fallback for unit tests that construct minimal commands without the hidden output-query flag.
		if f := cmd.Flags().Lookup("output-query"); f == nil {
			query, _ = cmd.Flags().GetString("query")
		}
	}
	itemsOnly, _ := cmd.Flags().GetBool("items-only")
	requestedOutputMode := getRequestedOutputMode(cmd, outputFormat)

	format := outfmt.FormatText
	if outputFormat == "json" {
		format = outfmt.FormatJSON
	}

	w := outWriter(cmd)

	f := outfmt.New(w, format, colorMode)
	if format == outfmt.FormatJSON {
		f = f.WithJSONMode(requestedOutputMode)
	}
	if prefix := idPrefixForCommand(cmd); prefix != "" {
		f = f.WithIDPrefix(prefix)
	}
	if query != "" {
		f = f.WithQuery(query)
	}
	if itemsOnly {
		f = f.WithItemsOnly(true)
	}
	return f
}

func getRequestedOutputMode(cmd *cobra.Command, outputFormat string) string {
	if cmd != nil && cmd.Flags().Lookup(outputModeFlagName) != nil {
		mode, _ := cmd.Flags().GetString(outputModeFlagName)
		normalizedMode := normalizeRequestedOutputValue(mode)
		switch normalizedMode {
		case "json", "jsonl", "ndjson", "text":
			return normalizedMode
		}
	}

	normalizedOutput := normalizeRequestedOutputValue(outputFormat)
	switch normalizedOutput {
	case "json", "jsonl", "ndjson", "text":
		return normalizedOutput
	}

	if normalizeOutputValue(outputFormat) == "json" {
		return "json"
	}
	return "text"
}
