package infrastructure

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNewrelicStartError(t *testing.T) {
	nr := NewRelicHandler{
		Appname: "Test",
		Key:     "NotAValidKey",
	}
	err := nr.Start()
	assert.Error(t, err)
}

type MockHandler struct {
	mock.Mock
}

func (m *MockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.Called(r, w)
	w.Write([]byte("been there"))
}

func TestNewrelicStartOk(t *testing.T) {
	var conf NewRelicConf
	LoadFromEnv(&conf)
	nr := NewRelicHandler{
		Appname: conf.Appname,
		Key:     conf.Key,
	}
	err := nr.Start()
	assert.NoError(t, err)

	m := MockHandler{}
	m.On("ServeHTTP",
		mock.AnythingOfType("*http.Request"),
		mock.AnythingOfType("newrelic.wrapF")).Return()

	handler := nr.TrackHandler("test", &m)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/someurl", strings.NewReader("{}"))
	handler.ServeHTTP(w, r)

	assert.Equal(t, "been there", w.Body.String())
	m.AssertExpectations(t)
}
