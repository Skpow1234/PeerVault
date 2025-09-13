package auth

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRBACManager(t *testing.T) {
	rm := NewRBACManager()
	assert.NotNil(t, rm)
	assert.NotNil(t, rm.users)
	assert.NotNil(t, rm.roles)
	assert.NotNil(t, rm.groups)
	assert.NotNil(t, rm.policies)

	// Check that default roles are initialized
	_, exists := rm.GetRole("admin")
	assert.True(t, exists, "admin role should be initialized")

	_, exists = rm.GetRole("user")
	assert.True(t, exists, "user role should be initialized")

	_, exists = rm.GetRole("editor")
	assert.True(t, exists, "editor role should be initialized")

	_, exists = rm.GetRole("auditor")
	assert.True(t, exists, "auditor role should be initialized")
}

func TestRBACManager_UserCRUD(t *testing.T) {
	rm := NewRBACManager()

	user := &User{
		ID:       "test-user",
		Username: "testuser",
		Email:    "test@example.com",
		Roles:    []string{"user"},
		Groups:   []string{},
		Active:   true,
	}

	// Create user
	err := rm.CreateUser(user)
	require.NoError(t, err)

	// Get user
	retrievedUser, exists := rm.GetUser("test-user")
	assert.True(t, exists)
	assert.Equal(t, "testuser", retrievedUser.Username)
	assert.Equal(t, "test@example.com", retrievedUser.Email)
	assert.True(t, retrievedUser.Active)

	// Update user
	user.Username = "updateduser"
	err = rm.UpdateUser(user)
	require.NoError(t, err)

	updatedUser, exists := rm.GetUser("test-user")
	assert.True(t, exists)
	assert.Equal(t, "updateduser", updatedUser.Username)

	// Delete user
	err = rm.DeleteUser("test-user")
	require.NoError(t, err)

	_, exists = rm.GetUser("test-user")
	assert.False(t, exists)
}

func TestRBACManager_UserCRUD_Errors(t *testing.T) {
	rm := NewRBACManager()

	user := &User{
		ID:       "test-user",
		Username: "testuser",
	}

	// Create user
	err := rm.CreateUser(user)
	require.NoError(t, err)

	// Try to create user with same ID
	err = rm.CreateUser(user)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")

	// Try to update non-existent user
	nonExistentUser := &User{
		ID:       "non-existent",
		Username: "nonexistent",
	}
	err = rm.UpdateUser(nonExistentUser)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Try to delete non-existent user
	err = rm.DeleteUser("non-existent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRBACManager_RoleCRUD(t *testing.T) {
	rm := NewRBACManager()

	role := &Role{
		ID:          "test-role",
		Name:        "Test Role",
		Description: "A test role",
		Permissions: []Permission{
			{
				ID:       "test_perm",
				Name:     "Test Permission",
				Resource: "test",
				Actions:  []string{"read"},
			},
		},
	}

	// Create role
	err := rm.CreateRole(role)
	require.NoError(t, err)

	// Get role
	retrievedRole, exists := rm.GetRole("test-role")
	assert.True(t, exists)
	assert.Equal(t, "Test Role", retrievedRole.Name)
	assert.Equal(t, "A test role", retrievedRole.Description)
	assert.Len(t, retrievedRole.Permissions, 1)

	// Update role
	role.Name = "Updated Role"
	err = rm.UpdateRole(role)
	require.NoError(t, err)

	updatedRole, exists := rm.GetRole("test-role")
	assert.True(t, exists)
	assert.Equal(t, "Updated Role", updatedRole.Name)

	// Delete role
	err = rm.DeleteRole("test-role")
	require.NoError(t, err)

	_, exists = rm.GetRole("test-role")
	assert.False(t, exists)
}

func TestRBACManager_GroupCRUD(t *testing.T) {
	rm := NewRBACManager()

	group := &Group{
		ID:          "test-group",
		Name:        "Test Group",
		Description: "A test group",
		Roles:       []string{"user"},
		Members:     []string{"user1", "user2"},
	}

	// Create group
	err := rm.CreateGroup(group)
	require.NoError(t, err)

	// Get group
	retrievedGroup, exists := rm.GetGroup("test-group")
	assert.True(t, exists)
	assert.Equal(t, "Test Group", retrievedGroup.Name)
	assert.Equal(t, []string{"user"}, retrievedGroup.Roles)
	assert.Equal(t, []string{"user1", "user2"}, retrievedGroup.Members)

	// Update group
	group.Name = "Updated Group"
	err = rm.UpdateGroup(group)
	require.NoError(t, err)

	updatedGroup, exists := rm.GetGroup("test-group")
	assert.True(t, exists)
	assert.Equal(t, "Updated Group", updatedGroup.Name)

	// Delete group
	err = rm.DeleteGroup("test-group")
	require.NoError(t, err)

	_, exists = rm.GetGroup("test-group")
	assert.False(t, exists)
}

func TestRBACManager_PolicyCRUD(t *testing.T) {
	rm := NewRBACManager()

	policy := &Policy{
		ID:          "test-policy",
		Name:        "Test Policy",
		Description: "A test policy",
		Effect:      "allow",
		Subjects:    []string{"user"},
		Resources:   []string{"files"},
		Actions:     []string{"read"},
		Priority:    10,
	}

	// Create policy
	err := rm.CreatePolicy(policy)
	require.NoError(t, err)

	// Get policy
	retrievedPolicy, exists := rm.GetPolicy("test-policy")
	assert.True(t, exists)
	assert.Equal(t, "Test Policy", retrievedPolicy.Name)
	assert.Equal(t, "allow", retrievedPolicy.Effect)
	assert.Equal(t, []string{"user"}, retrievedPolicy.Subjects)
	assert.Equal(t, []string{"files"}, retrievedPolicy.Resources)
	assert.Equal(t, []string{"read"}, retrievedPolicy.Actions)

	// Update policy
	policy.Name = "Updated Policy"
	err = rm.UpdatePolicy(policy)
	require.NoError(t, err)

	updatedPolicy, exists := rm.GetPolicy("test-policy")
	assert.True(t, exists)
	assert.Equal(t, "Updated Policy", updatedPolicy.Name)

	// Delete policy
	err = rm.DeletePolicy("test-policy")
	require.NoError(t, err)

	_, exists = rm.GetPolicy("test-policy")
	assert.False(t, exists)
}

func TestRBACManager_CheckAccess(t *testing.T) {
	rm := NewRBACManager()

	// Create a test user
	user := &User{
		ID:       "test-user",
		Username: "testuser",
		Roles:    []string{"user"},
		Active:   true,
	}
	err := rm.CreateUser(user)
	require.NoError(t, err)

	ctx := context.Background()

	// Test access request that should be granted
	request := &AccessRequest{
		UserID:      "test-user",
		Resource:    "files",
		Action:      "read",
		RequestedAt: time.Now(),
	}

	decision, err := rm.CheckAccess(ctx, request)
	require.NoError(t, err)
	assert.True(t, decision.Granted)
	assert.Contains(t, decision.Reason, "policy")

	// Test access request that should be denied
	request.Action = "admin" // user role doesn't have admin permission
	decision, err = rm.CheckAccess(ctx, request)
	require.NoError(t, err)
	assert.False(t, decision.Granted)

	// Test access for non-existent user
	request.UserID = "non-existent"
	decision, err = rm.CheckAccess(ctx, request)
	require.NoError(t, err)
	assert.False(t, decision.Granted)
	assert.Contains(t, decision.Reason, "user not found")

	// Test access for inactive user
	user.Active = false
	err = rm.UpdateUser(user)
	require.NoError(t, err)

	request.UserID = "test-user"
	decision, err = rm.CheckAccess(ctx, request)
	require.NoError(t, err)
	assert.False(t, decision.Granted)
	assert.Contains(t, decision.Reason, "user is inactive")
}

func TestRBACManager_GetEffectiveRoles(t *testing.T) {
	rm := NewRBACManager()

	// Create a group with roles
	group := &Group{
		ID:    "test-group",
		Name:  "Test Group",
		Roles: []string{"editor", "auditor"},
	}
	err := rm.CreateGroup(group)
	require.NoError(t, err)

	// Create a user with direct roles and group membership
	user := &User{
		ID:     "test-user",
		Roles:  []string{"user"},
		Groups: []string{"test-group"},
		Active: true,
	}
	err = rm.CreateUser(user)
	require.NoError(t, err)

	// Get effective roles
	effectiveRoles := rm.getEffectiveRoles(user)

	// Should include direct roles and group roles
	assert.Contains(t, effectiveRoles, "user")
	assert.Contains(t, effectiveRoles, "editor")
	assert.Contains(t, effectiveRoles, "auditor")

	// Should not have duplicates
	roleCount := make(map[string]int)
	for _, role := range effectiveRoles {
		roleCount[role]++
	}
	for _, count := range roleCount {
		assert.Equal(t, 1, count, "Each role should appear only once")
	}
}

func TestRBACManager_ListOperations(t *testing.T) {
	rm := NewRBACManager()

	// Create test data
	user := &User{
		ID:       "test-user",
		Username: "testuser",
		Active:   true,
	}
	err := rm.CreateUser(user)
	require.NoError(t, err)

	role := &Role{
		ID:          "test-role",
		Name:        "Test Role",
		Description: "A test role",
	}
	err = rm.CreateRole(role)
	require.NoError(t, err)

	group := &Group{
		ID:          "test-group",
		Name:        "Test Group",
		Description: "A test group",
	}
	err = rm.CreateGroup(group)
	require.NoError(t, err)

	policy := &Policy{
		ID:          "test-policy",
		Name:        "Test Policy",
		Description: "A test policy",
		Effect:      "allow",
		Subjects:    []string{"user"},
		Resources:   []string{"files"},
		Actions:     []string{"read"},
		Priority:    10,
	}
	err = rm.CreatePolicy(policy)
	require.NoError(t, err)

	// Test list operations
	users := rm.ListUsers()
	assert.Len(t, users, 1) // 1 created user

	roles := rm.ListRoles()
	assert.Greater(t, len(roles), 4) // Should include default roles + created

	groups := rm.ListGroups()
	assert.Greater(t, len(groups), 0)

	policies := rm.ListPolicies()
	assert.Greater(t, len(policies), 3) // Should include default policies + created
}

func TestRBACManager_MatchesPolicy(t *testing.T) {
	rm := NewRBACManager()

	// Test policy matching
	policy := &Policy{
		Effect:    "allow",
		Subjects:  []string{"user", "admin"},
		Resources: []string{"files", "system"},
		Actions:   []string{"read", "write"},
	}

	roles := []string{"user", "editor"}

	// Test matching request
	request := &AccessRequest{
		UserID:   "test-user",
		Resource: "files",
		Action:   "read",
	}

	matches := rm.matchesPolicy(policy, roles, request)
	assert.True(t, matches)

	// Test non-matching subject
	roles = []string{"non-existent"}
	matches = rm.matchesPolicy(policy, roles, request)
	assert.False(t, matches)

	// Test non-matching resource
	request.Resource = "non-existent"
	roles = []string{"user"}
	matches = rm.matchesPolicy(policy, roles, request)
	assert.False(t, matches)

	// Test non-matching action
	request.Resource = "files"
	request.Action = "delete"
	matches = rm.matchesPolicy(policy, roles, request)
	assert.False(t, matches)

	// Test wildcard matching
	policy.Resources = []string{"*"}
	policy.Actions = []string{"*"}
	request.Resource = "anything"
	request.Action = "anything"

	matches = rm.matchesPolicy(policy, roles, request)
	assert.True(t, matches)
}

func TestGlobalRBACManager(t *testing.T) {
	// Test that global functions work
	user := &User{
		ID:       "global-test-user",
		Username: "globaltest",
		Roles:    []string{"user"},
		Active:   true,
	}

	err := CreateUser(user)
	require.NoError(t, err)

	retrievedUser, exists := GetUser("global-test-user")
	assert.True(t, exists)
	assert.Equal(t, "globaltest", retrievedUser.Username)

	// Clean up
	err = DeleteUser("global-test-user")
	require.NoError(t, err)
}

func TestAccessDecision(t *testing.T) {
	decision := &AccessDecision{
		Granted:   true,
		Reason:    "test reason",
		Policy:    "test-policy",
		DecidedAt: time.Now(),
	}

	assert.True(t, decision.Granted)
	assert.Equal(t, "test reason", decision.Reason)
	assert.Equal(t, "test-policy", decision.Policy)
	assert.NotZero(t, decision.DecidedAt)
}

func TestPermission_Struct(t *testing.T) {
	perm := Permission{
		ID:          "test-perm",
		Name:        "Test Permission",
		Description: "A test permission",
		Resource:    "test-resource",
		Actions:     []string{"read", "write"},
		Conditions:  []string{"condition1", "condition2"},
	}

	assert.Equal(t, "test-perm", perm.ID)
	assert.Equal(t, "Test Permission", perm.Name)
	assert.Equal(t, "A test permission", perm.Description)
	assert.Equal(t, "test-resource", perm.Resource)
	assert.Equal(t, []string{"read", "write"}, perm.Actions)
	assert.Equal(t, []string{"condition1", "condition2"}, perm.Conditions)
}

func TestRole_Struct(t *testing.T) {
	now := time.Now()
	role := Role{
		ID:          "test-role",
		Name:        "Test Role",
		Description: "A test role",
		Permissions: []Permission{
			{ID: "perm1", Name: "Permission 1"},
		},
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, "test-role", role.ID)
	assert.Equal(t, "Test Role", role.Name)
	assert.Equal(t, "A test role", role.Description)
	assert.Len(t, role.Permissions, 1)
	assert.Equal(t, now, role.CreatedAt)
	assert.Equal(t, now, role.UpdatedAt)
}

func TestUser_Struct(t *testing.T) {
	now := time.Now()
	user := User{
		ID:        "test-user",
		Username:  "testuser",
		Email:     "test@example.com",
		Roles:     []string{"user", "editor"},
		Groups:    []string{"group1"},
		Active:    true,
		CreatedAt: now,
		UpdatedAt: now,
		LastLogin: now,
	}

	assert.Equal(t, "test-user", user.ID)
	assert.Equal(t, "testuser", user.Username)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, []string{"user", "editor"}, user.Roles)
	assert.Equal(t, []string{"group1"}, user.Groups)
	assert.True(t, user.Active)
	assert.Equal(t, now, user.CreatedAt)
	assert.Equal(t, now, user.UpdatedAt)
	assert.Equal(t, now, user.LastLogin)
}

func TestGroup_Struct(t *testing.T) {
	now := time.Now()
	group := Group{
		ID:          "test-group",
		Name:        "Test Group",
		Description: "A test group",
		Roles:       []string{"user"},
		Members:     []string{"user1", "user2"},
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	assert.Equal(t, "test-group", group.ID)
	assert.Equal(t, "Test Group", group.Name)
	assert.Equal(t, "A test group", group.Description)
	assert.Equal(t, []string{"user"}, group.Roles)
	assert.Equal(t, []string{"user1", "user2"}, group.Members)
	assert.Equal(t, now, group.CreatedAt)
	assert.Equal(t, now, group.UpdatedAt)
}

func TestPolicy_Struct(t *testing.T) {
	now := time.Now()
	policy := Policy{
		ID:          "test-policy",
		Name:        "Test Policy",
		Description: "A test policy",
		Effect:      "allow",
		Subjects:    []string{"user"},
		Resources:   []string{"files"},
		Actions:     []string{"read", "write"},
		Conditions: map[string]interface{}{
			"time": "business_hours",
		},
		Priority:  10,
		CreatedAt: now,
		UpdatedAt: now,
	}

	assert.Equal(t, "test-policy", policy.ID)
	assert.Equal(t, "Test Policy", policy.Name)
	assert.Equal(t, "A test policy", policy.Description)
	assert.Equal(t, "allow", policy.Effect)
	assert.Equal(t, []string{"user"}, policy.Subjects)
	assert.Equal(t, []string{"files"}, policy.Resources)
	assert.Equal(t, []string{"read", "write"}, policy.Actions)
	assert.Equal(t, 10, policy.Priority)
	assert.Equal(t, now, policy.CreatedAt)
	assert.Equal(t, now, policy.UpdatedAt)
	assert.NotNil(t, policy.Conditions)
}
