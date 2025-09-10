package config

// UserClaimsContextKey is the key for the user claims in the request context.
type contextKey string

const UserClaimsContextKey contextKey = "userClaims"
const RoleContextKey = "roles"
