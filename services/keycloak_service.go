package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"ms-user/config"
	"ms-user/models"
	"net/http"

	"github.com/rs/zerolog/log"
)

// KeycloakService handles all interactions with Keycloak's Admin API.
// It manages token retrieval and refresh as well as CRUD operations for users, groups,
// and membership management.
type KeycloakService struct {
	config *config.Config
	client *http.Client
	token  string // Admin token used for authorization; token refresh logic is implemented.
}

// NewKeycloakService initializes a new KeycloakService with the provided configuration.
// It fetches an initial admin token and sets up the HTTP client.
func NewKeycloakService(cfg *config.Config) *KeycloakService {
	service := &KeycloakService{
		config: cfg,
		client: &http.Client{},
	}
	// Fetch initial admin token from Keycloak.
	token, err := service.getAdminToken()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get admin token from Keycloak")
	}
	service.token = token
	return service
}

// doRequest executes an HTTP request with the current admin token.
// If a 401 Unauthorized response is received, it refreshes the token and retries once.
// It returns the HTTP response or an error if the request ultimately fails.
//
// Input: A pointer to an http.Request (with no authorization header set).
// Output: *http.Response if successful; error otherwise.
func (k *KeycloakService) doRequest(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+k.token)
	resp, err := k.client.Do(req)
	if err != nil {
		return nil, err
	}
	// If the token is expired or invalid, refresh the token and retry once.
	if resp.StatusCode == http.StatusUnauthorized {
		resp.Body.Close() // Ensure the response body is closed.
		log.Info().Msg("Token expired. Refreshing token and retrying request.")
		newToken, err := k.getAdminToken()
		if err != nil {
			return nil, fmt.Errorf("failed to refresh token: %v", err)
		}
		k.token = newToken
		req.Header.Set("Authorization", "Bearer "+k.token)
		return k.client.Do(req)
	}
	return resp, nil
}

// getAdminToken fetches an admin access token from Keycloak.
// It sends a POST request to the token endpoint using admin credentials.
// Returns the access token as a string, or an error if the process fails.
func (k *KeycloakService) getAdminToken() (string, error) {
	url := fmt.Sprintf("%s/realms/%s/protocol/openid-connect/token", k.config.KeycloakURL, k.config.KeycloakRealm)
	data := "grant_type=password&client_id=admin-cli&username=" + k.config.KeycloakUsername + "&password=" + k.config.KeycloakPassword
	req, err := http.NewRequest("POST", url, bytes.NewBufferString(data))
	if err != nil {
		return "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := k.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Check for a successful response.
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return "", fmt.Errorf("failed to get token, status: %d, response: %s %s", resp.StatusCode, string(bodyBytes), string(k.config.KeycloakUsername))
	}

	// Decode the JSON response.
	var result map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", err
	}
	token, ok := result["access_token"].(string)
	if !ok {
		return "", fmt.Errorf("access token not found")
	}
	return token, nil
}

// ---------------------- User CRUD operations ----------------------

// ListUsers retrieves all users from Keycloak.
// Input: None.
// Output: Slice of models.User if successful; error otherwise.
func (k *KeycloakService) ListUsers() ([]models.User, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/users", k.config.KeycloakURL, k.config.KeycloakRealm)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := k.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body.
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for non-OK status and parse error message if available.
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("failed to list users: status %d, unable to parse error", resp.StatusCode)
		}
		return nil, fmt.Errorf("failed to list users: %v", errResp)
	}

	var users []models.User
	if err := json.Unmarshal(body, &users); err != nil {
		log.Error().Msgf("Unable to decode response into []models.User: %s", string(body))
		return nil, fmt.Errorf("json: %v", err)
	}
	return users, nil
}

// CreateUser creates a new user in Keycloak.
// Input: models.User representing the user to create.
// Output: Pointer to models.User on success (Keycloak does not return the full object by default); error otherwise.
func (k *KeycloakService) CreateUser(user models.User) (*models.User, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/users", k.config.KeycloakURL, k.config.KeycloakRealm)
	payload, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := k.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Successful creation may return 201 or 204.
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create user, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}
	// Return the input user as Keycloak does not return the created object.
	return &user, nil
}

// GetUser retrieves a user by ID from Keycloak.
// Input: User ID (string).
// Output: Pointer to models.User if found; error otherwise.
func (k *KeycloakService) GetUser(id string) (*models.User, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/users/%s", k.config.KeycloakURL, k.config.KeycloakRealm, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := k.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user not found, status: %d", resp.StatusCode)
	}
	var user models.User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	return &user, nil
}

// SearchUserByEmail retrieves users from Keycloak matching the provided email.
// Input: email (string) to search for.
// Output: A slice of models.User if found; error otherwise.
func (k *KeycloakService) SearchUserByEmail(email string) ([]models.User, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/users?email=%s", k.config.KeycloakURL, k.config.KeycloakRealm, email)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := k.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for non-OK status and return error if necessary.
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("failed to search users: status %d, unable to parse error", resp.StatusCode)
		}
		return nil, fmt.Errorf("failed to search users: %v", errResp)
	}

	// Unmarshal the response into a slice of models.User.
	var users []models.User
	if err := json.Unmarshal(body, &users); err != nil {
		log.Error().Msgf("Unable to decode response into []models.User: %s", string(body))
		return nil, fmt.Errorf("json: %v", err)
	}
	return users, nil
}

// UpdateUser updates an existing user in Keycloak.
// Input: User ID (string) and models.User containing updated data.
// Output: Pointer to updated models.User on success; error otherwise.
func (k *KeycloakService) UpdateUser(id string, user models.User) (*models.User, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/users/%s", k.config.KeycloakURL, k.config.KeycloakRealm, id)
	payload, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := k.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update user, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}
	return &user, nil
}

// DeleteUser deletes a user by ID in Keycloak.
// Input: User ID (string).
// Output: error if deletion fails; nil otherwise.
func (k *KeycloakService) DeleteUser(id string) error {
	url := fmt.Sprintf("%s/admin/realms/%s/users/%s", k.config.KeycloakURL, k.config.KeycloakRealm, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := k.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete user, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// ---------------------- Group CRUD operations ----------------------

// ListGroupsWithUsers retrieves all groups and for each group, fetches its associated users.
// Output: a slice of models.GroupWithUsers; error otherwise.
func (k *KeycloakService) ListGroupsWithUsers() ([]models.GroupWithUsers, error) {
	groups, err := k.ListGroups()
	if err != nil {
		return nil, err
	}

	var result []models.GroupWithUsers
	for _, group := range groups {
		users, err := k.ListGroupUsers(group.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get users for group %s: %v", group.ID, err)
		}
		result = append(result, models.GroupWithUsers{
			Group: group,
			Users: users,
		})
	}
	return result, nil
}

// ListGroups retrieves all groups from Keycloak.
// Input: None.
// Output: Slice of models.Group if successful; error otherwise.
func (k *KeycloakService) ListGroups() ([]models.Group, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/groups", k.config.KeycloakURL, k.config.KeycloakRealm)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := k.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Check for non-OK status.
	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("failed to list groups: status %d, unable to parse error", resp.StatusCode)
		}
		return nil, fmt.Errorf("failed to list groups: %v", errResp)
	}

	var groups []models.Group
	if err := json.Unmarshal(body, &groups); err != nil {
		log.Error().Msgf("Unable to decode response into []models.Group: %s", string(body))
		return nil, fmt.Errorf("json: %v", err)
	}
	return groups, nil
}

// CreateGroup creates a new group in Keycloak.
// Input: models.Group representing the group to create.
// Output: Pointer to models.Group on success; error otherwise.
func (k *KeycloakService) CreateGroup(group models.Group) (*models.Group, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/groups", k.config.KeycloakURL, k.config.KeycloakRealm)
	payload, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := k.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to create group, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}
	return &group, nil
}

// GetGroup retrieves a group by ID from Keycloak.
// Input: Group ID (string).
// Output: Pointer to models.Group if found; error otherwise.
func (k *KeycloakService) GetGroup(id string) (*models.Group, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/groups/%s", k.config.KeycloakURL, k.config.KeycloakRealm, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := k.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("group not found, status: %d", resp.StatusCode)
	}
	var group models.Group
	if err := json.NewDecoder(resp.Body).Decode(&group); err != nil {
		return nil, err
	}
	return &group, nil
}

// UpdateGroup updates an existing group in Keycloak.
// Input: Group ID (string) and models.Group with updated data.
// Output: Pointer to models.Group on success; error otherwise.
func (k *KeycloakService) UpdateGroup(id string, group models.Group) (*models.Group, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/groups/%s", k.config.KeycloakURL, k.config.KeycloakRealm, id)
	payload, err := json.Marshal(group)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("PUT", url, bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")

	resp, err := k.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to update group, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}
	return &group, nil
}

// DeleteGroup deletes a group by ID in Keycloak.
// Input: Group ID (string).
// Output: error if deletion fails; nil otherwise.
func (k *KeycloakService) DeleteGroup(id string) error {
	url := fmt.Sprintf("%s/admin/realms/%s/groups/%s", k.config.KeycloakURL, k.config.KeycloakRealm, id)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := k.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete group, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// ---------------------- Membership functions ----------------------

// ListUserGroups retrieves all groups a given user is a member of from Keycloak.
// Input: User ID (string).
// Output: Slice of models.Group if successful; error otherwise.
func (k *KeycloakService) ListUserGroups(userID string) ([]models.Group, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/users/%s/groups", k.config.KeycloakURL, k.config.KeycloakRealm, userID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := k.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("failed to list user groups: status %d, unable to parse error", resp.StatusCode)
		}
		return nil, fmt.Errorf("failed to list user groups: %v", errResp)
	}

	var groups []models.Group
	if err := json.Unmarshal(body, &groups); err != nil {
		log.Error().Msgf("Unable to decode response into []models.Group: %s", string(body))
		return nil, fmt.Errorf("json: %v", err)
	}
	return groups, nil
}

// AddUserToGroup assigns a user to a specific group in Keycloak.
// Input: User ID and Group ID (both strings).
// Output: error if the operation fails; nil otherwise.
func (k *KeycloakService) AddUserToGroup(userID string, groupID string) error {
	url := fmt.Sprintf("%s/admin/realms/%s/users/%s/groups/%s", k.config.KeycloakURL, k.config.KeycloakRealm, userID, groupID)
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		return err
	}

	resp, err := k.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to add user to group, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// AddUserToGroupByEmail searches for a user by the provided email and, if exactly one user is found,
// adds that user to the specified group.
// Input: email (string) and groupID (string).
// Output: error if the operation fails; nil otherwise.
func (k *KeycloakService) AddUserToGroupByEmail(email, groupID string) error {
	// Search for the user by email.
	users, err := k.SearchUserByEmail(email)
	if err != nil {
		return fmt.Errorf("error searching user by email: %v", err)
	}
	if len(users) == 0 {
		return fmt.Errorf("no user found with the provided email")
	}
	if len(users) > 1 {
		return fmt.Errorf("multiple users found with the provided email")
	}
	// Use the found user's ID to add the user to the group.
	return k.AddUserToGroup(users[0].ID, groupID)
}

// RemoveUserFromGroup removes a user from a specific group in Keycloak.
// Input: User ID and Group ID (both strings).
// Output: error if the operation fails; nil otherwise.
func (k *KeycloakService) RemoveUserFromGroup(userID string, groupID string) error {
	url := fmt.Sprintf("%s/admin/realms/%s/users/%s/groups/%s", k.config.KeycloakURL, k.config.KeycloakRealm, userID, groupID)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	resp, err := k.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("failed to remove user from group, status: %d, response: %s", resp.StatusCode, string(bodyBytes))
	}
	return nil
}

// ListGroupUsers retrieves all users that are members of a specific group in Keycloak.
// Input: Group ID (string).
// Output: Slice of models.User if successful; error otherwise.
func (k *KeycloakService) ListGroupUsers(groupID string) ([]models.User, error) {
	url := fmt.Sprintf("%s/admin/realms/%s/groups/%s/members", k.config.KeycloakURL, k.config.KeycloakRealm, groupID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := k.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		if err := json.Unmarshal(body, &errResp); err != nil {
			return nil, fmt.Errorf("failed to list group users: status %d, unable to parse error", resp.StatusCode)
		}
		return nil, fmt.Errorf("failed to list group users: %v", errResp)
	}

	var users []models.User
	if err := json.Unmarshal(body, &users); err != nil {
		log.Error().Msgf("Unable to decode response into []models.User: %s", string(body))
		return nil, fmt.Errorf("json: %v", err)
	}
	return users, nil
}

// ---------------------- Testing Helpers ----------------------

// SetToken allows overriding the admin token (useful for testing).
func (k *KeycloakService) SetToken(token string) {
	k.token = token
}

// SetClient allows overriding the HTTP client (useful for testing).
func (k *KeycloakService) SetClient(client *http.Client) {
	k.client = client
}
