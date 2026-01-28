package provider

// contextKey is a custom type for context keys to avoid collisions.
type contextKey string

// PreferredProviderKey is the context key for specifying a preferred LLM provider.
const PreferredProviderKey contextKey = "preferred_provider"
