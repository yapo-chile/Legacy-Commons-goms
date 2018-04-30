package handlers

import (
	"github.com/Yapo/goutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type MockHandler struct {
	mock.Mock
}

func (m *MockHandler) Input() HandlerInput {
	args := m.Called()
	return args.Get(0).(HandlerInput)
}

func (m *MockHandler) Execute(getter InputGetter) *goutils.Response {
	args := m.Called(getter)
	_, response := getter()
	if response != nil {
		return response
	}
	return args.Get(0).(*goutils.Response)
}

type DummyInput struct {
	X int
}

func TestJsonJandlerFuncOK(t *testing.T) {
	h := MockHandler{}
	input := &DummyInput{}
	response := &goutils.Response{
		Code: 42,
		Body: goutils.GenericError{"That's some bad hat, Harry"},
	}
	getter := mock.AnythingOfType("handlers.InputGetter")
	h.On("Execute", getter).Return(response).Once()
	h.On("Input").Return(input).Once()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/someurl", strings.NewReader("{}"))
	fn := MakeJSONHandlerFunc(&h)
	fn(w, r)

	assert.Equal(t, 42, w.Code)
	assert.Equal(t, `{"ErrorMessage":"That's some bad hat, Harry"}`+"\n", w.Body.String())
	h.AssertExpectations(t)
}

func TestJsonJandlerFuncParseError(t *testing.T) {
	h := MockHandler{}
	input := &DummyInput{}
	getter := mock.AnythingOfType("handlers.InputGetter")
	h.On("Execute", getter)
	h.On("Input").Return(input).Once()

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/someurl", strings.NewReader("{"))
	fn := MakeJSONHandlerFunc(&h)
	fn(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, `{"ErrorMessage":"unexpected EOF"}`+"\n", w.Body.String())
	h.AssertExpectations(t)
}
