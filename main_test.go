package main

import (
	"encoding/json"
	"github.com/stretchr/testify/mock"
	"matchwork/mailgun-log-fetcher/fetcher"
	"matchwork/mailgun-log-fetcher/pusher"
	"matchwork/mailgun-log-fetcher/utils"
	"os"
	"strconv"
	"strings"
	"testing"
)

type PushMock struct {
	mock.Mock
}

func (m *PushMock) Push(items []json.RawMessage) error {
	m.Called(items)
	return nil
}

type Mocks struct {
	mock.Mock
}

func (m *Mocks) fetch(url string, client fetcher.HttpClientInterface, clock fetcher.ClockInterface) fetcher.Response {
	if len(m.Calls) == 2 {
		panic("Break infinite loop, done")
	}
	args := m.Called(url, client, clock)
	return args.Get(0).(fetcher.Response)
}

func TestItemsPackedToPusher(t *testing.T) {
	utils.InitTestEnv()
	firstUrl := mailgunEuDomain + os.Getenv("MAIL_DOMAIN") +"/events?begin="+ strconv.FormatInt(now, 10) +"&ascending=yes"
	response := fetcher.Response{
		Items:  nil,
		Paging: fetcher.Paging{
			Next: "next url",
		},
	}

	fetchFunction := new(Mocks)
	fetchFunction.
		On("fetch", firstUrl, mock.Anything, mock.Anything ).Return(response).Once().
		On("fetch", response.Paging.Next, mock.Anything, mock.Anything ).Return(response).Once()

	pushMock := new(PushMock)
	pushMock.On("Push", response.Items).Return(nil).Twice()

	fetchAction = fetchFunction.fetch
	pusherCreator = func () pusher.PusherInterface {
		return pushMock
	}

	defer func() {
		f := recover()
		if !strings.Contains(f.(string), "Break infinite") {
			t.Errorf("another panic expected. %s", f)
		}
	}()

	main()

	pushMock.AssertExpectations(t)
	fetchFunction.AssertExpectations(t)
}

func TestItemsPackedToPusherWithUsRegion(t *testing.T) {
	utils.InitTestEnv()
	originalVar := os.Getenv("MAILGUN_REGION")
	os.Setenv("MAILGUN_REGION", "us")
	firstUrl := mailgunUsDomain + os.Getenv("MAIL_DOMAIN") +"/events?begin="+ strconv.FormatInt(now, 10) +"&ascending=yes"
	response := fetcher.Response{
		Items:  nil,
		Paging: fetcher.Paging{
			Next: "next url",
		},
	}

	fetchFunction := new(Mocks)
	fetchFunction.
		On("fetch", firstUrl, mock.Anything, mock.Anything ).Return(response).Once().
		On("fetch", response.Paging.Next, mock.Anything, mock.Anything ).Return(response).Once()

	pushMock := new(PushMock)
	pushMock.On("Push", response.Items).Return(nil).Twice()

	fetchAction = fetchFunction.fetch
	pusherCreator = func () pusher.PusherInterface {
		return pushMock
	}

	defer func() {
		f := recover()
		if !strings.Contains(f.(string), "Break infinite") {
			t.Errorf("another panic expected. %s", f)
		}
	}()

	main()
	os.Setenv("MAILGUN_DOMAIN", originalVar)
}

func TestInvalidRegionFailed(t *testing.T) {
	utils.InitTestEnv()
	originalVar := os.Getenv("MAILGUN_REGION")
	os.Setenv("MAILGUN_REGION", "")
	fetchFunction := new(Mocks)
	pushMock := new(PushMock)

	fetchAction = fetchFunction.fetch
	pusherCreator = func () pusher.PusherInterface {
		return pushMock
	}

	defer func() {
		f := recover()
		if !strings.Contains(f.(string), "no url for current region") {
			t.Errorf("another panic expected. %s", f)
		}
	}()

	main()
	os.Setenv("MAILGUN_DOMAIN", originalVar)
}