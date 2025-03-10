package handlers

import (
	"ms-user/config"
	"ms-user/models"
	"ms-user/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// UserHandler handles HTTP requests related to user management.
// It utilizes the KeycloakService to perform CRUD operations on users through Keycloak's Admin API.
type UserHandler struct {
	keycloakService *services.KeycloakService
}

// NewUserHandler initializes and returns a new UserHandler instance.
// It sets up a new KeycloakService using the provided configuration.
func NewUserHandler(cfg *config.Config) *UserHandler {
	return &UserHandler{
		keycloakService: services.NewKeycloakService(cfg),
	}
}

// ListUsers handles the HTTP GET request for retrieving all users.
// Endpoint: GET /users
//
// Input: No body parameters. The request may include headers (e.g., for authentication).
// Output: On success, returns HTTP 200 with a JSON array of user objects.
//
//	On error, returns HTTP 500 with an error message.
func (h *UserHandler) ListUsers(c *gin.Context) {
	users, err := h.keycloakService.ListUsers()
	if err != nil {
		log.Error().Err(err).Msg("Error listing users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// CreateUser handles the HTTP POST request for creating a new user.
// Endpoint: POST /users
//
// Input: A JSON body representing the user to be created (models.User).
// Output: On success, returns HTTP 201 with the created user object.
//
//	On error (e.g., validation issues or internal errors), returns HTTP 400 or 500 with an error message.
func (h *UserHandler) CreateUser(c *gin.Context) {
	var user models.User
	// Bind the incoming JSON payload to the user model.
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	createdUser, err := h.keycloakService.CreateUser(user)
	if err != nil {
		log.Error().Err(err).Msg("Error creating user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, createdUser)
}

// GetUser handles the HTTP GET request for retrieving a specific user by ID.
// Endpoint: GET /users/:id
//
// Input: The user ID is provided as a URL path parameter.
// Output: On success, returns HTTP 200 with the user object.
//
//	On error (e.g., user not found), returns HTTP 404 with an error message.
func (h *UserHandler) GetUser(c *gin.Context) {
	id := c.Param("id")
	user, err := h.keycloakService.GetUser(id)
	if err != nil {
		log.Error().Err(err).Msg("Error fetching user")
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// SearchUserByEmail handles the HTTP GET request to search for users by email.
// Endpoint: GET /ms-user/v1/users/search?email=<email>
// Input: Query parameter "email".
// Output: On success, returns HTTP 200 with a JSON array of matching users.
//
//	On error, returns an appropriate HTTP status with an error message.
func (h *UserHandler) SearchUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email query parameter is required"})
		return
	}

	users, err := h.keycloakService.SearchUserByEmail(email)
	if err != nil {
		log.Error().Err(err).Msg("Error searching user by email")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, users)
}

// UpdateUser handles the HTTP PUT request for updating an existing user.
// Endpoint: PUT /users/:id
//
// Input: The user ID is provided as a URL path parameter, and the request body contains the updated user data in JSON format.
// Output: On success, returns HTTP 200 with the updated user object.
//
//	On error, returns HTTP 400 for invalid input or HTTP 500 for internal errors.
func (h *UserHandler) UpdateUser(c *gin.Context) {
	id := c.Param("id")
	var user models.User
	// Bind the JSON payload to the user model.
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	updatedUser, err := h.keycloakService.UpdateUser(id, user)
	if err != nil {
		log.Error().Err(err).Msg("Error updating user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updatedUser)
}

// DeleteUser handles the HTTP DELETE request for removing a user by ID.
// Endpoint: DELETE /users/:id
//
// Input: The user ID is provided as a URL path parameter.
// Output: On success, returns HTTP 204 with no content.
//
//	On error, returns HTTP 500 with an error message.
func (h *UserHandler) DeleteUser(c *gin.Context) {
	id := c.Param("id")
	err := h.keycloakService.DeleteUser(id)
	if err != nil {
		log.Error().Err(err).Msg("Error deleting user")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// SetKeycloakService overrides the underlying KeycloakService (useful for testing).
func (h *UserHandler) SetKeycloakService(svc *services.KeycloakService) {
	h.keycloakService = svc
}
