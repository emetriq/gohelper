package auth

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUserPasswordAuthSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		request := new(AuthRequest)
		b, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		assert.Nil(t, err)
		err = json.Unmarshal(b, request)
		assert.Nil(t, err)
		assert.Equal(t, "testuser", request.AuthParameters.Username)
		assert.Equal(t, "testpassword", request.AuthParameters.Password)
		assert.Equal(t, "1234", request.ClientId)
		assert.Equal(t, "USER_PASSWORD_AUTH", request.AuthFlow)
		assert.Equal(t, "ItjfzE2hxevxTggU0D25jlM5uAhlF+c2Lugh3S9JKHM=", request.AuthParameters.SecretHash)
		response := TokenResult{
			AuthenticationResult: AuthenticationResult{
				AccessToken:  "accessToken",
				IdToken:      "idToken",
				RefreshToken: "refreshToken",
				TokenType:    "tokenType",
				ExpiresIn:    3600,
			},
			ChallengeParameters: map[string]string{},
		}
		responseBody, _ := json.Marshal(response)
		w.Write(responseBody)
	}))

	user := "testuser"
	password := "testpassword"
	result, err := GetToken(user, password, server.URL, "1234", "abcdefghijklmnopqrstuvwxyz")

	assert.Nil(t, err)
	assert.NotNil(t, result)
}

func TestUserPasswordAuthFail(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("401 - Not Authorized"))
	}))

	user := "testuser"
	password := "testpassword"
	result, err := GetToken(user, password, server.URL, "1234", "abcdefghijklmnopqrstuvwxyz")

	assert.NotNil(t, err)
	assert.Nil(t, result)
}
