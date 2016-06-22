package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	conf "github.com/QubitProducts/bamboo/configuration"
	"github.com/QubitProducts/bamboo/services/service"
)

type testStorage struct {
	services         []service.Service
	err              error
	upsertChan       chan service.Service
	upsertResultChan chan error
	deleteChan       chan string
	deleteResultChan chan error
}

func newTestStorage(services []service.Service, err error) *testStorage {
	return &testStorage{
		services:         services,
		err:              err,
		upsertChan:       make(chan service.Service),
		upsertResultChan: make(chan error),
		deleteChan:       make(chan string),
		deleteResultChan: make(chan error),
	}
}

func (s *testStorage) All() ([]service.Service, error) {
	return s.services, s.err
}

func (s *testStorage) Upsert(service service.Service) error {
	s.upsertChan <- service
	return <-s.upsertResultChan
}

func (s *testStorage) Delete(serviceID string) error {
	s.deleteChan <- serviceID
	return <-s.deleteResultChan
}

func TestServiceAll(t *testing.T) {
	for _, test := range []struct {
		services []service.Service
		err      error
		status   int
		output   string
	}{
		{
			services: []service.Service{},
			err:      nil,
			status:   http.StatusOK,
			output:   "{}",
		},
		{
			services: []service.Service{},
			err:      errors.New("test error"),
			status:   http.StatusBadRequest,
			output:   "test error\n",
		},
		{
			services: []service.Service{
				service.Service{
					Id:     "/some/service",
					Acl:    "path_beg /some/service",
					Config: make(map[string]string),
				},
			},
			err:    nil,
			status: http.StatusOK,
			output: `{"/some/service":{"Id":"/some/service","Acl":"path_beg /some/service","Config":{}}}`,
		},
	} {
		c := &conf.Configuration{}
		store := newTestStorage(test.services, test.err)
		s := &ServiceAPI{
			Config:  c,
			Storage: store,
		}

		r, err := http.NewRequest("GET", "/api/services", nil)
		if err != nil {
			t.Fatalf("Error creating request: %s", err)
		}
		w := httptest.NewRecorder()

		s.All(w, r)

		if w.Code != test.status {
			t.Errorf("got %d, wanted %d", w.Code, test.status)
		}

		if w.Body.String() != test.output {
			t.Errorf("got '%s', wanted '%s'", w.Body.String(), test.output)
		}
	}
}

func TestServiceCreate(t *testing.T) {
	for _, test := range []struct {
		body     string
		expected *service.Service
		err      error
		status   int
		output   string
	}{
		{
			body:   "",
			status: http.StatusBadRequest,
			output: "Unable to decode JSON request\n",
		},
		{
			body:   `{}`,
			status: http.StatusBadRequest,
			output: "can not use empty ID\n",
		},
		{
			body: `{"Id":"/some/service","Acl":"path_beg /some/service"}`,
			expected: &service.Service{
				Id:     "/some/service",
				Acl:    "path_beg /some/service",
				Config: nil,
			},
			status: http.StatusOK,
			output: `{"Id":"/some/service","Acl":"path_beg /some/service","Config":null}`,
		},
		{
			body: `{"Id":"some/service","Acl":"path_beg /some/service"}`,
			expected: &service.Service{
				Id:     "/some/service",
				Acl:    "path_beg /some/service",
				Config: nil,
			},
			status: http.StatusOK,
			output: `{"Id":"/some/service","Acl":"path_beg /some/service","Config":null}`,
		},
		{
			body: `{"Id":"/some/service","Acl":"path_beg /some/service"}`,
			expected: &service.Service{
				Id:     "/some/service",
				Acl:    "path_beg /some/service",
				Config: nil,
			},
			err:    errors.New("test error"),
			status: http.StatusBadRequest,
			output: "test error\n",
		},
	} {
		c := &conf.Configuration{}
		store := newTestStorage([]service.Service{}, nil)
		s := &ServiceAPI{
			Config:  c,
			Storage: store,
		}

		r, err := http.NewRequest("POST", "/api/services", strings.NewReader(test.body))
		if err != nil {
			t.Fatalf("Error creating request: %s", err)
		}
		w := httptest.NewRecorder()

		go func() {
			service := <-store.upsertChan
			if !reflect.DeepEqual(service, *test.expected) {
				t.Errorf("got %#v, wanted %#v", service, test.expected)
			}
			store.upsertResultChan <- test.err
		}()

		s.Create(w, r)

		if w.Code != test.status {
			t.Errorf("got %d, wanted %d", w.Code, test.status)
		}

		if w.Body.String() != test.output {
			t.Errorf("got '%s', wanted '%s'", w.Body.String(), test.output)
		}
	}
}

func TestServicePut(t *testing.T) {
	for _, test := range []struct {
		body     string
		expected *service.Service
		err      error
		status   int
		output   string
	}{
		{
			body:   "",
			status: http.StatusBadRequest,
			output: "Unable to decode JSON request\n",
		},
		{
			body:   `{}`,
			status: http.StatusBadRequest,
			output: "can not use empty ID\n",
		},
		{
			body: `{"Id":"/some/service","Acl":"path_beg /some/service"}`,
			expected: &service.Service{
				Id:     "/some/service",
				Acl:    "path_beg /some/service",
				Config: nil,
			},
			status: http.StatusOK,
			output: `{"Id":"/some/service","Acl":"path_beg /some/service","Config":null}`,
		},
		{
			body: `{"Id":"some/service","Acl":"path_beg /some/service"}`,
			expected: &service.Service{
				Id:     "/some/service",
				Acl:    "path_beg /some/service",
				Config: nil,
			},
			status: http.StatusOK,
			output: `{"Id":"/some/service","Acl":"path_beg /some/service","Config":null}`,
		},
		{
			body: `{"Id":"/some/service","Acl":"path_beg /some/service"}`,
			expected: &service.Service{
				Id:     "/some/service",
				Acl:    "path_beg /some/service",
				Config: nil,
			},
			err:    errors.New("test error"),
			status: http.StatusBadRequest,
			output: "test error\n",
		},
	} {
		c := &conf.Configuration{}
		store := newTestStorage([]service.Service{}, nil)
		s := &ServiceAPI{
			Config:  c,
			Storage: store,
		}

		r, err := http.NewRequest("POST", "/api/services", strings.NewReader(test.body))
		if err != nil {
			t.Fatalf("Error creating request: %s", err)
		}
		w := httptest.NewRecorder()
		params := make(map[string]string)

		go func() {
			service := <-store.upsertChan
			if !reflect.DeepEqual(service, *test.expected) {
				t.Errorf("got %#v, wanted %#v", service, test.expected)
			}
			store.upsertResultChan <- test.err
		}()

		s.Put(params, w, r)

		if w.Code != test.status {
			t.Errorf("got %d, wanted %d", w.Code, test.status)
		}

		if w.Body.String() != test.output {
			t.Errorf("got '%s', wanted '%s'", w.Body.String(), test.output)
		}
	}
}

func TestServiceDelete(t *testing.T) {
	for _, test := range []struct {
		path     string
		expected string
		err      error
		status   int
		output   string
	}{
		{
			path:     "",
			expected: "",
			err:      nil,
			status:   http.StatusBadRequest,
			output:   "can not use empty ID\n",
		},
		{
			path:     "some/service",
			expected: "/some/service",
			err:      nil,
			status:   http.StatusOK,
			output:   "null",
		},
		{
			path:     "/some/service",
			expected: "/some/service",
			err:      nil,
			status:   http.StatusOK,
			output:   "null",
		},
		{
			path:     "/some/service",
			expected: "/some/service",
			err:      errors.New("test error"),
			status:   http.StatusBadRequest,
			output:   "test error\n",
		},
	} {
		c := &conf.Configuration{}
		store := newTestStorage([]service.Service{}, nil)
		s := &ServiceAPI{
			Config:  c,
			Storage: store,
		}

		r, err := http.NewRequest("POST", "/api/services/"+test.path, nil)
		if err != nil {
			t.Fatalf("Error creating request: %s", err)
		}
		w := httptest.NewRecorder()
		params := make(map[string]string)
		params["_1"] = test.path

		go func() {
			id := <-store.deleteChan
			if id != test.expected {
				t.Errorf("got '%s', wanted '%s'", id, test.expected)
			}
			store.deleteResultChan <- test.err
		}()

		s.Delete(params, w, r)

		if w.Code != test.status {
			t.Errorf("got %d, wanted %d", w.Code, test.status)
		}

		if w.Body.String() != test.output {
			t.Errorf("got '%s', wanted '%s'", w.Body.String(), test.output)
		}
	}
}
