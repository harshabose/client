package client

import (
	"encoding/json"
	"os"
	"strings"

	"google.golang.org/api/option"
)

type firebaseConfig struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
	UniverseDomain          string `json:"universe_domain"`
}

func GetFirebaseConfiguration() (option.ClientOption, error) {
	config := firebaseConfig{
		Type:                    os.Getenv("FIREBASE_TYPE"),
		ProjectID:               os.Getenv("FIREBASE_PROJECT_ID"),
		PrivateKeyID:            os.Getenv("FIREBASE_PRIVATE_KEY_ID"),
		PrivateKey:              strings.ReplaceAll(os.Getenv("FIREBASE_PRIVATE_KEY"), "\\n", "\n"),
		ClientEmail:             os.Getenv("FIREBASE_CLIENT_EMAIL"),
		ClientID:                os.Getenv("FIREBASE_CLIENT_ID"),
		AuthURI:                 os.Getenv("FIREBASE_AUTH_URI"),
		TokenURI:                os.Getenv("FIREBASE_AUTH_TOKEN_URI"),
		AuthProviderX509CertURL: os.Getenv("FIREBASE_AUTH_PROVIDER_X509_CERT_URL"),
		ClientX509CertURL:       os.Getenv("FIREBASE_AUTH_CLIENT_X509_CERT_URL"),
		UniverseDomain:          os.Getenv("FIREBASE_UNIVERSE_DOMAIN"),
	}

	configBytes, err := json.Marshal(config)
	if err != nil {
		return nil, err
	}
	return option.WithCredentialsJSON(configBytes), nil
}
