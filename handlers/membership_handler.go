package handlers

import (
	"ms-user/config"
	"ms-user/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// MembershipHandler handles HTTP requests for user-group membership operations.
// It leverages the KeycloakService to interact with Keycloak's Admin API for membership management.
type MembershipHandler struct {
	keycloakService *services.KeycloakService
}

// NewMembershipHandler creates a new MembershipHandler instance.
// It initializes a KeycloakService using the provided configuration.
func NewMembershipHandler(cfg *config.Config) *MembershipHandler {
	return &MembershipHandler{
		keycloakService: services.NewKeycloakService(cfg),
	}
}

// ListUserGroups handles the HTTP GET request for retrieving all groups that a given user belongs to.
// Endpoint: GET /users/:id/groups
//
// Input:
//   - userID from URL path parameter.
//
// Output:
//   - On success: HTTP 200 with a JSON array of groups.
//   - On error: An error message with HTTP 500.
func (h *MembershipHandler) ListUserGroups(c *gin.Context) {
	userID := c.Param("id")
	groups, err := h.keycloakService.ListUserGroups(userID)
	if err != nil {
		log.Error().Err(err).Msg("Error listing groups for user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// AddUserToGroup handles the HTTP PUT request to assign a user to a group.
// Endpoint: PUT /users/:id/groups/:groupId
//
// Input:
//   - userID and groupID from URL path parameters.
//
// Output:
//   - On success: HTTP 204 No Content.
//   - On error: An error message with HTTP 500.
func (h *MembershipHandler) AddUserToGroup(c *gin.Context) {
	userID := c.Param("id")
	groupID := c.Param("groupId")
	err := h.keycloakService.AddUserToGroup(userID, groupID)
	if err != nil {
		log.Error().Err(err).Msg("Error adding user to group")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// AddUserToGroupByEmail handles the HTTP PUT request to add a user (searched by email) to a group.
// Endpoint: PUT /ms-user/v1/users/email/:email/groups/:groupId
//
// Input:
//   - email: provided as a URL path parameter.
//   - groupId: provided as a URL path parameter.
//
// Output:
//   - On success: HTTP 204 No Content.
//   - On error: Returns an appropriate HTTP status and error message.
func (h *MembershipHandler) AddUserToGroupByEmail(c *gin.Context) {
	email := c.Param("email")
	groupID := c.Param("groupId")

	if email == "" || groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email and groupId are required"})
		return
	}

	// Search for the user by email.
	users, err := h.keycloakService.SearchUserByEmail(email)
	if err != nil {
		log.Error().Err(err).Msg("Error searching user by email")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(users) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no user found with the provided email"})
		return
	}
	if len(users) > 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "multiple users found with the provided email"})
		return
	}

	// Use the found user's ID to add the user to the group.
	userID := users[0].ID
	err = h.keycloakService.AddUserToGroup(userID, groupID)
	if err != nil {
		log.Error().Err(err).Msg("Error adding user to group by email")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// RemoveUserFromGroup handles the HTTP DELETE request to remove a user from a group.
// Endpoint: DELETE /users/:id/groups/:groupId
//
// Input:
//   - userID and groupID from URL path parameters.
//
// Output:
//   - On success: HTTP 204 No Content.
//   - On error: An error message with HTTP 500.
func (h *MembershipHandler) RemoveUserFromGroup(c *gin.Context) {
	userID := c.Param("id")
	groupID := c.Param("groupId")
	err := h.keycloakService.RemoveUserFromGroup(userID, groupID)
	if err != nil {
		log.Error().Err(err).Msg("Error removing user from group")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// ListGroupUsers handles the HTTP GET request for retrieving all users that are members of a specific group.
// Endpoint: GET /groups/:id/users
//
// Input:
//   - groupID from URL path parameter.
//
// Output:
//   - On success: HTTP 200 with a JSON array of users.
//   - On error: An error message with HTTP 500.
func (h *MembershipHandler) ListGroupUsers(c *gin.Context) {
	groupID := c.Param("id")
	users, err := h.keycloakService.ListGroupUsers(groupID)
	if err != nil {
		log.Error().Err(err).Msg("Error listing users in group")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// SetKeycloakService overrides the underlying KeycloakService (useful for testing).
func (h *MembershipHandler) SetKeycloakService(svc *services.KeycloakService) {
	h.keycloakService = svc
}
