package security

// NewAuthManager creates a new authentication manager
func NewAuthManager(configDir string) *AuthManager {
	return &AuthManager{
		configDir: configDir,
		users:     make(map[string]*User),
		roles:     make(map[string]*Role),
		sessions:  make(map[string]*Session),
	}
}
