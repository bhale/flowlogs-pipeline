package kubernetes

import (
	"testing"

	"github.com/netobserv/flowlogs-pipeline/pkg/api"
	"github.com/netobserv/flowlogs-pipeline/pkg/config"
	inf "github.com/netobserv/flowlogs-pipeline/pkg/pipeline/transform/kubernetes/informers"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var info = map[string]*inf.Info{
	"1.2.3.4": nil,
	"10.0.0.1": {
		ObjectMeta: v1.ObjectMeta{
			Name:      "pod-1",
			Namespace: "ns-1",
		},
		Type:     "Pod",
		HostName: "host-1",
		HostIP:   "100.0.0.1",
	},
	"10.0.0.2": {
		ObjectMeta: v1.ObjectMeta{
			Name:      "pod-2",
			Namespace: "ns-2",
		},
		Type:     "Pod",
		HostName: "host-2",
		HostIP:   "100.0.0.2",
	},
	"20.0.0.1": {
		ObjectMeta: v1.ObjectMeta{
			Name:      "service-1",
			Namespace: "ns-1",
		},
		Type: "Service",
	},
}

var nodes = map[string]*inf.Info{
	"host-1": {
		ObjectMeta: v1.ObjectMeta{
			Name: "host-1",
			Labels: map[string]string{
				nodeZoneLabelName: "us-east-1a",
			},
		},
		Type: "Node",
	},
	"host-2": {
		ObjectMeta: v1.ObjectMeta{
			Name: "host-2",
			Labels: map[string]string{
				nodeZoneLabelName: "us-east-1b",
			},
		},
		Type: "Node",
	},
}

var rules = api.NetworkTransformRules{
	{
		Type:   api.OpAddKubernetes,
		Input:  "SrcAddr",
		Output: "SrcK8s",
		Kubernetes: &api.K8sRule{
			AddZone: true,
		},
	},
	{
		Type:   api.OpAddKubernetes,
		Input:  "DstAddr",
		Output: "DstK8s",
		Kubernetes: &api.K8sRule{
			AddZone: true,
		},
	},
}

func TestEnrich(t *testing.T) {
	informers = inf.SetupStubs(info, nodes)

	// Pod to unknown
	entry := config.GenericMap{
		"SrcAddr": "10.0.0.1",    // pod-1
		"DstAddr": "42.42.42.42", // unknown
	}
	for _, r := range rules {
		Enrich(entry, r)
	}
	assert.Equal(t, config.GenericMap{
		"DstAddr":          "42.42.42.42",
		"SrcAddr":          "10.0.0.1",
		"SrcK8s_HostIP":    "100.0.0.1",
		"SrcK8s_HostName":  "host-1",
		"SrcK8s_Name":      "pod-1",
		"SrcK8s_Namespace": "ns-1",
		"SrcK8s_OwnerName": "",
		"SrcK8s_OwnerType": "",
		"SrcK8s_Type":      "Pod",
		"SrcK8s_Zone":      "us-east-1a",
	}, entry)

	// Pod to pod
	entry = config.GenericMap{
		"SrcAddr": "10.0.0.1", // pod-1
		"DstAddr": "10.0.0.2", // pod-2
	}
	for _, r := range rules {
		Enrich(entry, r)
	}
	assert.Equal(t, config.GenericMap{
		"DstAddr":          "10.0.0.2",
		"DstK8s_HostIP":    "100.0.0.2",
		"DstK8s_HostName":  "host-2",
		"DstK8s_Name":      "pod-2",
		"DstK8s_Namespace": "ns-2",
		"DstK8s_OwnerName": "",
		"DstK8s_OwnerType": "",
		"DstK8s_Type":      "Pod",
		"DstK8s_Zone":      "us-east-1b",
		"SrcAddr":          "10.0.0.1",
		"SrcK8s_HostIP":    "100.0.0.1",
		"SrcK8s_HostName":  "host-1",
		"SrcK8s_Name":      "pod-1",
		"SrcK8s_Namespace": "ns-1",
		"SrcK8s_OwnerName": "",
		"SrcK8s_OwnerType": "",
		"SrcK8s_Type":      "Pod",
		"SrcK8s_Zone":      "us-east-1a",
	}, entry)

	// Pod to service
	entry = config.GenericMap{
		"SrcAddr": "10.0.0.2", // pod-2
		"DstAddr": "20.0.0.1", // service-1
	}
	for _, r := range rules {
		Enrich(entry, r)
	}
	assert.Equal(t, config.GenericMap{
		"DstAddr":          "20.0.0.1",
		"DstK8s_Name":      "service-1",
		"DstK8s_Namespace": "ns-1",
		"DstK8s_OwnerName": "",
		"DstK8s_OwnerType": "",
		"DstK8s_Type":      "Service",
		"SrcAddr":          "10.0.0.2",
		"SrcK8s_HostIP":    "100.0.0.2",
		"SrcK8s_HostName":  "host-2",
		"SrcK8s_Name":      "pod-2",
		"SrcK8s_Namespace": "ns-2",
		"SrcK8s_OwnerName": "",
		"SrcK8s_OwnerType": "",
		"SrcK8s_Type":      "Pod",
		"SrcK8s_Zone":      "us-east-1b",
	}, entry)
}

var otelRules = api.NetworkTransformRules{
	{
		Type:     api.OpAddKubernetes,
		Input:    "source.ip",
		Output:   "source.",
		Assignee: "otel",
		Kubernetes: &api.K8sRule{
			AddZone: true,
		},
	},
	{
		Type:     api.OpAddKubernetes,
		Input:    "destination.ip",
		Output:   "destination.",
		Assignee: "otel",
		Kubernetes: &api.K8sRule{
			AddZone: true,
		},
	},
}

func TestEnrich_Otel(t *testing.T) {
	informers = inf.SetupStubs(info, nodes)

	// Pod to unknown
	entry := config.GenericMap{
		"source.ip":      "10.0.0.1",    // pod-1
		"destination.ip": "42.42.42.42", // unknown
	}
	for _, r := range otelRules {
		Enrich(entry, r)
	}
	assert.Equal(t, config.GenericMap{
		"destination.ip":            "42.42.42.42",
		"source.ip":                 "10.0.0.1",
		"source.k8s.host.ip":        "100.0.0.1",
		"source.k8s.host.name":      "host-1",
		"source.k8s.name":           "pod-1",
		"source.k8s.namespace.name": "ns-1",
		"source.k8s.pod.name":       "pod-1",
		"source.k8s.pod.uid":        types.UID(""),
		"source.k8s.owner.name":     "",
		"source.k8s.owner.type":     "",
		"source.k8s.type":           "Pod",
		"source.k8s.zone":           "us-east-1a",
	}, entry)

	// Pod to pod
	entry = config.GenericMap{
		"source.ip":      "10.0.0.1", // pod-1
		"destination.ip": "10.0.0.2", // pod-2
	}
	for _, r := range otelRules {
		Enrich(entry, r)
	}
	assert.Equal(t, config.GenericMap{
		"destination.ip":                 "10.0.0.2",
		"destination.k8s.host.ip":        "100.0.0.2",
		"destination.k8s.host.name":      "host-2",
		"destination.k8s.name":           "pod-2",
		"destination.k8s.namespace.name": "ns-2",
		"destination.k8s.pod.name":       "pod-2",
		"destination.k8s.pod.uid":        types.UID(""),
		"destination.k8s.owner.name":     "",
		"destination.k8s.owner.type":     "",
		"destination.k8s.type":           "Pod",
		"destination.k8s.zone":           "us-east-1b",
		"source.ip":                      "10.0.0.1",
		"source.k8s.host.ip":             "100.0.0.1",
		"source.k8s.host.name":           "host-1",
		"source.k8s.name":                "pod-1",
		"source.k8s.namespace.name":      "ns-1",
		"source.k8s.pod.name":            "pod-1",
		"source.k8s.pod.uid":             types.UID(""),
		"source.k8s.owner.name":          "",
		"source.k8s.owner.type":          "",
		"source.k8s.type":                "Pod",
		"source.k8s.zone":                "us-east-1a",
	}, entry)

	// Pod to service
	entry = config.GenericMap{
		"source.ip":      "10.0.0.2", // pod-2
		"destination.ip": "20.0.0.1", // service-1
	}
	for _, r := range otelRules {
		Enrich(entry, r)
	}
	assert.Equal(t, config.GenericMap{
		"destination.ip":                 "20.0.0.1",
		"destination.k8s.name":           "service-1",
		"destination.k8s.namespace.name": "ns-1",
		"destination.k8s.service.name":   "service-1",
		"destination.k8s.service.uid":    types.UID(""),
		"destination.k8s.owner.name":     "",
		"destination.k8s.owner.type":     "",
		"destination.k8s.type":           "Service",
		"source.ip":                      "10.0.0.2",
		"source.k8s.host.ip":             "100.0.0.2",
		"source.k8s.host.name":           "host-2",
		"source.k8s.name":                "pod-2",
		"source.k8s.namespace.name":      "ns-2",
		"source.k8s.pod.name":            "pod-2",
		"source.k8s.pod.uid":             types.UID(""),
		"source.k8s.owner.name":          "",
		"source.k8s.owner.type":          "",
		"source.k8s.type":                "Pod",
		"source.k8s.zone":                "us-east-1b",
	}, entry)
}

func TestEnrich_EmptyNamespace(t *testing.T) {
	informers = inf.SetupStubs(info, nodes)

	// We need to check that, whether it returns NotFound or just an empty namespace,
	// there is no map entry for that namespace (an empty-valued map entry is not valid)
	entry := config.GenericMap{
		"SrcAddr": "1.2.3.4", // would return an empty namespace
		"DstAddr": "3.2.1.0", // would return NotFound
	}

	for _, r := range rules {
		Enrich(entry, r)
	}

	assert.NotContains(t, entry, "SrcK8s_Namespace")
	assert.NotContains(t, entry, "DstK8s_Namespace")
}
