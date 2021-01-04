package proxy

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
)

func Test_mergeURLs(t *testing.T) {
	tests := []struct {
		name            string
		request, target *url.URL
		want            *url.URL
	}{{
		name:    "Empty request and target",
		request: &url.URL{},
		target:  &url.URL{},
		want:    &url.URL{Path: "/"},
	}, {
		name:    "With non empty hosts",
		request: &url.URL{Host: "localhost:80"},
		target:  &url.URL{Host: "localhost:8080"},
		want:    &url.URL{Host: "localhost:8080", Path: "/"},
	}, {
		name:    "With request path",
		request: &url.URL{Path: "/api"},
		target:  &url.URL{},
		want:    &url.URL{Path: "/api"},
	}, {
		name:    "With target path",
		request: &url.URL{},
		target:  &url.URL{Path: "/api"},
		want:    &url.URL{Path: "/api/"},
	}, {
		name:    "With request and target paths",
		request: &url.URL{Path: "/entities"},
		target:  &url.URL{Path: "/api"},
		want:    &url.URL{Path: "/api/entities"},
	}, {
		name:    "With request and target query parameters",
		request: &url.URL{RawQuery: "q=test"},
		target:  &url.URL{RawQuery: "q2=test2"},
		want:    &url.URL{Path: "/", RawQuery: "q2=test2&q=test"},
	}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeURLs(tt.request, tt.target); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("mergeURLs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_copyResponse(t *testing.T) {
	bodyContent := "This is a test text"
	rr := httptest.NewRecorder()
	proxyServerResponse := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"X-Test": []string{"value1", "value2"},
		},
		Body: ioutil.NopCloser(strings.NewReader(bodyContent)),
	}

	if err := copyResponse(proxyServerResponse, rr); err != nil {
		t.Errorf("Expected no error but got: %v", err)
		return
	}

	response := rr.Result()
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Errorf("Status code not expected want %v but got %v", http.StatusOK, response.StatusCode)
		return
	}

	header, ok := response.Header["X-Test"]
	if !ok {
		t.Errorf("Header X-Test not forwaded to response rr")
		return
	}

	if len(header) != len(response.Header["X-Test"]) {
		t.Errorf("X-Test header partially forwarded, got %v but expect %v", header, response.Header["X-Test"])
		return
	}

	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Errorf("Could not read response body: %v", err)
		return
	}

	if string(content) != bodyContent {
		t.Errorf("Response rr hasn't the same content than the response. got %v, expected %v", string(content), bodyContent)
		return
	}
}

func startTestServer() *httptest.Server {
	testMux := http.NewServeMux()

	testMux.HandleFunc("/api/data", func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
	})

	return httptest.NewServer(testMux)
}
