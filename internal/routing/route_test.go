package routing

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"reflect"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttputil"

	"github.com/krylphi/helloworld-data-handler/internal/domain"
	"github.com/krylphi/helloworld-data-handler/internal/mock"
	"github.com/krylphi/helloworld-data-handler/internal/utils"
)

type testDefinition struct {
	name               string
	method             string
	endpoint           string
	body               []byte
	wantStatusCode     int
	responseValidation func(t *testing.T, tt testDefinition, response *http.Response, expected interface{})
	expectErr          error
	expectedCallResult interface{}
	expectedResponse   interface{}
}

func initRouting(t *testing.T) (Router, *mock.MockHandler) {
	mockCtrl := gomock.NewController(t)
	handler := mock.NewMockHandler(mockCtrl)
	return NewRouter(handler), handler
}

func serve(handler fasthttp.RequestHandler, req *http.Request) (*http.Response, error) {
	ln := fasthttputil.NewInmemoryListener()
	defer ln.Close()

	go func() {
		err := fasthttp.Serve(ln, handler)
		if err != nil {
			panic(fmt.Errorf("failed to serve: %v", err))
		}
	}()

	client := http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return ln.Dial()
			},
		},
	}

	return client.Do(req)
}

func performHTTPEndpointsTest(t *testing.T, tt testDefinition, router Router, calls []*gomock.Call) {
	var handler fasthttp.RequestHandler = router.HTTPRouter
	url := utils.Concat("http://0.0.0.0:8080", tt.endpoint)
	request, err := http.NewRequest(tt.method, url, bytes.NewBuffer(tt.body))
	if err != nil {
		t.Fatalf("error, while request initialization: %v", err)
	}
	for _, call := range calls {
		call.AnyTimes()
	}
	result, err := serve(handler, request)
	if err != nil {
		t.Fatalf("%v error, while performing request: %v", tt.name, err.Error())
	}
	defer func() {
		if result != nil {
			err := result.Body.Close()
			if err != nil {
				t.Fatalf("%v error, while closing request body: %v", tt.name, err.Error())
			}
		}
	}()

	if result.StatusCode != tt.wantStatusCode {
		t.Fatalf("%v; code=%v, expected %v", tt.name, result.StatusCode, tt.wantStatusCode)
	}

	if tt.responseValidation != nil {
		tt.responseValidation(t, tt, result, tt.expectedResponse)
	}
}

func TestRouter_Shutdown(t *testing.T) {
	router, handler := initRouting(t)
	t.Run("Shutdown call", func(t *testing.T) {
		handler.EXPECT().Flush().Times(1)
		router.Shutdown()
	})
}

func TestRouter_handleGetLog(t *testing.T) {
	tests := []testDefinition{
		{
			name:           "OK",
			method:         "GET",
			endpoint:       "/log",
			body:           nil,
			wantStatusCode: fasthttp.StatusOK,
			responseValidation: func(t *testing.T, tt testDefinition, response *http.Response, expected interface{}) {
				b, err := ioutil.ReadAll(response.Body)
				if err != nil {
					t.Fatalf("%v error, while reading request body: %v", tt.name, err.Error())
				}
				if !reflect.DeepEqual(string(b), expected) {
					t.Errorf("%v = %v, want %v", tt.name, string(b), expected)
				}
				defer func() {
					if response != nil {
						err := response.Body.Close()
						if err != nil {
							t.Fatalf("%v error, while closing request body: %v", tt.name, err.Error())
						}
					}
				}()
			},
			expectErr:          nil,
			expectedCallResult: nil,
			expectedResponse:   "OK",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			router, _ := initRouting(t)
			var calls []*gomock.Call
			performHTTPEndpointsTest(t, tt, router, calls)
		})
	}
}

func TestRouter_handleLog(t *testing.T) {
	tests := []testDefinition{
		{
			name:           "PUT",
			method:         "PUT",
			endpoint:       "/log",
			wantStatusCode: fasthttp.StatusMethodNotAllowed,
		},
		{
			name:           "DELETE",
			method:         "DELETE",
			endpoint:       "/log",
			wantStatusCode: fasthttp.StatusMethodNotAllowed,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			router, _ := initRouting(t)
			var calls []*gomock.Call
			performHTTPEndpointsTest(t, tt, router, calls)
		})
	}
}

func TestRouter_handlePostLog(t *testing.T) {
	tests := []testDefinition{
		{
			name:           "OK",
			method:         "POST",
			endpoint:       "/log",
			body:           []byte("{\"text\":\"hello world\",\"content_id\":1,\"client_id\":1,\"timestamp\":1586846680064}"),
			wantStatusCode: fasthttp.StatusOK,
			responseValidation: func(t *testing.T, tt testDefinition, response *http.Response, expected interface{}) {
				b, err := ioutil.ReadAll(response.Body)
				if err != nil {
					t.Fatalf("%v error, while reading request body: %v", tt.name, err.Error())
				}
				if !reflect.DeepEqual(string(b), expected) {
					t.Errorf("%v = %v, want %v", tt.name, string(b), expected)
				}
				defer func() {
					if response != nil {
						err := response.Body.Close()
						if err != nil {
							t.Fatalf("%v error, while closing request body: %v", tt.name, err.Error())
						}
					}
				}()
			},
			expectErr:          nil,
			expectedCallResult: nil,
			expectedResponse:   "OK",
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			router, handler := initRouting(t)
			calls := []*gomock.Call{
				handler.EXPECT().Send(&domain.Entry{
					ContentID: 1,
					Timestamp: 1586846680064,
					ClientID:  1,
					Text:      "hello world",
				}).Return(tt.expectedCallResult),
			}
			performHTTPEndpointsTest(t, tt, router, calls)
		})
	}
}
