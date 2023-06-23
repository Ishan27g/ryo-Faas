package internal

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/Ishan27g/sentri/internal/transport"
)

const ()

var r *reader

type reader struct {
	name string
	r    *io.PipeReader
}

func pushToNats(i int, o *Output) {
	b, _ := json.Marshal(o)
	transport.NatsPublish(Channel+fmt.Sprintf("%d", i), b)
}

func Ready(name string, pr *io.PipeReader) {
	r = &reader{name, pr}
	var i = 0
	go func() {
		//p := pool.GetPool()
		s := bufio.NewScanner(r.r)
		for s.Scan() {
			//p.Submit(func() {
			o := Output{From: r.name, Log: s.Text(), Timestamp: time.Now()}
			pushToNats(i, &o)
			//})
			i++
		}
	}()
}
