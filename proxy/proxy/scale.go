package proxy

type scaler struct {
	current map[string]chan []deployment
}
type deployment struct {
	entrypoint    string
	containerName string
}

func (s *scaler) get(entrypoint string) []deployment {
	d := <-s.current[entrypoint]
	defer func() { s.current[entrypoint] <- d }()
	return d
}
func (s *scaler) add(entrypoint, containerName string) {
	if s.current[entrypoint] == nil {
		s.current[entrypoint] = make(chan []deployment, 1)
		s.current[entrypoint] <- []deployment{}
	}
	d := <-s.current[entrypoint]
	s.current[entrypoint] <- append(d, deployment{
		entrypoint:    entrypoint,
		containerName: containerName,
	})
}
func (s *scaler) scale(scaleFactors map[string]int) {
	for fnName, i := range scaleFactors {
		current := s.get(fnName)
		if len(current) != i {
			for index := 0; index < i; index++ {

			}
		}
	}
}
