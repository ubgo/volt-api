package volt

import (
	"slices"
	"strings"
)

// knownDevDefaultKeys are the placeholder encryption keys shipped for local
// development — the values committed in the PKL config templates. They are
// public knowledge and MUST never encrypt real data in a stag/prod
// deployment. Keep this list in sync with
// modules/config/pkl/env/{default,sample}/app.pkl `encryptionKey`.
var knownDevDefaultKeys = []string{
	"change-me-encryption-key-32ch", // default/app.pkl placeholder
	"xxxx",                          // sample/app.pkl placeholder
}

// IsWeakOrDefaultKey reports whether key is unusable for protecting real
// secrets: empty/blank, or one of the well-known dev-default placeholders.
// Boot wiring calls this in non-dev environments to fail closed rather than
// silently encrypting production data with a guessable key.
func IsWeakOrDefaultKey(key string) bool {
	if strings.TrimSpace(key) == "" {
		return true
	}
	return slices.Contains(knownDevDefaultKeys, key)
}
