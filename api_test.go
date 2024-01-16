package aboutmyemail

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

const testServer = `http://127.0.0.1:3000/api/v1`
const testKey = `myemail_0LekAwu5Wob2kSru`

const sampleEmail = `From: <steve@blighty.com>
To: <steve@blighty.com>
Subject: test sausage

body
`

func TestApi(t *testing.T) {
	client, err := New(WithServer(testServer), WithApiKey(testKey))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	//ctx = context.Background()
	token := "potato"

	result, err := client.EmailWithResponse(ctx, EmailJSONRequestBody{
		From:    "steve@blighty.com",
		Ip:      "10.11.12.13",
		Payload: sampleEmail,
		To:      "steve@blighty.com",
		Token:   &token,
	})
	if err != nil {
		t.Errorf("client.Email() failed: %v", err)
		return
	}

	if result.StatusCode() != 200 {
		t.Logf("result=%v", string(result.Body))
		t.Errorf("expected 200 response to Email, got %d %s", result.StatusCode(), result.Status())
		return
	}

	if result.JSON200 == nil {
		t.Errorf("expected non-nil success response")
		return
	}
	id := result.JSON200.Id
	for {
		result, err := client.EmailStatusWithResponse(ctx, id)
		if err != nil {
			t.Errorf("client.EmailStatus() failed: %v", err)
			return
		}
		if result.StatusCode() != 200 {
			t.Logf("result=%v", string(result.Body))
			t.Errorf("expected 200 response to EmailStatus, got %d %s", result.StatusCode(), result.Status())
			return
		}
		if result.JSON200 == nil {
			t.Errorf("expected non-nil success response")
			return
		}
		if result.JSON200.Url != nil {
			want := fmt.Sprintf(strings.TrimSuffix(testServer, "api/v1") + id)
			got := *result.JSON200.Url
			if want != got {
				t.Errorf("url want '%s', got '%s'", want, got)
			}
			if result.JSON200.Id != id {
				t.Errorf("id want '%s', got '%s'", id, result.JSON200.Id)
			}
			if result.JSON200.Token == nil {
				t.Errorf("token want non-nil, got nil")
			} else {
				got := *result.JSON200.Token
				want := token
				if want != got {
					t.Errorf("token want '%s', got '%s'", want, got)
				}
			}
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
}

func TestApi_Callbacks(t *testing.T) {
	client, err := New(WithServer(testServer), WithApiKey(testKey))
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	//ctx = context.Background()
	token := "banana"

	finishedChan := make(chan struct{})
	var callbackBody []byte
	messageCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "message") {
			messageCount++
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if strings.HasSuffix(r.URL.Path, "finished") {
			callbackBody, err = io.ReadAll(r.Body)
			if err != nil {
				panic(fmt.Errorf("failed to read callback body: %w", err))
			}
			w.WriteHeader(http.StatusNoContent)
			close(finishedChan)
			return
		}
		panic(fmt.Errorf("unrecognised query to callback server: %s", r.URL.Path))
	}))
	defer ts.Close()

	finishedUrl := ts.URL + "/finished"
	messageUrl := ts.URL + "/message"
	result, err := client.EmailWithResponse(ctx, EmailJSONRequestBody{
		From:        "steve@blighty.com",
		Ip:          "10.11.12.13",
		Payload:     sampleEmail,
		To:          "steve@blighty.com",
		Token:       &token,
		FinishedUrl: &finishedUrl,
		ProgressUrl: &messageUrl,
	})
	if err != nil {
		t.Errorf("client.Email() failed: %v", err)
		return
	}

	select {
	case <-ctx.Done():
		t.Errorf("context timeout")
		return
	case <-finishedChan:
	}

	var finResult StatusResult
	err = json.Unmarshal(callbackBody, &finResult)
	if err != nil {
		t.Logf("callback body:\n%s\n", string(callbackBody))
		t.Errorf("failed to unmarshal callback result: %s", err)
	}
	if finResult.Url == nil {
		t.Errorf("Wanted non-nil result url, got nil")
	} else {
		wantUrl := fmt.Sprintf(strings.TrimSuffix(testServer, "api/v1") + result.JSON200.Id)
		gotUrl := *finResult.Url
		if wantUrl != gotUrl {
			t.Errorf("Finished URL want '%s', got '%s'", wantUrl, gotUrl)
		}
	}
	if finResult.Token == nil {
		t.Errorf("Wanted non-nil result token, got nil")
	} else {
		got := *finResult.Token
		if got != token {
			t.Errorf("Token want '%s', got '%s", token, got)
		}
	}
}
