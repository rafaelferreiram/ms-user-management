# MS-USER Microservice

This microservice integrates with Keycloak's Admin API to provide CRUD operations for users and groups, along with endpoints to manage group membership. It is built in Go using the Gin framework, with logging via zerolog and a simple token-based authentication middleware.

## Features

- **User CRUD**: List, create, retrieve, update, and delete users in Keycloak.
- **Group CRUD**: List, create, retrieve, update, and delete groups in Keycloak.
- **Membership Management**:
  - List groups for a user.
  - Add and remove a user from a group (by user ID).
  - Search for users by email.
  - Add a user to a group by email.
  - List groups with their associated users.
- **Keycloak Integration**: Direct communication with Keycloak's REST Admin API.
- **Middleware**: Logging and simple token-based authentication.
- **Unit Tests**: Comprehensive unit tests are provided.
- **Dockerized**: Includes a Dockerfile for containerization.

## Prerequisites
- [Go 1.20+](https://golang.org/dl/)
- [Docker](https://docs.docker.com/get-docker/) (for containerization and running Keycloak)
- Keycloak running locally

## Running Keycloak Locally with Docker
You can run Keycloak locally using Docker. For example, execute the following command:

```bash
docker run -p 8080:8080 \
  -e KEYCLOAK_ADMIN=admin \
  -e KEYCLOAK_ADMIN_PASSWORD=admin \
  quay.io/keycloak/keycloak:20.0.1 start-dev
```
This command starts Keycloak on http://localhost:8080 with the admin credentials admin/admin.

## Setup and Installation
### Download Dependencies:
```bash
cd ms-user
go mod tidy
```

### Build the Application:
```bash
go build -o ms-user ./cmd
```
### Run the Application:
```bash
go run ./cmd/main.go
```
The service listens on port 18080 and exposes its endpoints under the base path /ms-user/v1.

## API Documentation with OpenAPI
The API is fully documented with an OpenAPI specification. The file `/ms-user/openapi.yaml` is included in the repository. You can import this file into an online editor such as [Swagger Editor](https://editor.swagger.io/) to interactively explore and test the API.
### To Import in Swagger Editor:
1. Open Swagger Editor.
2. Click on **File** → **Import File**.
3. Select the `openapi.yaml` file from your repository.
4. The editor will render the API documentation, and you can explore all endpoints interactively.

## API Endpoints
### Users
#### List Users
```bash
GET /ms-user/v1/users
#Description: List all users.
#Response: JSON array of user objects.
```
#### Create User
```bash
POST /ms-user/v1/users
#Description: Create a new user.
#Request Body: JSON object with user details (username, email, firstName, lastName).
#Response: The created user object.
```
#### Get User by Id
```bash
GET /ms-user/v1/users/{id}
#Description: Retrieve a user by ID.
#Response: JSON object with user details.
```
#### Update User
```bash
PUT /ms-user/v1/users/{id}
#Description: Update an existing user by ID.
#Request Body: JSON object with updated user details.
#Response: The updated user object.
```
#### Delete User
```bash
DELETE /ms-user/v1/users/{id}
#Description: Delete a user by ID.
```
#### Get User by Email
```bash
GET /ms-user/v1/users/search?email={email}
#Description: Search for users by email.
#Response: JSON array of user objects matching the email.
```
#### Add user to a group by email
```bash
PUT /ms-user/v1/users/email/{email}/groups/{groupId}
#Description: Add a user to a group by searching for the user via email.
#Note: The email must be URL-encoded (e.g., rafael%40example.com).
```

### Groups
#### List Groups
```bash
GET /ms-user/v1/groups
#Description: List all groups.
#Response: JSON array of group objects.
```
#### Create Group
```bash
POST /ms-user/v1/groups
#Description: Create a new group.
#Request Body: JSON object with group details (name).
#Response: The created group object.
```
#### Get Group by Id
```bash
GET /ms-user/v1/groups/{id}
#Description: Retrieve a group by ID.
#Response: JSON object with group details.
```
#### Update Group
```bash
PUT /ms-user/v1/groups/{id}
#Description: Update an existing group by ID.
#Request Body: JSON object with updated group details.
#Response: The updated group object.
```
#### Delete Group
```bash
DELETE /ms-user/v1/groups/{id}
#Description: Delete a group by ID.
```
#### List Groups with its users
```bash
GET /ms-user/v1/groups/with-users
#Description: List all groups along with the users that belong to each group.
#Response: JSON array where each object contains a group and an array of its users.
```
#### List Users from a Group Id
```bash
GET /ms-user/v1/groups/{id}/users
#Description: List all users in a specific group.
#Response: JSON array of user objects.
```

### Membership
#### List Groups that a User belongs
```bash
GET /ms-user/v1/users/{id}/groups
#Description: List all groups that a specific user belongs to.
```
#### Add User to a Groups by userId
```bash
PUT /ms-user/v1/users/{id}/groups/{groupId}
#Description: Add a user to a group using the user’s ID.
```
#### Remove User from Group
```bash
DELETE /ms-user/v1/users/{id}/groups/{groupId}
#Description: Remove a user from a group using the user’s ID.
```

## Running Tests
To run unit tests from the project root, execute:

```bash
go test ./...
```
This command will run tests in all packages.

## Docker
A Dockerfile is provided for containerization. To build and run the Docker image:

### Build the Docker Image:

```bash
docker build -t ms-user .
```
Run the Docker Container:

```bash
docker run -p 18080:18080 ms-user
```

## Authentication
The microservice uses a simple token-based authentication mechanism. For testing purposes, include the following header in your API requests:

```bash
Authorization: Bearer secret-token
```

## Postman
The postman collenction and enviroment can be found at \ms-user\postman\ms-user.postman_collection and \ms-user.postman_environment , importing into Postman you will be able to interact with the APIs once it is running locally.
