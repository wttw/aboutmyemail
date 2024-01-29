package aboutmyemail

import (
	"context"
	"net/http"
	"os"
)

const apiEndpoint = "https://api.aboutmy.email/api/v1"

const envMyemailServer = "MYEMAIL_SERVER"
const envMyemailApikey = "MYEMAIL_APIKEY"

//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest --config=api-model.cfg.yaml ameapi.yaml
//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen@latest --config=api-client.cfg.yaml ameapi.yaml

// New creates a new ClientWithResponses for AboutMy.email, with reasonable defaults
func New(opts ...ClientOption) (*ClientWithResponses, error) {
	if server := os.Getenv(envMyemailServer); server != "" {
		opts = append([]ClientOption{WithServer(server)}, opts...)
	}
	if apikey := os.Getenv(envMyemailApikey); apikey != "" {
		opts = append([]ClientOption{WithApiKey(apikey)}, opts...)
	}
	return NewClientWithResponses(apiEndpoint, opts...)
}

func doNothing(*Client) error {
	return nil
}

// WithApiKey sets the client Authorization header with this key, if key isn't empty
func WithApiKey(key string) ClientOption {
	if key == "" {
		return doNothing
	}
	authData := "Bearer " + key
	return WithRequestEditorFn(func(ctx context.Context, r *http.Request) error {
		r.Header.Set("Authorization", authData)
		return nil
	})
}

// WithServer sets the client API endpoint, if endpoint isn't empty
func WithServer(endpoint string) ClientOption {
	if endpoint == "" {
		return doNothing
	}
	return WithBaseURL(endpoint)
}
