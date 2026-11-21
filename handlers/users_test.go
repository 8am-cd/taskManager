package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegisterUser(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		wantStatus int
		wantError  string
	}{
		{
			name:       "malformed JSON",
			body:       `{"email":`,
			wantStatus: http.StatusBadRequest,
			wantError:  "bad request",
		},
		{
			name:       "empty email",
			body:       `{"email":"  ","password":"123"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "email is empty",
		},
		{
			name:       "empty password",
			body:       `{"email":"test@example.com","password":"  \t\t"}`,
			wantStatus: http.StatusBadRequest,
			wantError:  "password is empty",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := NewHandler(nil)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tc.body))
			req.Header.Set("Content-Type", "application/json")
			res := httptest.NewRecorder()

			h.RegisterUser(res, req)

			if res.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body: %s", res.Code, tc.wantStatus, res.Body)
			}

			var body struct {
				Error string `json:"error"`
			}
			if err := json.Unmarshal(res.Body.Bytes(), &body); err != nil {
				t.Fatalf("invalid response JSON: %v; body: %s", err, res.Body)
			}
			if body.Error != tc.wantError {
				t.Errorf("error = %q, want %q", body.Error, tc.wantError)
			}
		})
	}
}
