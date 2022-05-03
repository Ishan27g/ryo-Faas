package scale

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
	"github.com/Ishan27g/ryo-Faas/pkg/transport"
	"github.com/Ishan27g/ryo-Faas/pkg/types"
	"github.com/gin-gonic/gin"
)

type deployment types.Metric

var sc = scaler{
	Mutex:   sync.Mutex{},
	current: map[string]chan *deployment{},
}

type scaler struct {
	sync.Mutex
	current map[string]chan *deployment
}

func (s *scaler) get(entrypoint string) *deployment {
	if s.current[entrypoint] == nil {
		return nil
	}
	var dp = new(deployment)
	d := <-s.current[entrypoint]
	dp = d
	s.current[entrypoint] <- d
	return dp
}
func (s *scaler) add(m types.Metric) bool {
	if s.get(m.Name) != nil {
		return false
	}
	s.current[m.Name] = make(chan *deployment, 1)
	d := deployment(m)
	s.current[m.Name] <- &d
	fmt.Println("added ", m.Name, " for ", m.Count)
	return true
}
func (s *scaler) update(m types.Metric) {
	s.current[m.Name] = make(chan *deployment, 1)
	d := deployment(m)
	s.current[m.Name] <- &d
	fmt.Println("updated ", m.Name, " to ", m.Count)
}
func (s *scaler) scale(mt []types.Metric) {
	s.Lock()
	defer s.Unlock()
	var result []types.Metric
	for _, newMetric := range mt {
		if newMetric.Name == "scale" {
			continue
		}
		if s.get(newMetric.Name) == nil {
			s.add(types.Metric{Name: newMetric.Name, Count: 1})
		}
		fmt.Println("Received ", newMetric.Name, newMetric.Count)
		for _, deployments := range s.current {
			scale := func(deployments chan *deployment) {
				d := <-deployments
				defer func() { s.current[d.Name] <- d }()
				if newMetric.Name != d.Name {
					return
				}
				if newMetric.Count == d.Count {
					return
				}
				fmt.Println("Existing ", d.Name, d.Count)
				fmt.Println("Change detected: ", d.Name, " : count - ", newMetric.Count-d.Count)
				result = append(result, types.Metric{
					Name:  newMetric.Name,
					Count: newMetric.Count - d.Count, // +/- based on up/down
				})
			}
			scale(deployments)
		}
		if !s.add(newMetric) {
			s.update(newMetric)
		}
	}
	proxy := transport.ProxyGrpcClient("rfa-proxy:9998")
	ctx, can := context.WithTimeout(context.Background(), 60*time.Second)
	defer can()

	for _, metric := range result {
		var fns []*deploy.Function
		fns = append(fns, &deploy.Function{
			Entrypoint: metric.Name,
		})
		if metric.Count > 0 {
			for i := 0; i < metric.Count; i++ {
				if _, err := proxy.Deploy(ctx, &deploy.DeployRequest{Functions: fns}); err != nil {
					fmt.Println(err.Error())
				}
			}
		} else {
			for i := metric.Count; i < 0; i++ {
				if _, err := proxy.Stop(ctx, &deploy.Empty{Rsp: &deploy.Empty_Entrypoint{Entrypoint: metric.Name}}); err != nil {
					fmt.Println(err.Error())
				}
			}
		}
	}
}

// Scale
func Scale(c *gin.Context) {
	var m []types.Metric
	if c.ShouldBindJSON(&m) != nil {
		c.JSON(http.StatusBadRequest, nil)
	}
	go sc.scale(m)
	c.JSON(http.StatusAccepted, nil)
}
