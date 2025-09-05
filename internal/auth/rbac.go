package auth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Role represents a user role
type Role struct {
	ID          string       `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Permission represents a permission
type Permission struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Resource    string   `json:"resource"`
	Actions     []string `json:"actions"`
	Conditions  []string `json:"conditions,omitempty"`
}

// User represents a user
type User struct {
	ID        string                 `json:"id"`
	Username  string                 `json:"username"`
	Email     string                 `json:"email"`
	Roles     []string               `json:"roles"`
	Groups    []string               `json:"groups"`
	Metadata  map[string]interface{} `json:"metadata"`
	CreatedAt time.Time              `json:"created_at"`
	UpdatedAt time.Time              `json:"updated_at"`
	LastLogin time.Time              `json:"last_login,omitempty"`
	Active    bool                   `json:"active"`
}

// Group represents a user group
type Group struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Roles       []string  `json:"roles"`
	Members     []string  `json:"members"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// AccessRequest represents an access request
type AccessRequest struct {
	UserID      string                 `json:"user_id"`
	Resource    string                 `json:"resource"`
	Action      string                 `json:"action"`
	Context     map[string]interface{} `json:"context,omitempty"`
	RequestedAt time.Time              `json:"requested_at"`
}

// AccessDecision represents an access decision
type AccessDecision struct {
	Granted   bool                   `json:"granted"`
	Reason    string                 `json:"reason"`
	Policy    string                 `json:"policy,omitempty"`
	Context   map[string]interface{} `json:"context,omitempty"`
	DecidedAt time.Time              `json:"decided_at"`
}

// Policy represents an access control policy
type Policy struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Effect      string                 `json:"effect"`   // allow, deny
	Subjects    []string               `json:"subjects"` // users, roles, groups
	Resources   []string               `json:"resources"`
	Actions     []string               `json:"actions"`
	Conditions  map[string]interface{} `json:"conditions,omitempty"`
	Priority    int                    `json:"priority"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RBACManager manages role-based access control
type RBACManager struct {
	users    map[string]*User
	roles    map[string]*Role
	groups   map[string]*Group
	policies map[string]*Policy
	mu       sync.RWMutex
}

// NewRBACManager creates a new RBAC manager
func NewRBACManager() *RBACManager {
	rbac := &RBACManager{
		users:    make(map[string]*User),
		roles:    make(map[string]*Role),
		groups:   make(map[string]*Group),
		policies: make(map[string]*Policy),
	}

	// Initialize default roles and permissions
	rbac.initializeDefaults()

	return rbac
}

// initializeDefaults initializes default roles and permissions
func (rm *RBACManager) initializeDefaults() {
	// Create default permissions
	permissions := []Permission{
		{
			ID:          "read_files",
			Name:        "Read Files",
			Description: "Read files and directories",
			Resource:    "files",
			Actions:     []string{"read", "list"},
		},
		{
			ID:          "write_files",
			Name:        "Write Files",
			Description: "Create, update, and delete files",
			Resource:    "files",
			Actions:     []string{"create", "update", "delete"},
		},
		{
			ID:          "admin_system",
			Name:        "Admin System",
			Description: "Administrative access to system",
			Resource:    "system",
			Actions:     []string{"*"},
		},
		{
			ID:          "manage_users",
			Name:        "Manage Users",
			Description: "Manage users and permissions",
			Resource:    "users",
			Actions:     []string{"create", "read", "update", "delete"},
		},
		{
			ID:          "view_audit",
			Name:        "View Audit Logs",
			Description: "View audit logs and reports",
			Resource:    "audit",
			Actions:     []string{"read"},
		},
	}

	// Create default roles
	roles := []Role{
		{
			ID:          "admin",
			Name:        "Administrator",
			Description: "Full system access",
			Permissions: permissions,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "user",
			Name:        "User",
			Description: "Standard user access",
			Permissions: []Permission{permissions[0]}, // read_files
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "editor",
			Name:        "Editor",
			Description: "File editing access",
			Permissions: []Permission{permissions[0], permissions[1]}, // read_files, write_files
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "auditor",
			Name:        "Auditor",
			Description: "Audit and compliance access",
			Permissions: []Permission{permissions[0], permissions[4]}, // read_files, view_audit
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Add roles to manager
	for _, role := range roles {
		rm.roles[role.ID] = &role
	}

	// Create default policies
	policies := []Policy{
		{
			ID:          "admin_policy",
			Name:        "Administrator Policy",
			Description: "Full access for administrators",
			Effect:      "allow",
			Subjects:    []string{"admin"},
			Resources:   []string{"*"},
			Actions:     []string{"*"},
			Priority:    100,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "user_policy",
			Name:        "User Policy",
			Description: "Standard user access",
			Effect:      "allow",
			Subjects:    []string{"user", "editor"},
			Resources:   []string{"files"},
			Actions:     []string{"read", "create", "update", "delete"},
			Priority:    50,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          "audit_policy",
			Name:        "Audit Policy",
			Description: "Audit access policy",
			Effect:      "allow",
			Subjects:    []string{"auditor"},
			Resources:   []string{"audit", "files"},
			Actions:     []string{"read"},
			Priority:    75,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	// Add policies to manager
	for _, policy := range policies {
		rm.policies[policy.ID] = &policy
	}
}

// CreateUser creates a new user
func (rm *RBACManager) CreateUser(user *User) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.users[user.ID]; exists {
		return fmt.Errorf("user %s already exists", user.ID)
	}

	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.Active = true

	rm.users[user.ID] = user
	return nil
}

// GetUser retrieves a user by ID
func (rm *RBACManager) GetUser(userID string) (*User, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	user, exists := rm.users[userID]
	return user, exists
}

// UpdateUser updates an existing user
func (rm *RBACManager) UpdateUser(user *User) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.users[user.ID]; !exists {
		return fmt.Errorf("user %s not found", user.ID)
	}

	user.UpdatedAt = time.Now()
	rm.users[user.ID] = user
	return nil
}

// DeleteUser deletes a user
func (rm *RBACManager) DeleteUser(userID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.users[userID]; !exists {
		return fmt.Errorf("user %s not found", userID)
	}

	delete(rm.users, userID)
	return nil
}

// CreateRole creates a new role
func (rm *RBACManager) CreateRole(role *Role) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.roles[role.ID]; exists {
		return fmt.Errorf("role %s already exists", role.ID)
	}

	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()

	rm.roles[role.ID] = role
	return nil
}

// GetRole retrieves a role by ID
func (rm *RBACManager) GetRole(roleID string) (*Role, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	role, exists := rm.roles[roleID]
	return role, exists
}

// UpdateRole updates an existing role
func (rm *RBACManager) UpdateRole(role *Role) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.roles[role.ID]; !exists {
		return fmt.Errorf("role %s not found", role.ID)
	}

	role.UpdatedAt = time.Now()
	rm.roles[role.ID] = role
	return nil
}

// DeleteRole deletes a role
func (rm *RBACManager) DeleteRole(roleID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.roles[roleID]; !exists {
		return fmt.Errorf("role %s not found", roleID)
	}

	delete(rm.roles, roleID)
	return nil
}

// CreateGroup creates a new group
func (rm *RBACManager) CreateGroup(group *Group) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.groups[group.ID]; exists {
		return fmt.Errorf("group %s already exists", group.ID)
	}

	group.CreatedAt = time.Now()
	group.UpdatedAt = time.Now()

	rm.groups[group.ID] = group
	return nil
}

// GetGroup retrieves a group by ID
func (rm *RBACManager) GetGroup(groupID string) (*Group, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	group, exists := rm.groups[groupID]
	return group, exists
}

// UpdateGroup updates an existing group
func (rm *RBACManager) UpdateGroup(group *Group) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.groups[group.ID]; !exists {
		return fmt.Errorf("group %s not found", group.ID)
	}

	group.UpdatedAt = time.Now()
	rm.groups[group.ID] = group
	return nil
}

// DeleteGroup deletes a group
func (rm *RBACManager) DeleteGroup(groupID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.groups[groupID]; !exists {
		return fmt.Errorf("group %s not found", groupID)
	}

	delete(rm.groups, groupID)
	return nil
}

// CreatePolicy creates a new policy
func (rm *RBACManager) CreatePolicy(policy *Policy) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.policies[policy.ID]; exists {
		return fmt.Errorf("policy %s already exists", policy.ID)
	}

	policy.CreatedAt = time.Now()
	policy.UpdatedAt = time.Now()

	rm.policies[policy.ID] = policy
	return nil
}

// GetPolicy retrieves a policy by ID
func (rm *RBACManager) GetPolicy(policyID string) (*Policy, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	policy, exists := rm.policies[policyID]
	return policy, exists
}

// UpdatePolicy updates an existing policy
func (rm *RBACManager) UpdatePolicy(policy *Policy) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.policies[policy.ID]; !exists {
		return fmt.Errorf("policy %s not found", policy.ID)
	}

	policy.UpdatedAt = time.Now()
	rm.policies[policy.ID] = policy
	return nil
}

// DeletePolicy deletes a policy
func (rm *RBACManager) DeletePolicy(policyID string) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if _, exists := rm.policies[policyID]; !exists {
		return fmt.Errorf("policy %s not found", policyID)
	}

	delete(rm.policies, policyID)
	return nil
}

// CheckAccess checks if a user has access to a resource
func (rm *RBACManager) CheckAccess(ctx context.Context, request *AccessRequest) (*AccessDecision, error) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	// Get user
	user, exists := rm.users[request.UserID]
	if !exists {
		return &AccessDecision{
			Granted:   false,
			Reason:    "user not found",
			DecidedAt: time.Now(),
		}, nil
	}

	// Check if user is active
	if !user.Active {
		return &AccessDecision{
			Granted:   false,
			Reason:    "user is inactive",
			DecidedAt: time.Now(),
		}, nil
	}

	// Get user's effective roles
	effectiveRoles := rm.getEffectiveRoles(user)

	// Check policies
	decision := rm.evaluatePolicies(effectiveRoles, request)

	return decision, nil
}

// getEffectiveRoles gets all roles for a user (direct roles + group roles)
func (rm *RBACManager) getEffectiveRoles(user *User) []string {
	roles := make(map[string]bool)

	// Add direct roles
	for _, roleID := range user.Roles {
		roles[roleID] = true
	}

	// Add group roles
	for _, groupID := range user.Groups {
		if group, exists := rm.groups[groupID]; exists {
			for _, roleID := range group.Roles {
				roles[roleID] = true
			}
		}
	}

	// Convert to slice
	var effectiveRoles []string
	for roleID := range roles {
		effectiveRoles = append(effectiveRoles, roleID)
	}

	return effectiveRoles
}

// evaluatePolicies evaluates policies against the access request
func (rm *RBACManager) evaluatePolicies(roles []string, request *AccessRequest) *AccessDecision {
	// Sort policies by priority (highest first)
	var sortedPolicies []*Policy
	for _, policy := range rm.policies {
		sortedPolicies = append(sortedPolicies, policy)
	}

	// Simple sort by priority (in a real implementation, use proper sorting)
	for i := 0; i < len(sortedPolicies); i++ {
		for j := i + 1; j < len(sortedPolicies); j++ {
			if sortedPolicies[i].Priority < sortedPolicies[j].Priority {
				sortedPolicies[i], sortedPolicies[j] = sortedPolicies[j], sortedPolicies[i]
			}
		}
	}

	// Evaluate policies
	for _, policy := range sortedPolicies {
		if rm.matchesPolicy(policy, roles, request) {
			return &AccessDecision{
				Granted:   policy.Effect == "allow",
				Reason:    fmt.Sprintf("policy %s matched", policy.Name),
				Policy:    policy.ID,
				DecidedAt: time.Now(),
			}
		}
	}

	// Default deny
	return &AccessDecision{
		Granted:   false,
		Reason:    "no matching policy found",
		DecidedAt: time.Now(),
	}
}

// matchesPolicy checks if a policy matches the request
func (rm *RBACManager) matchesPolicy(policy *Policy, roles []string, request *AccessRequest) bool {
	// Check subjects (roles)
	subjectMatch := false
	for _, subject := range policy.Subjects {
		if subject == "*" {
			subjectMatch = true
			break
		}
		for _, role := range roles {
			if subject == role {
				subjectMatch = true
				break
			}
		}
		if subjectMatch {
			break
		}
	}

	if !subjectMatch {
		return false
	}

	// Check resources
	resourceMatch := false
	for _, resource := range policy.Resources {
		if resource == "*" || resource == request.Resource {
			resourceMatch = true
			break
		}
	}

	if !resourceMatch {
		return false
	}

	// Check actions
	actionMatch := false
	for _, action := range policy.Actions {
		if action == "*" || action == request.Action {
			actionMatch = true
			break
		}
	}

	return actionMatch
}

// ListUsers returns all users
func (rm *RBACManager) ListUsers() []*User {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var users []*User
	for _, user := range rm.users {
		users = append(users, user)
	}

	return users
}

// ListRoles returns all roles
func (rm *RBACManager) ListRoles() []*Role {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var roles []*Role
	for _, role := range rm.roles {
		roles = append(roles, role)
	}

	return roles
}

// ListGroups returns all groups
func (rm *RBACManager) ListGroups() []*Group {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var groups []*Group
	for _, group := range rm.groups {
		groups = append(groups, group)
	}

	return groups
}

// ListPolicies returns all policies
func (rm *RBACManager) ListPolicies() []*Policy {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var policies []*Policy
	for _, policy := range rm.policies {
		policies = append(policies, policy)
	}

	return policies
}

// Global RBAC manager instance
var GlobalRBACManager = NewRBACManager()

// Convenience functions
func CreateUser(user *User) error {
	return GlobalRBACManager.CreateUser(user)
}

func GetUser(userID string) (*User, bool) {
	return GlobalRBACManager.GetUser(userID)
}

func UpdateUser(user *User) error {
	return GlobalRBACManager.UpdateUser(user)
}

func DeleteUser(userID string) error {
	return GlobalRBACManager.DeleteUser(userID)
}

func CheckAccess(ctx context.Context, request *AccessRequest) (*AccessDecision, error) {
	return GlobalRBACManager.CheckAccess(ctx, request)
}

func CreateRole(role *Role) error {
	return GlobalRBACManager.CreateRole(role)
}

func GetRole(roleID string) (*Role, bool) {
	return GlobalRBACManager.GetRole(roleID)
}

func CreatePolicy(policy *Policy) error {
	return GlobalRBACManager.CreatePolicy(policy)
}

func GetPolicy(policyID string) (*Policy, bool) {
	return GlobalRBACManager.GetPolicy(policyID)
}
