package versioning

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// APIVersion represents a semantic version
type APIVersion struct {
	Major int
	Minor int
	Patch int
}

// String returns the string representation of the version
func (v APIVersion) String() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// Compare compares two versions
// Returns: -1 if v < other, 0 if v == other, 1 if v > other
func (v APIVersion) Compare(other APIVersion) int {
	if v.Major != other.Major {
		if v.Major < other.Major {
			return -1
		}
		return 1
	}
	if v.Minor != other.Minor {
		if v.Minor < other.Minor {
			return -1
		}
		return 1
	}
	if v.Patch != other.Patch {
		if v.Patch < other.Patch {
			return -1
		}
		return 1
	}
	return 0
}

// SupportedVersions defines all supported API versions
var (
	Version_1_0_0  = APIVersion{Major: 1, Minor: 0, Patch: 0}
	Version_1_1_0  = APIVersion{Major: 1, Minor: 1, Patch: 0}
	CurrentVersion = Version_1_1_0
	DefaultVersion = Version_1_0_0
)

// SupportedVersionsList contains all supported versions in descending order
var SupportedVersionsList = []APIVersion{
	Version_1_1_0,
	Version_1_0_0,
}

// DeprecatedVersions contains versions that are deprecated
var DeprecatedVersions = map[APIVersion]time.Time{
	Version_1_0_0: time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC), // Deprecated end of 2025
}

// VersionContextKey is the key for version in context
type VersionContextKey struct{}

// VersionConfig holds versioning configuration
type VersionConfig struct {
	DefaultVersion     APIVersion
	CurrentVersion     APIVersion
	SupportedVersions  []APIVersion
	DeprecatedVersions map[APIVersion]time.Time
}

// NewVersionConfig creates a new version configuration
func NewVersionConfig() *VersionConfig {
	return &VersionConfig{
		DefaultVersion:     DefaultVersion,
		CurrentVersion:     CurrentVersion,
		SupportedVersions:  SupportedVersionsList,
		DeprecatedVersions: DeprecatedVersions,
	}
}

// ParseVersion parses a version string (e.g., "v1.1.0", "1.1.0")
func ParseVersion(versionStr string) (APIVersion, error) {
	versionStr = strings.TrimPrefix(versionStr, "v")

	var major, minor, patch int
	n, err := fmt.Sscanf(versionStr, "%d.%d.%d", &major, &minor, &patch)
	if err != nil || n != 3 {
		return APIVersion{}, fmt.Errorf("invalid version format: %s", versionStr)
	}

	return APIVersion{Major: major, Minor: minor, Patch: patch}, nil
}

// IsVersionSupported checks if a version is supported
func (vc *VersionConfig) IsVersionSupported(version APIVersion) bool {
	for _, supported := range vc.SupportedVersions {
		if supported.Compare(version) == 0 {
			return true
		}
	}
	return false
}

// IsVersionDeprecated checks if a version is deprecated
func (vc *VersionConfig) IsVersionDeprecated(version APIVersion) bool {
	_, exists := vc.DeprecatedVersions[version]
	return exists
}

// GetDeprecationDate returns the deprecation date for a version
func (vc *VersionConfig) GetDeprecationDate(version APIVersion) (time.Time, bool) {
	date, exists := vc.DeprecatedVersions[version]
	return date, exists
}

// NegotiateVersion determines the appropriate version based on client request
func (vc *VersionConfig) NegotiateVersion(r *http.Request) APIVersion {
	// Check Accept-Version header
	if versionStr := r.Header.Get("Accept-Version"); versionStr != "" {
		if version, err := ParseVersion(versionStr); err == nil {
			if vc.IsVersionSupported(version) {
				return version
			}
		}
	}

	// Check Accept header for version negotiation
	if accept := r.Header.Get("Accept"); accept != "" {
		if version := vc.parseAcceptHeader(accept); version != (APIVersion{}) {
			return version
		}
	}

	// Check URL path for version (e.g., /api/v1.1/files)
	if version := vc.parseURLPath(r.URL.Path); version != (APIVersion{}) {
		return version
	}

	// Default to current version
	return vc.CurrentVersion
}

// parseAcceptHeader parses version from Accept header
func (vc *VersionConfig) parseAcceptHeader(accept string) APIVersion {
	// Look for version parameter in Accept header
	if strings.Contains(accept, "version=") {
		parts := strings.Split(accept, "version=")
		if len(parts) > 1 {
			versionStr := strings.Split(parts[1], ";")[0]
			if version, err := ParseVersion(versionStr); err == nil {
				if vc.IsVersionSupported(version) {
					return version
				}
			}
		}
	}
	return APIVersion{}
}

// parseURLPath parses version from URL path
func (vc *VersionConfig) parseURLPath(path string) APIVersion {
	// Look for version in path (e.g., /api/v1.0/files or /api/v1.1/files)
	if strings.Contains(path, "/api/v") {
		parts := strings.Split(path, "/api/v")
		if len(parts) > 1 {
			remaining := parts[1]
			// Find the first slash to extract version
			slashIndex := strings.Index(remaining, "/")
			var versionStr string
			if slashIndex != -1 {
				versionStr = remaining[:slashIndex]
			} else {
				versionStr = remaining
			}

			if version, err := ParseVersion(versionStr); err == nil {
				if vc.IsVersionSupported(version) {
					return version
				}
			}
		}
	}
	return APIVersion{}
}

// VersionMiddleware is the HTTP middleware for version handling
func VersionMiddleware(config *VersionConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Negotiate version
			version := config.NegotiateVersion(r)

			// Add version to context
			ctx := context.WithValue(r.Context(), VersionContextKey{}, version)
			r = r.WithContext(ctx)

			// Add version headers to response
			w.Header().Set("X-API-Version", version.String())
			w.Header().Set("X-API-Current-Version", config.CurrentVersion.String())

			// Add deprecation warning if needed
			if config.IsVersionDeprecated(version) {
				if deprecationDate, exists := config.GetDeprecationDate(version); exists {
					w.Header().Set("X-API-Deprecated", "true")
					w.Header().Set("X-API-Deprecation-Date", deprecationDate.Format(time.RFC3339))
					w.Header().Set("Warning", fmt.Sprintf(`299 - "API version %s is deprecated and will be removed on %s"`,
						version.String(), deprecationDate.Format("2006-01-02")))
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

// GetVersionFromContext retrieves the API version from request context
func GetVersionFromContext(ctx context.Context) (APIVersion, bool) {
	version, ok := ctx.Value(VersionContextKey{}).(APIVersion)
	return version, ok
}

// RequireVersion is a helper to ensure minimum version requirements
func RequireVersion(ctx context.Context, required APIVersion) error {
	current, ok := GetVersionFromContext(ctx)
	if !ok {
		return fmt.Errorf("API version not found in context")
	}

	if current.Compare(required) < 0 {
		return fmt.Errorf("API version %s is required, but request uses %s",
			required.String(), current.String())
	}

	return nil
}
