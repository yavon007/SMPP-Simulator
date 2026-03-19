package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"smpp-simulator/pkg/jwt"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
		wantError  bool
	}{
		{
			name:       "valid credentials",
			body:       `{"username":"admin","password":"test123"}`,
			wantStatus: http.StatusOK,
			wantError:  false,
		},
		{
			name:       "invalid username",
			body:       `{"username":"wrong","password":"test123"}`,
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
		{
			name:       "invalid password",
			body:       `{"username":"admin","password":"wrong"}`,
			wantStatus: http.StatusUnauthorized,
			wantError:  true,
		},
		{
			name:       "missing username",
			body:       `{"password":"test123"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "missing password",
			body:       `{"username":"admin"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
			wantError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewAuthHandler("test123", "test-secret", 24)

			router := gin.New()
			router.POST("/login", handler.Login)

			req := httptest.NewRequest("POST", "/login", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Login() status = %v, want %v", w.Code, tt.wantStatus)
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if tt.wantError {
				if _, ok := response["error"]; !ok {
					t.Error("Login() expected error response")
				}
			} else {
				if _, ok := response["token"]; !ok {
					t.Error("Login() expected token in response")
				}
			}
		})
	}
}

func TestAuthHandler_Status(t *testing.T) {
	jwtSecret := "test-secret"
	handler := NewAuthHandler("test123", jwtSecret, 24)

	// Generate a valid token for testing
	validToken, _ := generateTestToken("admin", jwtSecret)

	tests := []struct {
		name           string
		authHeader     string
		wantAuth       bool
		wantUsername   string
	}{
		{
			name:       "no auth header",
			authHeader: "",
			wantAuth:   false,
		},
		{
			name:       "invalid auth format",
			authHeader: "InvalidFormat",
			wantAuth:   false,
		},
		{
			name:       "invalid token",
			authHeader: "Bearer invalid-token",
			wantAuth:   false,
		},
		{
			name:         "valid token",
			authHeader:   "Bearer " + validToken,
			wantAuth:     true,
			wantUsername: "admin",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.GET("/status", handler.Status)

			req := httptest.NewRequest("GET", "/status", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			var response StatusResponse
			json.Unmarshal(w.Body.Bytes(), &response)

			if response.Authenticated != tt.wantAuth {
				t.Errorf("Status() authenticated = %v, want %v", response.Authenticated, tt.wantAuth)
			}

			if tt.wantAuth && response.Username != tt.wantUsername {
				t.Errorf("Status() username = %v, want %v", response.Username, tt.wantUsername)
			}
		})
	}
}

// Helper function to generate test token
func generateTestToken(username, secret string) (string, error) {
	return jwt.GenerateToken(username, secret, 24)
}
