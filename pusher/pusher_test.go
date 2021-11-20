package pusher

import (
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/mock"
	"matchwork/mailgun-log-fetcher/utils"
	"os"
	"testing"
	"time"
)

var itemsAsString = `[
        {
            "geolocation": {
                "country": "HU",
                "region": "Unknown",
                "city": "Unknown"
            },
            "tags": [],
            "url": "https://mindenallas.hu/employer/login?token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6MjQsInJvbGUiOiJlbXBsb3llciIsImlhdCI6MTYzNjI5NTIwNSwiZXhwIjoxNjM2OTAwMDA1fQ.HL-NaUpM2ywICrJdDsJgQfagJQ5VvzaQHP6_KCqD4bw",
            "ip": "91.120.140.136",
            "log-level": "info",
            "event": "clicked",
            "campaigns": [],
            "user-variables": {},
            "recipient-domain": "bocskayker.hu",
            "timestamp": 1636532241.823582,
            "client-info": {
                "client-name": "Firefox",
                "client-type": "browser",
                "user-agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:94.0) Gecko/20100101 Firefox/94.0",
                "device-type": "desktop",
                "client-os": "Windows"
            },
            "message": {
                "headers": {
                    "message-id": "20211107142645.768e2d2688611fc6@jobs.mindenallas.hu"
                }
            },
            "recipient": "info@bocskayker.hu",
            "id": "ZYqv-femT-eZdNugraadzQ"
        },
        {
            "tags": null,
            "timestamp": 1636532679.442269,
            "storage": {
                "url": "https://storage.eu.mailgun.net/v3/domains/jobs.mindenallas.hu/messages/AwABBXlYun-yAYEIzS5EILAUFyRLjLYtZA==",
                "key": "AwABBXlYun-yAYEIzS5EILAUFyRLjLYtZA=="
            },
            "envelope": {
                "sender": "applicant-YTWNLhMCjIXB@jobs.mindenallas.hu",
                "transport": "smtp",
                "targets": "info@bocskayker.hu"
            },
            "recipient-domain": "bocskayker.hu",
            "id": "LMbm5AS1S8SQDHvUE-VGug",
            "method": "HTTP",
            "user-variables": {},
            "flags": {
                "is-authenticated": true,
                "is-test-mode": false
            },
            "log-level": "info",
            "message": {
                "headers": {
                    "to": "info@bocskayker.hu",
                    "message-id": "20211110082439.c2c2aae2c82c7083@jobs.mindenallas.hu",
                    "from": "Hurka <applicant-YTWNLhMCjIXB@jobs.mindenallas.hu>",
                    "subject": "Jelentkező a hirdetésedre - Hurka"
                },
                "size": 50994
            },
            "recipient": "info@bocskayker.hu",
            "event": "accepted"
        },
        {
            "tags": [],
            "storage": {
                "url": "https://storage.eu.mailgun.net/v3/domains/jobs.mindenallas.hu/messages/AwABBXlYun-yAYEIzS5EILAUFyRLjLYtZA==",
                "key": "AwABBXlYun-yAYEIzS5EILAUFyRLjLYtZA=="
            },
            "envelope": {
                "transport": "smtp",
                "sender": "applicant-YTWNLhMCjIXB@jobs.mindenallas.hu",
                "sending-ip": "185.250.239.5",
                "targets": "info@bocskayker.hu"
            },
            "delivery-status": {
                "tls": true,
                "mx-host": "bocskayker.hu",
                "attempt-no": 1,
                "description": "",
                "session-seconds": 47.27060604095459,
                "code": 250,
                "message": "OK",
                "certificate-verified": true
            },
            "event": "delivered",
            "campaigns": [],
            "log-level": "info",
            "user-variables": {},
            "flags": {
                "is-routed": false,
                "is-authenticated": true,
                "is-system-test": false,
                "is-test-mode": false
            },
            "recipient-domain": "bocskayker.hu",
            "timestamp": 1636532734.023108,
            "message": {
                "headers": {
                    "to": "info@bocskayker.hu",
                    "message-id": "20211110082439.c2c2aae2c82c7083@jobs.mindenallas.hu",
                    "from": "Hurka <applicant-YTWNLhMCjIXB@jobs.mindenallas.hu>",
                    "subject": "Jelentkező a hirdetésedre - Hurka"
                },
                "attachments": [],
                "size": 50994
            },
            "recipient": "info@bocskayker.hu",
            "id": "EEbmpTfvS2amOYYAK8JsIA"
        }
    ]`

type MockConn struct {
	mock.Mock
}

func (t *MockConn) Write(b []byte) (int, error) {
	t.Called(b)
	return 0, nil
}

func (t *MockConn) Close() error {
	t.Called()
	return nil
}

var nowString = "now string"

type MockNow struct {
	mock.Mock
}

func (m *MockNow) Format(layout string) string {
	m.Called(layout)
	return  nowString
}

func TestItemsPushedToHostAndPortWithTls(t *testing.T) {
	utils.InitTestEnv()
	var items []json.RawMessage

	json.Unmarshal([]byte(itemsAsString), &items)
	expectedItems := getExpectedItems(items)
	mockNow := new(MockNow)
	mockNow.On("Format", time.RFC3339).Once()
	now = func() TimeInterface {
		return mockNow
	}

	mockConn := new(MockConn)
	mockConn.
		On("Write", []byte(expectedItems[0])).Once().
		On("Write", []byte("\n")).Once().
		On("Write", []byte(expectedItems[1])).Once().
		On("Write", []byte("\n")).Once().
		On("Write", []byte(expectedItems[2])).Once().
		On("Write", []byte("\n")).Once()
	mockConn.On("Close").Once()

	pusher := Pusher{mockConn}
	ok := pusher.Push(items)

	assertNoErrors(t, ok)
	mockConn.AssertExpectations(t)
}

func assertNoErrors(t *testing.T, ok error) {
	if ok != nil {
		t.Errorf("Ok expected for result.")
	}
}

func getExpectedItems(items []json.RawMessage) []json.RawMessage {
	message := fmt.Sprintf("<80>1 %s %s %s %d - - ", nowString, os.Getenv("LOG_HOSTNAME"), os.Getenv("MAIL_DOMAIN"), os.Getpid())
	var decoratedItems []json.RawMessage
	for _, item := range items {
		syslogFields := []byte(message)
		row := append(syslogFields, item...)
		decoratedItems = append(decoratedItems, row)
	}

	return decoratedItems
}