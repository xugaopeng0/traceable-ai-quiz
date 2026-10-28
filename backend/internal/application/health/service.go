package health

import "traceable-ai-quiz/internal/domain"

type Service struct{ name, version string }

func New(name, version string) Service { return Service{name, version} }

func (s Service) Status() domain.Health {
	return domain.Health{Status: "ok", Service: s.name, Version: s.version}
}
