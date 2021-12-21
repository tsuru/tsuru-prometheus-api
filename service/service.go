package service

import (
	"fmt"

	"github.com/prometheus/prometheus/pkg/rulefmt"
)

type Service interface {
	EnsurePrometheusRule(pool string, ruleName string, ruleGroups rulefmt.RuleGroups) error
}

type serviceImpl struct {
	tsuruHost  string
	tsuruToken string
}

func NewService(tsuruHost, tsuruToken string) Service {
	return &serviceImpl{
		tsuruHost:  tsuruHost,
		tsuruToken: tsuruToken,
	}
}

func (s *serviceImpl) EnsurePrometheusRule(pool string, ruleName string, ruleGroups rulefmt.RuleGroups) error {
	fmt.Println("TODO: implement")
	fmt.Println("pool:", pool)
	fmt.Println("ruleName:", ruleName)
	return nil
}
