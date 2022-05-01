package scale

import (
	"fmt"
	"net/http"

	deploy "github.com/Ishan27g/ryo-Faas/pkg/proto"
	"github.com/Ishan27g/ryo-Faas/pkg/types"
	"github.com/gin-gonic/gin"
)

type deployment types.Metric

var sc = scaler{
	current: map[string]chan *deployment{},
}

type scaler struct {
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
	return true
}
func (s *scaler) update(m types.Metric) {
	s.current[m.Name] = make(chan *deployment, 1)
	d := deployment(m)
	s.current[m.Name] <- &d
}
func (s *scaler) scale(mt []types.Metric) {
	var result []types.Metric
	for _, newMetric := range mt {
		if newMetric.Name == "scale" {
			continue
		}

		for _, deployments := range s.current {
			scale := func() {
				d := <-deployments
				defer func() { s.current[d.Name] <- d }()
				if newMetric.Name != d.Name {
					return
				}
				if newMetric.Count == d.Count {
					return
				}
				fmt.Println("Change detected: ", d.Name, " : count - ", newMetric.Count-d.Count)
				result = append(result, types.Metric{
					Name:    newMetric.Name,
					Count:   newMetric.Count - d.Count, // +/- based on up/down
					IsAsync: newMetric.IsAsync,
					IsMain:  newMetric.IsMain,
				})
			}
			scale()
		}
		if !s.add(newMetric) {
			s.update(newMetric)
		}
	}
	//proxy := transport.ProxyGrpcClient("rfa-proxy:9998")
	//ctx, can := context.WithTimeout(context.Background(), 60*time.Second)
	//defer can()

	for _, metric := range result {
		var fns []*deploy.Function
		fns = append(fns, &deploy.Function{
			Entrypoint: metric.Name,
			Async:      metric.IsAsync,
			IsMain:     metric.IsMain,
		})
		if metric.Count > 0 {
			for i := 0; i < metric.Count; i++ {
				fmt.Println("--- proxy DEPLOY ", metric.Name)
				//if _, err := proxy.Deploy(ctx, &deploy.DeployRequest{Functions: fns}); err != nil {
				//	fmt.Println(err.Error())
				//}
			}
		} else {
			for i := metric.Count; i < 0; i++ {
				fmt.Println("---- proxy STOP ", metric.Name)
				//if _, err := proxy.Stop(ctx, &deploy.Empty{Rsp: &deploy.Empty_Entrypoint{Entrypoint: metric.Name}}); err != nil {
				//	fmt.Println(err.Error())
				//}
			}
		}
		//can()
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
