package main

import (
	"ms-user/config"
	"ms-user/handlers"
	"ms-user/middleware"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// main is the entry point of the ms-user microservice.
// It loads configuration, sets up middleware and API routes,
// and starts the HTTP server on the specified port.
func main() {
	// Load configuration from environment variables or defaults.
	cfg := config.LoadConfig()

	// Create a new Gin router instance.
	r := gin.New()

	// Register global middleware.
	// LoggingMiddleware logs each incoming request.
	// AuthMiddleware enforces a simple token-based authentication.
	r.Use(middleware.LoggingMiddleware())
	r.Use(middleware.AuthMiddleware())

	// Initialize handler instances for user, group, and membership operations.
	// Handlers interact with Keycloak via the service layer.
	userHandler := handlers.NewUserHandler(cfg)
	groupHandler := handlers.NewGroupHandler(cfg)
	membershipHandler := handlers.NewMembershipHandler(cfg)

	// Register User-related routes under the base path "ms-user/v1/users".
	// These endpoints handle user CRUD operations and membership management.
	userRoutes := r.Group("ms-user/v1/users")
	{
		// GET /ms-user/v1/users - List all users.
		userRoutes.GET("", userHandler.ListUsers)
		// Search user by email: GET /ms-user/v1/users/search?email=<email>
		userRoutes.GET("/search", userHandler.SearchUserByEmail)
		// POST /ms-user/v1/users - Create a new user.
		userRoutes.POST("", userHandler.CreateUser)
		// GET /ms-user/v1/users/:id - Retrieve a specific user by ID.
		userRoutes.GET("/:id", userHandler.GetUser)
		// PUT /ms-user/v1/users/:id - Update an existing user by ID.
		userRoutes.PUT("/:id", userHandler.UpdateUser)
		// DELETE /ms-user/v1/users/:id - Delete a user by ID.
		userRoutes.DELETE("/:id", userHandler.DeleteUser)

		// Membership endpoints for users:
		// GET /ms-user/v1/users/:id/groups - List groups for a specific user.
		userRoutes.GET("/:id/groups", membershipHandler.ListUserGroups)
		// Add user to group by email: PUT /ms-user/v1/users/email/:email/groups/:groupId
		userRoutes.PUT("/email/:email/groups/:groupId", membershipHandler.AddUserToGroupByEmail)
		// PUT /ms-user/v1/users/:id/groups/:groupId - Add a user to a group.
		userRoutes.PUT("/:id/groups/:groupId", membershipHandler.AddUserToGroup)
		// DELETE /ms-user/v1/users/:id/groups/:groupId - Remove a user from a group.
		userRoutes.DELETE("/:id/groups/:groupId", membershipHandler.RemoveUserFromGroup)

	}

	// Register Group-related routes under the base path "ms-user/v1/groups".
	// These endpoints handle group CRUD operations and listing users within a group.
	groupRoutes := r.Group("ms-user/v1/groups")
	{
		// GET /ms-user/v1/groups - List all groups.
		groupRoutes.GET("", groupHandler.ListGroups)
		// POST /ms-user/v1/groups - Create a new group.
		groupRoutes.POST("", groupHandler.CreateGroup)
		// GET /ms-user/v1/groups/:id - Retrieve a specific group by ID.
		groupRoutes.GET("/:id", groupHandler.GetGroup)
		// PUT /ms-user/v1/groups/:id - Update an existing group by ID.
		groupRoutes.PUT("/:id", groupHandler.UpdateGroup)
		// DELETE /ms-user/v1/groups/:id - Delete a group by ID.
		groupRoutes.DELETE("/:id", groupHandler.DeleteGroup)

		// Membership endpoint for groups:
		// GET /ms-user/v1/groups/:id/users - List all users in a specific group.
		groupRoutes.GET("/:id/users", membershipHandler.ListGroupUsers)

		// New endpoint: List groups with their associated users.
		groupRoutes.GET("/with-users", groupHandler.ListGroupsWithUsers)
	}

	// Log the startup information and start the HTTP server on port 18080.
	log.Info().Msg("Starting ms-user service on port 18080")
	if err := r.Run(":18080"); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
