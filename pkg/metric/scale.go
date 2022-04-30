package metric

const (
	ScaleZero = 0
	ScaleMin  = 1
	ScaleTwo  = 2
	ScaleMax  = 3
)

type Scale struct {
	functions map[string]*int
}

func (s *Scale) scale(invocations ...invocation) map[string]int {
	rsp := make(map[string]int)
	for _, i := range invocations {
		var prv, curr int
		if s.functions[i.name] != nil {
			prv = *s.functions[i.name]
		}
		if i.count == 0 {
			curr = ScaleZero
		} else if i.count >= 1 && i.count < 4 {
			curr = ScaleMin
		} else if i.count >= 4 && i.count < 7 {
			curr = ScaleTwo
		} else {
			curr = ScaleMax
		}
		if prv != curr {
			s.functions[i.name] = &curr
		}
		rsp[i.name] = curr
	}
	return rsp
}
