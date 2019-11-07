package usecases

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserProfileRepository struct {
	mock.Mock
}

func (m *MockUserProfileRepository) GetUserProfileData(mail string) (UserBasicData, error) {
	args := m.Called(mail)
	return args.Get(0).(UserBasicData), args.Error(1)
}

type mockGetUserPrometheusDefaultLogger struct {
	mock.Mock
}

func (m *mockGetUserPrometheusDefaultLogger) LogBadInput(mail string) {
	m.Called(mail)
}

func TestGetUserOk(t *testing.T) {
	m := MockUserProfileRepository{}
	var userb UserBasicData
	m.On("GetUserProfileData", "").Return(userb, nil)

	i := GetUserDataInteractor{
		UserProfileRepository: &m,
	}
	expected := UserBasicData{"", "", "", "", "", ""}
	output, err := i.GetUser("")
	assert.NoError(t, err)
	assert.Equal(t, expected, output)
	m.AssertExpectations(t)

}
func TestGetUserError(t *testing.T) {
	m := MockUserProfileRepository{}
	mLogger := &mockGetUserPrometheusDefaultLogger{}
	var userb UserBasicData

	mLogger.On("LogBadInput", "")
	m.On("GetUserProfileData", "").Return(userb, fmt.Errorf("error"))

	i := GetUserDataInteractor{
		UserProfileRepository: &m,
		Logger:                mLogger,
	}

	output, err := i.GetUser("")
	assert.Error(t, err)
	assert.Empty(t, output)
	mLogger.AssertExpectations(t)
	m.AssertExpectations(t)
}