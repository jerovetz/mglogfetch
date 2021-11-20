package pusher

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"os"
	"time"
)

var now = func() TimeInterface {
	return time.Now()
}

type TimeInterface interface {
	Format(layout string) string
}

type PusherInterface interface {
	Push(items []json.RawMessage) error
}

type ConnInterface interface {
	Write(b []byte) (int, error)
	Close() error
}

type Pusher struct {
	connection ConnInterface
}

func New() PusherInterface {
	config := &tls.Config{}
	con, err := tls.Dial("tcp", os.Getenv("REMOTE_LOG_HOST"), config)
	if err != nil {
		panic("Failed to connect to remote host.")
	}
	return &Pusher{connection: con}
}

func (p *Pusher) Push(items []json.RawMessage) error {
	hostnameTagPid := fmt.Sprintf("%s %s %d", os.Getenv("LOG_HOSTNAME"), os.Getenv("MAIL_DOMAIN"), os.Getpid())
	syslogFields := []byte(fmt.Sprintf("<80>1 %s %s - - ", now().Format(time.RFC3339), hostnameTagPid))

	for _, item := range items {
		item = append(syslogFields, item...)
		p.connection.Write(item)
		p.connection.Write([]byte("\n"))
	}
	time.Sleep(100 * time.Millisecond)
	p.connection.Close()
	return nil
}
