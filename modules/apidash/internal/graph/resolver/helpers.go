package resolver

import "strings"

// buildFullName joins a first and last name into a single display name.
//
// Why it lives here (not in a *.resolvers.go file): gqlgen owns the
// *.resolvers.go files and, on `gqlgen generate`, keeps ONLY resolver methods
// — any free function placed there is dropped. Helpers the resolvers call must
// live in a non-resolvers file (this one) so codegen never removes them.
// Returns nil when both parts are empty so the GraphQL field resolves to null.
func buildFullName(firstName, lastName string) *string {
	var parts []string
	if firstName != "" {
		parts = append(parts, firstName)
	}
	if lastName != "" {
		parts = append(parts, lastName)
	}
	if len(parts) == 0 {
		return nil
	}
	fullName := strings.Join(parts, " ")
	return &fullName
}
