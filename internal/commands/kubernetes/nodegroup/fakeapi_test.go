package nodegroup

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	upcloudv9 "github.com/UpCloudLtd/upcloud-go-api/v9/pkg/upcloud"
	"github.com/stretchr/testify/require"
)

type apiRequest struct {
	Method string
	Path   string
	Query  url.Values
	Body   []byte
}

type fakeResponse struct {
	status      int
	body        string
	contentType string
}

// fakeAPI serves canned responses keyed by "METHOD /path" and records every request.
type fakeAPI struct {
	mu       sync.Mutex
	requests []apiRequest
	client   *upcloudv9.ClientWithResponses
}

func newFakeAPI(t *testing.T, responses map[string]fakeResponse) *fakeAPI {
	t.Helper()
	api := &fakeAPI{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("reading request body: %v", err)
		}
		key := r.Method + " " + r.URL.Path
		api.mu.Lock()
		api.requests = append(api.requests, apiRequest{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query(), Body: body})
		api.mu.Unlock()

		res, ok := responses[key]
		if !ok {
			t.Errorf("unexpected request %s", key)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		contentType := res.contentType
		if contentType == "" {
			contentType = "application/json"
		}
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(res.status)
		_, _ = io.WriteString(w, res.body)
	}))
	t.Cleanup(server.Close)

	client, err := upcloudv9.NewClientWithResponses(server.URL)
	require.NoError(t, err)
	api.client = client
	return api
}

func (a *fakeAPI) all() []apiRequest {
	a.mu.Lock()
	defer a.mu.Unlock()
	return append([]apiRequest(nil), a.requests...)
}
