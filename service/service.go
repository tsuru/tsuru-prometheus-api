package service

import (
	"context"
	"fmt"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/pkg/rulefmt"
	"github.com/tsuru/go-tsuruclient/pkg/client"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/client-go/rest"
	sigsk8sclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type Service interface {
	EnsurePrometheusRule(pool string, ruleName string, ruleGroups rulefmt.RuleGroups) error
}

type serviceImpl struct {
	tsuruToken      string
	kubernetesToken string
}

func NewService(tsuruToken, kubernetesToken string) Service {
	return &serviceImpl{
		tsuruToken:      tsuruToken,
		kubernetesToken: kubernetesToken,
	}
}

var (
	scheme = runtime.NewScheme()
	_      = monitoringv1.AddToScheme(scheme)
)

func (s *serviceImpl) EnsurePrometheusRule(pool string, ruleName string, ruleGroups rulefmt.RuleGroups) error {
	ctx := context.Background()

	tsuruClient, err := client.ClientFromEnvironment(&tsuru.Configuration{
		DefaultHeader: map[string]string{
			"Authorization": fmt.Sprintf("Bearer %s", s.tsuruToken),
		},
	})

	if err != nil {
		return err
	}

	cluster, err := getPoolCluster(ctx, tsuruClient, pool)
	if err != nil {
		return err
	}

	kubernetesRestConfig := &rest.Config{
		Host:        cluster.Addresses[0],
		BearerToken: s.kubernetesToken,
	}

	k8sClient, err := sigsk8sclient.New(kubernetesRestConfig, sigsk8sclient.Options{Scheme: scheme})
	if err != nil {
		return err
	}

	namespace := "tsuru-" + pool
	desiredPrometheusRule := &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ruleName,
			Namespace: namespace,
			Annotations: map[string]string{
				"app.kubernetes.io/managed-by": "tsuru",
				"app.kubernetes.io/part-of":    "tsuru-prometheus-api",
			},
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: kubernetizeRuleGroups(ruleGroups.Groups),
		},
	}

	currentPrometheusRule := &monitoringv1.PrometheusRule{}

	err = k8sClient.Get(ctx, types.NamespacedName{Namespace: namespace, Name: ruleName}, currentPrometheusRule)

	needsCreation := k8sErrors.IsNotFound(err)
	if needsCreation {
		err = k8sClient.Create(ctx, desiredPrometheusRule)
	}

	if err != nil {
		return err
	}

	if !needsCreation {
		err := k8sClient.Update(ctx, desiredPrometheusRule)
		if err != nil {
			return err
		}
	}

	return nil
}

func getPoolCluster(ctx context.Context, tsuruClient *tsuru.APIClient, pool string) (*tsuru.Cluster, error) {
	clusters, _, err := tsuruClient.ClusterApi.ClusterList(ctx)
	if err != nil {
		return nil, err
	}

	var chosenCluster *tsuru.Cluster
	for i, c := range clusters {
		if c.Provisioner != "kubernetes" {
			continue
		}
		if c.Default {
			chosenCluster = &clusters[i]
		}
		for _, p := range c.Pools {
			if p == pool {
				return &c, nil
			}
		}
	}
	if chosenCluster == nil {
		return nil, fmt.Errorf("no cluster found for pool %s", pool)
	}
	return chosenCluster, nil
}

func kubernetizeRuleGroups(groups []rulefmt.RuleGroup) []monitoringv1.RuleGroup {
	result := []monitoringv1.RuleGroup{}
	for _, group := range groups {
		groupInterval := group.Interval.String()
		if groupInterval == "0s" {
			groupInterval = ""
		}
		result = append(result, monitoringv1.RuleGroup{
			Name:     group.Name,
			Interval: groupInterval,
			Rules:    kubernetizeRules(group.Rules),
		})
	}
	return result
}

func kubernetizeRules(rules []rulefmt.RuleNode) []monitoringv1.Rule {
	result := []monitoringv1.Rule{}
	for _, rule := range rules {
		ruleFor := rule.For.String()
		if ruleFor == "0s" {
			ruleFor = ""
		}
		result = append(result, monitoringv1.Rule{
			Record:      rule.Record.Value,
			Alert:       rule.Alert.Value,
			Expr:        intstr.FromString(rule.Expr.Value),
			For:         ruleFor,
			Labels:      rule.Labels,
			Annotations: rule.Annotations,
		})
	}
	return result
}
