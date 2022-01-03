package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	monitoringv1 "github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring/v1"
	"github.com/prometheus/prometheus/pkg/rulefmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tsuru/go-tsuruclient/pkg/tsuru"
	"gopkg.in/yaml.v3"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	sigsk8sclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestEnsureCreation(t *testing.T) {
	setupFakeTsuruAPI()

	client := fake.NewClientBuilder().WithScheme(scheme).Build()

	svc := NewService("fake-tsuru-token", func(c *tsuru.Cluster) (sigsk8sclient.Client, error) {
		return client, nil
	})

	record := yaml.Node{}
	record.SetString("record")

	expr := yaml.Node{}
	expr.SetString("up > 0")

	err := svc.EnsurePrometheusRule("fake-pool", "fake-rule-name", rulefmt.RuleGroups{
		Groups: []rulefmt.RuleGroup{
			{
				Name: "fake-rule-group",
				Rules: []rulefmt.RuleNode{
					{
						Record: record,
						Expr:   expr,
					},
				},
			},
		},
	})

	require.NoError(t, err)

	currentPrometheusRule := &monitoringv1.PrometheusRule{}

	err = client.Get(context.TODO(), sigsk8sclient.ObjectKey{
		Name:      "fake-rule-name",
		Namespace: "tsuru-fake-pool",
	}, currentPrometheusRule)

	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"app.kubernetes.io/managed-by": "tsuru",
		"app.kubernetes.io/part-of":    "tsuru-prometheus-api",
	}, currentPrometheusRule.Annotations)
	assert.Equal(t, map[string]string{
		"tsuru.io/pool": "fake-pool",
	}, currentPrometheusRule.Labels)
	assert.Equal(t, monitoringv1.PrometheusRuleSpec{
		Groups: []monitoringv1.RuleGroup{
			{
				Name: "fake-rule-group",
				Rules: []monitoringv1.Rule{
					{
						Record: "record",
						Expr:   intstr.FromString("up > 0"),
					},
				},
			},
		},
	}, currentPrometheusRule.Spec)
}

func TestEnsureUpdate(t *testing.T) {
	setupFakeTsuruAPI()

	existingPrometheusRule := &monitoringv1.PrometheusRule{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "fake-rule-name",
			Namespace: "tsuru-fake-pool",
		},
		Spec: monitoringv1.PrometheusRuleSpec{
			Groups: []monitoringv1.RuleGroup{
				{
					Name: "fake-rule-group",
					Rules: []monitoringv1.Rule{
						{
							Record: "record",
							Expr:   intstr.FromString("tobereplaced > 0"),
						},
					},
				},
			},
		},
	}

	client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(existingPrometheusRule).Build()

	svc := NewService("fake-tsuru-token", func(c *tsuru.Cluster) (sigsk8sclient.Client, error) {
		return client, nil
	})

	record := yaml.Node{}
	record.SetString("record")

	expr := yaml.Node{}
	expr.SetString("up > 0")

	err := svc.EnsurePrometheusRule("fake-pool", "fake-rule-name", rulefmt.RuleGroups{
		Groups: []rulefmt.RuleGroup{
			{
				Name: "fake-rule-group",
				Rules: []rulefmt.RuleNode{
					{
						Record: record,
						Expr:   expr,
					},
				},
			},
		},
	})

	require.NoError(t, err)

	currentPrometheusRule := &monitoringv1.PrometheusRule{}

	err = client.Get(context.TODO(), sigsk8sclient.ObjectKey{
		Name:      "fake-rule-name",
		Namespace: "tsuru-fake-pool",
	}, currentPrometheusRule)

	require.NoError(t, err)
	assert.Equal(t, map[string]string{
		"app.kubernetes.io/managed-by": "tsuru",
		"app.kubernetes.io/part-of":    "tsuru-prometheus-api",
	}, currentPrometheusRule.Annotations)
	assert.Equal(t, map[string]string{
		"tsuru.io/pool": "fake-pool",
	}, currentPrometheusRule.Labels)
	assert.Equal(t, monitoringv1.PrometheusRuleSpec{
		Groups: []monitoringv1.RuleGroup{
			{
				Name: "fake-rule-group",
				Rules: []monitoringv1.Rule{
					{
						Record: "record",
						Expr:   intstr.FromString("up > 0"),
					},
				},
			},
		},
	}, currentPrometheusRule.Spec)
}

func setupFakeTsuruAPI() {
	fakeTsuruAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path

		if path == "/1.3/provisioner/clusters" {
			w.Header().Add("Content-Type", "application/json")
			json.NewEncoder(w).Encode([]tsuru.Cluster{
				{
					Name:        "cluster-1",
					Provisioner: "kubernetes",
					Addresses: []string{
						"http://cluste01",
					},
					Pools: []string{
						"fake-pool",
					},
				},
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))

	os.Setenv("TSURU_TARGET", fakeTsuruAPI.URL)
}
