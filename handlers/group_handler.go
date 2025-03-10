package handlers

import (
	"ms-user/config"
	"ms-user/models"
	"ms-user/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GroupHandler handles HTTP requests for group-related operations.
// It leverages the KeycloakService to interact with Keycloak's Admin API.
type GroupHandler struct {
	keycloakService *services.KeycloakService
}

// NewGroupHandler creates and returns a new GroupHandler instance.
// It initializes a new KeycloakService with the provided configuration.
func NewGroupHandler(cfg *config.Config) *GroupHandler {
	return &GroupHandler{
		keycloakService: services.NewKeycloakService(cfg),
	}
}

// ListGroups handles the HTTP GET request for retrieving all groups.
// It calls the KeycloakService.ListGroups method and returns the result.
// On success, it responds with HTTP 200 and the list of groups.
// On error, it logs the error and responds with HTTP 500.
func (h *GroupHandler) ListGroups(c *gin.Context) {
	groups, err := h.keycloakService.ListGroups()
	if err != nil {
		log.Error().Err(err).Msg("Error listing groups")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groups)
}

// CreateGroup handles the HTTP POST request for creating a new group.
// It expects a valid JSON body that matches the models.Group structure.
// On success, it responds with HTTP 201 and the created group.
// On validation error, it responds with HTTP 400, or HTTP 500 for internal errors.
func (h *GroupHandler) CreateGroup(c *gin.Context) {
	var group models.Group
	// Bind the incoming JSON payload to the group model.
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createdGroup, err := h.keycloakService.CreateGroup(group)
	if err != nil {
		log.Error().Err(err).Msg("Error creating group")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, createdGroup)
}

// ListGroupsWithUsers handles GET /groups/with-users.
// It retrieves all groups along with their associated users.
func (h *GroupHandler) ListGroupsWithUsers(c *gin.Context) {
	groupsWithUsers, err := h.keycloakService.ListGroupsWithUsers()
	if err != nil {
		log.Error().Err(err).Msg("Error listing groups with users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, groupsWithUsers)
}

// GetGroup handles the HTTP GET request for retrieving a specific group by ID.
// It expects the group ID as a path parameter.
// On success, it responds with HTTP 200 and the group details.
// If the group is not found, it responds with HTTP 404.
func (h *GroupHandler) GetGroup(c *gin.Context) {
	id := c.Param("id")
	group, err := h.keycloakService.GetGroup(id)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching group")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, group)
}

// UpdateGroup handles the HTTP PUT request for updating an existing group.
// It expects the group ID as a path parameter and a valid JSON body with the updated data.
// On success, it responds with HTTP 200 and the updated group.
// On validation error or internal error, it responds with HTTP 400 or 500 respectively.
func (h *GroupHandler) UpdateGroup(c *gin.Context) {
	id := c.Param("id")
	var group models.Group
	// Bind the JSON payload to the group model.
	if err := c.ShouldBindJSON(&group); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedGroup, err := h.keycloakService.UpdateGroup(id, group)
	if err != nil {
		log.Error().Err(err).Msg("Error updating group")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedGroup)
}

// DeleteGroup handles the HTTP DELETE request for deleting a group by ID.
// It expects the group ID as a path parameter.
// On success, it responds with HTTP 204 and no content.
// On error, it logs the error and responds with HTTP 500.
func (h *GroupHandler) DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	err := h.keycloakService.DeleteGroup(id)
	if err != nil {
		log.Error().Err(err).Msg("Error deleting group")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Respond with HTTP 204 No Content when deletion is successful.
	c.JSON(http.StatusNoContent, nil)
}

// SetKeycloakService overrides the underlying KeycloakService (useful for testing).
func (h *GroupHandler) SetKeycloakService(svc *services.KeycloakService) {
	h.keycloakService = svc
}
