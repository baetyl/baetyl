package mock

import (
	"net/http"
	"net/http/httptest"
)

type response struct {
	status int
	body   []byte
}

type mockServer struct {
	*httptest.Server
	responses []response
}

func newMockServer(responses []response) *mockServer {
	ms := &mockServer{responses: responses}
	ms.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if len(ms.responses) == 0 {
			return
		}
		w.WriteHeader(ms.responses[0].status)
		w.Write(ms.responses[0].body)
		ms.responses = ms.responses[1:]
	}))
	return ms
}
