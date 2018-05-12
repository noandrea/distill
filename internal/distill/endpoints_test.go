package distill

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRegisterEndpoints(t *testing.T) {

	tests := []struct {
		name    string
		route   string
		method  string
		hanlder func(w http.ResponseWriter, r *http.Request)
	}{
		{
			"hc",
			"/health-check",
			"GET",
			healthCheckHanlder,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req, err := http.NewRequest(tt.method, tt.route, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler := http.HandlerFunc(tt.hanlder)

			// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
			// directly and pass in our Request and ResponseRecorder.
			handler.ServeHTTP(rr, req)

			// Check the status code is what we expect.
			if status := rr.Code; status != http.StatusOK {
				t.Errorf("handler returned wrong status code: got %v want %v",
					status, http.StatusOK)
			}
		})
	}
}
