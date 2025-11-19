package testenv

// Env is a minimal interface for test environment configuration
type Env interface {
	// Tenant returns a tenant name
	Tenant() (string, error)
}
