package scale

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"github.com/Ishan27g/ryo-Faas/pkg/types"
)

type ExportMetrics func() []types.Metric

func (em ExportMetrics) Start(ctx context.Context, endpoint string) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-time.After(6 * time.Second):
			b, err := json.Marshal(em())
			if err != nil {
				continue
			}
			_, status := transport.SendHttp("POST", endpoint, b)
			fmt.Println("metric exported : status- ", status)
		}
	}
}
