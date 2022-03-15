package transport

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/nats-io/nats.go"
)

var opts []nats.Option
var urls = flag.String("s", nats.DefaultURL, "The nats server URLs (separated by comma)")
var showTime = flag.Bool("t", false, "Display timestamps")

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
	totalWait := 10 * time.Minute
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

	var userCreds = flag.String("creds", "", "User Credentials File")
	var nkeyFile = flag.String("nkey", "", "NKey Seed File")
	var tlsClientCert = flag.String("tlscert", "", "TLS client certificate file")
	var tlsClientKey = flag.String("tlskey", "", "Private key file for client certificate")
	var tlsCACert = flag.String("tlscacert", "", "CA certificate to verify peer against")
	var showHelp = flag.Bool("h", false, "Show help message")

	log.SetFlags(0)
	flag.Usage = usage
	flag.Parse()

	if *showHelp {
		showUsageAndExit(0)
	}

	args := flag.Args()
	if len(args) != 1 {
		showUsageAndExit(1)
	}

	// Connect Options.
	opts := []nats.Option{nats.Name("NATS Sample Subscriber")}
	opts = setupConnOptions(opts)

	if *userCreds != "" && *nkeyFile != "" {
		log.Fatal("specify -seed or -creds")
	}

	// Use UserCredentials
	if *userCreds != "" {
		opts = append(opts, nats.UserCredentials(*userCreds))
	}

	// Use TLS client authentication
	if *tlsClientCert != "" && *tlsClientKey != "" {
		opts = append(opts, nats.ClientCert(*tlsClientCert, *tlsClientKey))
	}

	// Use specific CA certificate
	if *tlsCACert != "" {
		opts = append(opts, nats.RootCAs(*tlsCACert))
	}

	// Use Nkey authentication.
	if *nkeyFile != "" {
		opt, err := nats.NkeyOptionFromSeed(*nkeyFile)
		if err != nil {
			log.Fatal(err)
		}
		opts = append(opts, opt)
	}
}

func sub(subj string, cb func(msg *nats.Msg)) {
	// Connect to NATS
	nc, err := nats.Connect(*urls, opts...)
	if err != nil {
		log.Fatal(err)
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
	if *showTime {
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

	if subjects[subj] == nil {
		subjects[subj] = &subjectMeta{subjectName: subj, docId: ""}
	}
	var err error
	var nc *nats.Conn
	nc, err = nats.Connect(*urls, opts...)
	if err != nil {
		log.Fatal(err)
	}
	defer nc.Close()
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
