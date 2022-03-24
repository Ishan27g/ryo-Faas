package transport

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

const (
	HttpAsync string = "httpAsync"
)

var opts []nats.Option
var urls = os.Getenv("NATS")

// "nats://localhost:4222"
//var urls = "nats://raf-nats:4222"
var showTime = false

var subjects map[string]*subjectMeta

type SubCb func()

type subjectMeta struct {
	subjectName string
	docId       string
}

func usage() {
	log.Printf("Usage: nats-sub [-s server] [-creds file] [-nkey file] [-tlscert file] [-tlskey file] [-tlscacert file] [-t] <subject>\n")
	flag.PrintDefaults()
}

func showUsageAndExit(exitcode int) {
	usage()
	os.Exit(exitcode)
}
func printMsg(m *nats.Msg, i int) {
	log.Printf("[#%d] Received on [%s]: '%s'", i, m.Subject, string(m.Data))
}
func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Second
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))
	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))
	opts = append(opts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		log.Printf("Disconnected due to:%s, will attempt reconnects for %.0fm", err, totalWait.Minutes())
	}))
	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatalf("Exiting: %v", nc.LastError())
	}))
	return opts
}
func init() {

	subjects = make(map[string]*subjectMeta)

	// Connect Options.
	opts := []nats.Option{nats.Name("NATS Sample Subscriber")}
	opts = setupConnOptions(opts)

}

func sub(subj string, cb func(msg *nats.Msg)) {
	// Connect to NATS
	nc, err := nats.Connect(urls, opts...)
	//nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Println(err)
		return
	}
	i := 0
	nc.Subscribe(subj, func(msg *nats.Msg) {
		i += 1
		cb(msg)
	})
	nc.Flush()

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on [%s]", subj)
	if showTime {
		log.SetFlags(log.LstdFlags)
	}
}

func NatsSubscribeDoc(subj string, docId string, cb func(msg *nats.Msg)) {

	if subjects[subj] == nil {
		subjects[subj] = &subjectMeta{subjectName: subj, docId: docId}
	}
	sub(subj+"."+docId, cb)
}

func NatsSubscribe(subj string, cb func(msg *nats.Msg)) {
	if subjects[subj] == nil {
		subjects[subj] = &subjectMeta{subjectName: subj, docId: ""}
	}
	sub(subj, cb)
}
func NatsPublish(subj string, msg string, reply *string) bool {

	nc, err := nats.Connect(urls, opts...)
	//nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Println(err)
		return false
	}
	defer nc.Close()
	if subjects[subj] == nil {
		subjects[subj] = &subjectMeta{subjectName: subj, docId: ""}
	}
	if reply != nil && *reply != "" {
		err = nc.PublishRequest(subj, *reply, []byte(msg))
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	} else {
		err = nc.Publish(subj, []byte(msg))
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	}

	err = nc.Flush()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if err := nc.LastError(); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Published [%s] : '%s'\n", subj, msg)
	}
	return true
}

type AsyncNats struct {
	Callback   string
	Entrypoint string
	Req        []byte
	// HttpFunction
}

func NatsPublishJson(subj string, msg AsyncNats, reply *string) bool {
	nc, err := nats.Connect(urls, opts...)
	if err != nil {
		log.Println(err)
		return false
	}
	defer nc.Close()
	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Println(err)
		return false
	}
	defer ec.Close()
	// Publish the message
	if reply != nil && *reply != "" {
		err = ec.PublishRequest(subj, *reply, &msg)
		if err != nil {
			fmt.Println(err.Error())
			return false
		}
	}
	if err := ec.Publish(subj, &msg); err != nil {
		log.Println(err)
		return false
	}

	err = ec.Flush()
	if err != nil {
		fmt.Println(err.Error())
		return false
	}

	if err := ec.LastError(); err != nil {
		log.Fatal(err)
	} else {
		log.Printf("Published [%s] : '%s'\n", subj, msg)
	}
	return true
}

func NatsSubscribeJson(subj string, cb func(msg *AsyncNats)) {
	if subjects[subj] == nil {
		subjects[subj] = &subjectMeta{subjectName: subj, docId: ""}
	}
	opts = append(opts, nats.ErrorHandler(func(nc *nats.Conn, s *nats.Subscription, err error) {
		if s != nil {
			log.Printf("Async error in %q/%q: %v", s.Subject, s.Queue, err)
		} else {
			log.Printf("Async error outside subscription: %v", err)
		}
	}))
	nc, err := nats.Connect(urls, opts...)
	if err != nil {
		log.Println(err)
		return
	}
	defer nc.Close()
	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}
	defer ec.Close()
	wg := sync.WaitGroup{}
	wg.Add(1)

	if _, err := ec.Subscribe(subj, func(s *AsyncNats) {
		cb(s)
		wg.Done()
	}); err != nil {
		log.Fatal(err)
	}

	// Wait for a message to come in
	wg.Wait()

	if err := ec.LastError(); err != nil {
		log.Fatal(err)
	}

	log.Printf("Listening on [%s]", subj)
	if showTime {
		log.SetFlags(log.LstdFlags)
	}

}
