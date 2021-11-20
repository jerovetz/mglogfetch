package pusher

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"matchwork/mailgun-log-fetcher/utils"
	"os"
	"testing"
	"time"
)

func TestItemsPushedToHostAndPortWithTlsWithRealConnection(t *testing.T) {
	os.Remove("socket_out")
	utils.InitTestEnv()
	var items []json.RawMessage
	done := make(chan bool)
	file, _ := os.Create("socket_out")

	mockNow := new(MockNow)
	mockNow.On("Format", time.RFC3339).Once()
	now = func() TimeInterface {
		return mockNow
	}

	go listenOnHost(done, file)
	waitUntilServerIsReady(done)

	json.Unmarshal([]byte(itemsAsString), &items)
	expectedItems := prepareExpectedItems(items)

	config := &tls.Config{InsecureSkipVerify: true}
	con, _ := tls.Dial("tcp", os.Getenv("REMOTE_LOG_HOST"), config)

	pusher := Pusher{connection: con}
	ok := pusher.Push(items)

	assertNoErrors(t, ok)
	assertItemsSent(t, expectedItems)

	file.Close()
	os.Remove("socket_out")
}

func assertItemsSent(t *testing.T, expectedItems json.RawMessage) {
	content, _ := ioutil.ReadFile("socket_out")
	if !bytes.Equal(content, expectedItems) {
		t.Errorf("Failed asserting socket output %s is equal with %s", content, expectedItems)
	}
}

func prepareExpectedItems(items []json.RawMessage) json.RawMessage {
	message := []byte(fmt.Sprintf("<80>1 %s %s %s %d - - ", nowString, os.Getenv("LOG_HOSTNAME"), os.Getenv("MAIL_DOMAIN"), os.Getpid()))
	expectedItems := append(message, items[0]...)
	for index, item := range items {
		if index != 0 {
			withHostname := append(message, item...)
			expectedItems = append(append(expectedItems, []byte("\n")...), withHostname...)
		}
	}
	expectedItems = append(expectedItems, []byte("\n")...)
	return expectedItems
}

func listenOnHost(done chan bool, file io.Writer) {
	cer, err := tls.LoadX509KeyPair("test.crt", "test.key")
	if err != nil {
		log.Println(err)
		return
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, _ := tls.Listen("tcp", os.Getenv("REMOTE_LOG_HOST"), config)
	done <- true
	con, _ := ln.Accept()
	reader := bufio.NewReader(con)
	reader.WriteTo(file)
}

func waitUntilServerIsReady(done chan bool) {
	ready := <-done
	for ready == false {
		ready = <-done
	}
}
