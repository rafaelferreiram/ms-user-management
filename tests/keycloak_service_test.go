package tests

import (
	"encoding/json"
	"io/ioutil"
	"ms-user/config"
	"ms-user/models"
	"ms-user/services"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// RoundTripFunc is a helper to override http.RoundTripper.
type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

// newTestKeycloakService returns a *services.KeycloakService configured to return dummy responses.
func newTestKeycloakService() *services.KeycloakService {
	// Create a test server that simulates all endpoints for the fake service.
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate token retrieval.
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/protocol/openid-connect/token") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "dummy-token"}`))
			return
		}
		// For users.
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/admin/realms/master/users") {
			// If searching by email.
			if r.URL.Query().Get("email") != "" {
				dummyUsers := []models.User{{ID: "1", Username: "user1", Email: r.URL.Query().Get("email")}}
				resp, _ := json.Marshal(dummyUsers)
				w.WriteHeader(http.StatusOK)
				w.Write(resp)
				return
			}
			// Otherwise, list users.
			dummyUsers := []models.User{{ID: "1", Username: "user1", Email: "user1@example.com"}}
			resp, _ := json.Marshal(dummyUsers)
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
			return
		}
		// For groups.
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/admin/realms/master/groups") {
			// If listing group members.
			if strings.HasSuffix(r.URL.Path, "/members") {
				dummyUsers := []models.User{{ID: "1", Username: "user1", Email: "user1@example.com"}}
				resp, _ := json.Marshal(dummyUsers)
				w.WriteHeader(http.StatusOK)
				w.Write(resp)
				return
			}
			dummyGroups := []models.Group{{ID: "1", Name: "Admins"}}
			resp, _ := json.Marshal(dummyGroups)
			w.WriteHeader(http.StatusOK)
			w.Write(resp)
			return
		}
		// For add, update, delete operations, simply return success.
		if r.Method == http.MethodPut || r.Method == http.MethodDelete || r.Method == http.MethodPost {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))

	cfg := &config.Config{
		KeycloakURL:      testServer.URL,
		KeycloakRealm:    "master",
		KeycloakUsername: "admin",
		KeycloakPassword: "admin",
	}
	ks := services.NewKeycloakService(cfg)
	ks.SetToken("dummy-token")
	ks.SetClient(newTestClientWithToken(testServer, nil))
	return ks
}

// newTestClientWithToken returns an HTTP client whose RoundTripper intercepts token requests.
// If the request is a POST to the token endpoint, it returns a dummy token;
// otherwise, it uses testServer.Client().Do(req) so that the original method (e.g. PUT) is preserved.
func newTestClientWithToken(testServer *httptest.Server, t *testing.T) *http.Client {
	return &http.Client{
		Transport: RoundTripFunc(func(req *http.Request) *http.Response {
			// Intercept token requests.
			if req.Method == http.MethodPost && strings.Contains(req.URL.Path, "/protocol/openid-connect/token") {
				tokenResponse := `{"access_token": "dummy-token"}`
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(strings.NewReader(tokenResponse)),
					Header:     make(http.Header),
				}
			}
			// For all other requests, use the test server's client preserving the method.
			res, err := testServer.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			body, _ := ioutil.ReadAll(res.Body)
			res.Body.Close()
			return &http.Response{
				StatusCode: res.StatusCode,
				Body:       ioutil.NopCloser(strings.NewReader(string(body))),
				Header:     res.Header,
			}
		}),
	}
}

// Test for ListUsers
func TestListUsers(t *testing.T) {
	dummyUsers := []models.User{
		{ID: "1", Username: "user1", Email: "user1@example.com"},
	}
	dummyResponse, _ := json.Marshal(dummyUsers)

	// Test server simulating token endpoint and ListUsers.
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Token endpoint.
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/protocol/openid-connect/token") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "dummy-token"}`))
			return
		}
		// ListUsers endpoint.
		w.WriteHeader(http.StatusOK)
		w.Write(dummyResponse)
	}))
	defer testServer.Close()

	cfg := &config.Config{
		KeycloakURL:      testServer.URL,
		KeycloakRealm:    "master",
		KeycloakUsername: "admin",
		KeycloakPassword: "admin",
	}
	kcService := services.NewKeycloakService(cfg)
	kcService.SetToken("dummy-token")
	kcService.SetClient(newTestClientWithToken(testServer, t))

	users, err := kcService.ListUsers()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 1 || users[0].Username != "user1" {
		t.Fatalf("unexpected users: %+v", users)
	}
}

// Test for SearchUserByEmail
func TestSearchUserByEmail(t *testing.T) {
	dummyUsers := []models.User{
		{ID: "1", Username: "user1", Email: "user1@example.com"},
	}
	dummyResponse, _ := json.Marshal(dummyUsers)

	// Test server simulating token endpoint and search.
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Token endpoint.
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/protocol/openid-connect/token") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "dummy-token"}`))
			return
		}
		// Search endpoint: expect GET to /admin/realms/master/users with query "email".
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/admin/realms/master/users") {
			if r.URL.Query().Get("email") == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write(dummyResponse)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer testServer.Close()

	cfg := &config.Config{
		KeycloakURL:      testServer.URL,
		KeycloakRealm:    "master",
		KeycloakUsername: "admin",
		KeycloakPassword: "admin",
	}
	kcService := services.NewKeycloakService(cfg)
	kcService.SetToken("dummy-token")
	kcService.SetClient(newTestClientWithToken(testServer, t))

	users, err := kcService.SearchUserByEmail("user1@example.com")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(users) != 1 || users[0].Email != "user1@example.com" {
		t.Fatalf("unexpected users: %+v", users)
	}
}

// Test for ListGroups
func TestListGroups(t *testing.T) {
	dummyGroups := []models.Group{
		{ID: "1", Name: "Admins"},
	}
	dummyResponse, _ := json.Marshal(dummyGroups)

	// Test server simulating token endpoint and ListGroups.
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Token endpoint.
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/protocol/openid-connect/token") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "dummy-token"}`))
			return
		}
		// ListGroups endpoint.
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/admin/realms/master/groups") {
			w.WriteHeader(http.StatusOK)
			w.Write(dummyResponse)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer testServer.Close()

	cfg := &config.Config{
		KeycloakURL:      testServer.URL,
		KeycloakRealm:    "master",
		KeycloakUsername: "admin",
		KeycloakPassword: "admin",
	}
	kcService := services.NewKeycloakService(cfg)
	kcService.SetToken("dummy-token")
	kcService.SetClient(newTestClientWithToken(testServer, t))

	groups, err := kcService.ListGroups()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(groups) != 1 || groups[0].Name != "Admins" {
		t.Fatalf("unexpected groups: %+v", groups)
	}
}

// Test for ListGroupsWithUsers
func TestListGroupsWithUsers(t *testing.T) {
	dummyGroups := []models.Group{
		{ID: "1", Name: "Admins"},
	}
	dummyUsers := []models.User{
		{ID: "10", Username: "user10", Email: "user10@example.com"},
	}
	groupResponse, _ := json.Marshal(dummyGroups)
	userResponse, _ := json.Marshal(dummyUsers)

	// Test server simulating token endpoint, group listing, and group members.
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Token endpoint.
		if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/protocol/openid-connect/token") {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"access_token": "dummy-token"}`))
			return
		}
		// If request is for group members.
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/members") {
			w.WriteHeader(http.StatusOK)
			w.Write(userResponse)
			return
		}
		// For group listing.
		if r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/admin/realms/master/groups") {
			w.WriteHeader(http.StatusOK)
			w.Write(groupResponse)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer testServer.Close()

	cfg := &config.Config{
		KeycloakURL:      testServer.URL,
		KeycloakRealm:    "master",
		KeycloakUsername: "admin",
		KeycloakPassword: "admin",
	}
	kcService := services.NewKeycloakService(cfg)
	kcService.SetToken("dummy-token")
	kcService.SetClient(newTestClientWithToken(testServer, t))

	result, err := kcService.ListGroupsWithUsers()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 group, got %d", len(result))
	}
	if result[0].Group.Name != "Admins" {
		t.Fatalf("unexpected group name: %s", result[0].Group.Name)
	}
	if len(result[0].Users) != 1 || result[0].Users[0].Username != "user10" {
		t.Fatalf("unexpected users: %+v", result[0].Users)
	}
}
