package auth

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/emetriq/gohelper/security/hash/sha256"
)

type AuthenticationResult struct {
	AccessToken  string `json:"AccessToken"`
	IdToken      string `json:"IdToken"`
	RefreshToken string `json:"RefreshToken"`
	TokenType    string `json:"TokenType"`
	ExpiresIn    int    `json:"ExpiresIn"`
}

type TokenResult struct {
	AuthenticationResult AuthenticationResult `json:"AuthenticationResult"`
	ChallengeParameters  map[string]string    `json:"ChallengeParameters"`
}

type AuthParameters struct {
	Username   string `json:"USERNAME"`
	Password   string `json:"PASSWORD"`
	SecretHash string `json:"SECRET_HASH"`
}

type AuthRequest struct {
	AuthParameters AuthParameters `json:"AuthParameters"`
	AuthFlow       string         `json:"AuthFlow"`
	ClientId       string         `json:"ClientId"`
}

func GetToken(user, pw, cognitoAuthURL, cognitoClientID, cognitoClientSecret string) (*TokenResult, error) {
	authRequest := AuthRequest{
		AuthParameters: AuthParameters{
			Username:   user,
			Password:   pw,
			SecretHash: sha256.Base64EncSha256([]byte(user+cognitoClientID), []byte(cognitoClientSecret)),
		},
		AuthFlow: "USER_PASSWORD_AUTH",
		ClientId: cognitoClientID,
	}
	postBody, _ := json.Marshal(authRequest)
	responseBody := bytes.NewBuffer(postBody)
	req, _ := http.NewRequest(http.MethodPost, cognitoAuthURL, responseBody)
	req.Header.Add("X-Amz-Target", "AWSCognitoIdentityProviderService.InitiateAuth")
	req.Header.Add("Content-Type", "application/x-amz-json-1.1")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	result := new(TokenResult)
	if err := json.Unmarshal(body, result); err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return result, nil
}
