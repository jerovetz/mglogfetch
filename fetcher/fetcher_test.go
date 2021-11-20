package fetcher

import (
	"bytes"
	"errors"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/stretchr/testify/mock"
	"io"
	"matchwork/mailgun-log-fetcher/utils"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"
)

var responseBodyJson = `
{
"items": [
{"timestamp": 1636646172.343453 },
{"timestamp": 1636646272.123123 }
],
"paging": {
        "next": "next url"
    }
}
`

var emptyResponse = `
{
"items": [],
"paging": {
        "next": "next url"
    }
}
`

var responseBodyJsonWithANew = `
{
"items": [
{"timestamp": 1636646172.232232 },
{"timestamp": 1636646192.324323},
{"timestamp": 1636646272.324232 }
],
"paging": {
        "next": "next url"
    }
}
`

var now = int64(1636646330)

type Clock struct {
	mock.Mock
}

func (t *Clock) Now() TimeInterface {
	args := t.Called()
	return args.Get(0).(TimeInterface)
}

func (t *Clock) Sleep(d time.Duration) {
	t.Called(d)
}

type FakeTime struct {
	mock.Mock
}

func (t *FakeTime) Unix() int64 {
	args := t.Called()
	return args.Get(0).(int64)
}

type HttpClient struct {
	mock.Mock
	RetryMax int
}

func (c *HttpClient) Do(req *retryablehttp.Request) (*http.Response, error) {
	args := c.Called(req)
	return args.Get(0).(*http.Response), args.Error(1)
}

func TestFetchProperRequestReturned(t *testing.T) {
	utils.InitTestEnv()
	url := "url"
	mockTime := new(FakeTime)
	mockClock := new(Clock)
	mockClock.
		On("Now").Return(mockTime)
	mockTime.
		On("Unix").Return(now + 1000000).Once()

	mockClient := new(HttpClient)
	request, _ := retryablehttp.NewRequest("GET", url, nil)
	request.Header.Add("Authorization", "Basic "+basicAuth(os.Getenv("MAILGUN_API_USERNAME"), os.Getenv("MAILGUN_API_SECRET")))
	mockClient.
		On("Do", request).Return(&http.Response{Body: io.NopCloser(bytes.NewBufferString(responseBodyJson)), StatusCode: 200}, nil).Once()

	response := Fetch(url, mockClient, mockClock)

	if len(response.Items) != 2 {
		t.Errorf("Expected item count is 2 (%d)", len(response.Items))
	}

	if !strings.Contains(string(response.Items[0]), "1636646172") {
		t.Errorf("First timestamp bad.")
	}

	if !strings.Contains(string(response.Items[1]), "1636646272") {
		t.Errorf("Second timestamp bad.")
	}

	if response.Paging.Next != "next url" {
		t.Errorf("Next page missing.")
	}

	mockClient.AssertExpectations(t)
}

func TestFetchMustWaitForRequestIsOldEnough(t *testing.T) {
	utils.InitTestEnv()
	url := "url"
	mockClient := new(HttpClient)
	mockTime := new(FakeTime)
	mockClock := new(Clock)
	mockClock.
		On("Now").Return(mockTime).
		On("Sleep", time.Second*10).Once()

	mockTime.
		On("Unix").Return(now).Once().
		On("Unix").Return(now+10).Once()

	request, _ := retryablehttp.NewRequest("GET", url, nil)
	request.Header.Add("Authorization", "Basic "+basicAuth(os.Getenv("MAILGUN_API_USERNAME"), os.Getenv("MAILGUN_API_SECRET")))
	mockClient.
		On("Do", request).Return(&http.Response{Body: io.NopCloser(bytes.NewBufferString(responseBodyJson)), StatusCode: 200}, nil).Once().
		On("Do", request).Return(&http.Response{Body: io.NopCloser(bytes.NewBufferString(responseBodyJsonWithANew)), StatusCode: 200}, nil).Once()

	response := Fetch(url, mockClient, mockClock)

	if len(response.Items) != 3 {
		t.Errorf("Expected item count is 3 (%d arrived)", len(response.Items))
	}

	mockClient.AssertExpectations(t)
	mockClock.AssertExpectations(t)
}

func TestUntilRequestHasItems(t *testing.T) {
	utils.InitTestEnv()
	url := "url"
	mockClient := new(HttpClient)
	mockTime := new(FakeTime)
	mockClock := new(Clock)
	mockClock.
		On("Now").Return(mockTime).
		On("Sleep", time.Second*10).Twice()
	mockTime.
		On("Unix").Return(now + 100000)

	request, _ := retryablehttp.NewRequest("GET", url, nil)
	request.Header.Add("Authorization", "Basic "+basicAuth(os.Getenv("MAILGUN_API_USERNAME"), os.Getenv("MAILGUN_API_SECRET")))
	mockClient.
		On("Do", request).Return(&http.Response{Body: io.NopCloser(bytes.NewBufferString(emptyResponse)), StatusCode: 200}, nil).Twice().
		On("Do", request).Return(&http.Response{Body: io.NopCloser(bytes.NewBufferString(responseBodyJson)), StatusCode: 200}, nil).Once()

	response := Fetch(url, mockClient, mockClock)

	if len(response.Items) != 2 {
		t.Errorf("Expected item count is 3 (%d arrived)", len(response.Items))
	}

	mockClient.AssertExpectations(t)
	mockClock.AssertExpectations(t)
}

func TestFetchFailedExitedWithPanic(t *testing.T) {
	utils.InitTestEnv()
	url := "url"
	mockTime := new(FakeTime)
	mockClock := new(Clock)
	mockClock.
		On("Now").Return(mockTime)

	mockTime.
		On("Unix").Return(now + 1000000).Once()

	mockClient := new(HttpClient)
	mockClient.On("Do", mock.Anything).Return(&http.Response{}, errors.New("irrelevant error"))

	defer func() {
		f := recover()
		if !strings.Contains(f.(string), "Client failed") {
			t.Errorf("Client panic expected.")
		}
	}()

	Fetch(url, mockClient, mockClock)
}

func TestFetchGotNo2xxAndRetriesExhaustedPanic(t *testing.T) {
	utils.InitTestEnv()
	url := "url"
	mockTime := new(FakeTime)
	mockClock := new(Clock)
	mockClock.
		On("Now").Return(mockTime)
	mockTime.
		On("Unix").Return(now + 1000000).Once()

	mockClient := new(HttpClient)
	mockClient.On("Do", mock.Anything).Return(&http.Response{StatusCode: 404}, nil)

	defer func() {
		f := recover()
		if !strings.Contains(f.(string), "Statuscode was") {
			t.Errorf("Statuscode panic expected.")
		}
	}()

	Fetch(url, mockClient, mockClock)
}

func TestJsonShouldBeCompacted(t *testing.T) {
	var formattedJson = `
{"items": [{
	"this":             { "is":
	"formatted"}
}]}
`
	utils.InitTestEnv()
	url := "url"
	mockTime := new(FakeTime)
	mockClock := new(Clock)
	mockClock.
		On("Now").Return(mockTime)
	mockTime.
		On("Unix").Return(now + 1000000).Once()

	mockClient := new(HttpClient)
	mockClient.On("Do", mock.Anything).Return(&http.Response{Body: io.NopCloser(bytes.NewBufferString(formattedJson)), StatusCode: 200}, nil)

	response := Fetch(url, mockClient, mockClock)
	if string(response.Items[0]) != `{"this":{"is":"formatted"}}` {
		t.Errorf("Json is not compact.")
	}
}
