package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/peeley/carpal/internal/config"
	"github.com/peeley/carpal/internal/driver/file"
)

func TestResourceHandlerFileDriver(t *testing.T) {
	config := config.Configuration{
		Driver: "file",
		FileConfiguration: &config.FileConfiguration{
			Directory: "../../test",
		},
	}

	fileDriver := file.NewFileDriver(config)

	handler := NewResourceHandler(fileDriver)
	httpHandler := http.HandlerFunc(handler.Handle)

	t.Run("can retrieve resources and serve via http", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		query := req.URL.Query()
		query.Add("resource", "acct:bob@foobar.com")
		req.URL.RawQuery = query.Encode()

		responseRecorder := httptest.NewRecorder()
		httpHandler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != http.StatusOK {
			t.Fatalf("expected 200 OK, got %v", responseRecorder.Code)
		}

		responseHeaders := responseRecorder.HeaderMap

		if len(responseHeaders["Content-Type"]) != 1 ||
			responseHeaders["Content-Type"][0] != "application/jrd+json" {
			t.Fatalf(
				"expected application/jrd+json content type, got %v",
				responseHeaders["Content-Type"],
			)
		}

		body := responseRecorder.Body.String()

		want := `{"subject":"acct:bob@foobar.com","aliases":["mailto:bob@foobar.com","https://mastodon/bob"],"properties":{"http://webfinger.example/ns/name":"Bob Smith"},"links":[{"rel":"http://webfinger.example/rel/profile-page","href":"https://www.example.com/~bob/"},{"rel":"http://webfinger.example/rel/businesscard","href":"https://www.example.com/~bob/bob.vcf"}]}`

		if body != want {
			t.Fatalf("got: %+v,\n want: %+v", body, want)
		}
	})

	t.Run("nonexistent resources return 404", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)
		query := req.URL.Query()
		query.Add("resource", "missingno")
		req.URL.RawQuery = query.Encode()

		responseRecorder := httptest.NewRecorder()
		httpHandler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != http.StatusNotFound {
			t.Fatalf(
				"expected 404, got %v, `%v`",
				responseRecorder.Code,
				responseRecorder.Body.String(),
			)
		}
	})

	t.Run("requests not including `resource` fail return 400", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/", nil)

		responseRecorder := httptest.NewRecorder()
		httpHandler.ServeHTTP(responseRecorder, req)

		if responseRecorder.Code != http.StatusBadRequest {
			t.Fatalf(
				"expected 400, got %v, `%v`",
				responseRecorder.Code,
				responseRecorder.Body.String(),
			)
		}
	})

	t.Run("non-GET requests return 405", func(t *testing.T) {
		reqMethods := []string{
			http.MethodConnect,
			http.MethodDelete,
			http.MethodHead,
			http.MethodOptions,
			http.MethodPatch,
			http.MethodPost,
			http.MethodPut,
			http.MethodTrace,
		}

		for _, method := range reqMethods {
			req, _ := http.NewRequest(method, "/", nil)

			responseRecorder := httptest.NewRecorder()
			httpHandler.ServeHTTP(responseRecorder, req)

			if responseRecorder.Code != http.StatusMethodNotAllowed {
				t.Fatalf(
					"expected 405, got %v, `%v`",
					responseRecorder.Code,
					responseRecorder.Body.String(),
				)
			}
		}
	})
}
