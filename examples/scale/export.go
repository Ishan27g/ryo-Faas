package scale

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"github.com/Ishan27g/ryo-Faas/pkg/types"
)

const exportTimeout = 6 * time.Second

var exporter *export
var one = sync.Once{}

type export struct {
	endpoint string
	context.Context
	context.CancelFunc
}

func StopExporter() {
	if exporter == nil {
		return
	}
	exporter.CancelFunc()
	exporter = nil
	one = sync.Once{}
}
func StartExporter(monitor *Monitor, endpoint string) {
	go one.Do(func() {
		ctx, can := context.WithCancel(context.Background())
		exporter = &export{endpoint, ctx, can}
		for {
			select {
			case <-exporter.Done():
				return
			case <-time.After(exportTimeout):
				b, err := json.Marshal(scaleMetrics(monitor.getInvocations()...))
				if err != nil {
					fmt.Println(err.Error())
					continue
				}
				_, _ = transport.SendHttp("POST", exporter.endpoint, b)
				// fmt.Println("metrics exported : status- ", status)
			}
		}
	})
}

func scaleMetrics(inv ...invocation) *[]types.Metric {
	var m []types.Metric
	for _, i := range inv {
		var curr int
		if i.count < 4 {
			curr = min
		} else if i.count >= 4 && i.count < 7 {
			curr = two
		} else {
			curr = max
		}
		m = append(m, types.Metric{
			Name:  i.name,
			Count: curr,
		})
	}
	fmt.Println("Sending - ", m)
	return &m
}
