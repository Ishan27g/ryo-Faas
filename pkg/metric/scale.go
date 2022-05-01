package metric

const (
	ScaleMin = 1
	ScaleTwo = 2
	ScaleMax = 3
)

type Scale struct{}

func (s *Scale) scale(invocations ...invocation) map[string]int {
	rsp := make(map[string]int)
	for _, i := range invocations {
		var curr int
		if i.count < 4 {
			curr = ScaleMin
		} else if i.count >= 4 && i.count < 7 {
			curr = ScaleTwo
		} else {
			curr = ScaleMax
		}
		rsp[i.name] = curr
	}
	return rsp
}
