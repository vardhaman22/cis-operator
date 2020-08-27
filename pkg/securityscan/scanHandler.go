package securityscan

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/rancher/wrangler/pkg/generic"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"strings"

	"github.com/blang/semver"
	v1 "github.com/rancher/cis-operator/pkg/apis/cis.cattle.io/v1"
	cisctlv1 "github.com/rancher/cis-operator/pkg/generated/controllers/cis.cattle.io/v1"
	ciscore "github.com/rancher/cis-operator/pkg/securityscan/core"
	cisjob "github.com/rancher/cis-operator/pkg/securityscan/job"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	kubeBenchJobManifest    = "{\r\n   \"apiVersion\": \"batch/v1\",\r\n   \"kind\": \"Job\",\r\n   \"metadata\": {\r\n      \"namespace\": \"cisscan-system\"\r\n   },\r\n   \"spec\": {\r\n      \"template\": {\r\n         \"metadata\": {\r\n            \"labels\": {\r\n               \"app\": \"kube-bench\"\r\n            }\r\n         },\r\n         \"spec\": {\r\n            \"hostPID\": true,\r\n            \"containers\": [\r\n               {\r\n                  \"name\": \"kube-bench\",\r\n                  \"image\": \"aquasec/kube-bench:latest\",\r\n                  \"command\": [\r\n                     \"kube-bench\"\r\n                  ],\r\n                  \"volumeMounts\": [\r\n                     {\r\n                        \"name\": \"var-lib-etcd\",\r\n                        \"mountPath\": \"/var/lib/etcd\",\r\n                        \"readOnly\": true\r\n                     },\r\n                     {\r\n                        \"name\": \"var-lib-kubelet\",\r\n                        \"mountPath\": \"/var/lib/kubelet\",\r\n                        \"readOnly\": true\r\n                     },\r\n                     {\r\n                        \"name\": \"etc-systemd\",\r\n                        \"mountPath\": \"/etc/systemd\",\r\n                        \"readOnly\": true\r\n                     },\r\n                     {\r\n                        \"name\": \"etc-kubernetes\",\r\n                        \"mountPath\": \"/etc/kubernetes\",\r\n                        \"readOnly\": true\r\n                     },\r\n                     {\r\n                        \"name\": \"usr-bin\",\r\n                        \"mountPath\": \"/usr/local/mount-from-host/bin\",\r\n                        \"readOnly\": true\r\n                     }\r\n                  ]\r\n               }\r\n            ],\r\n            \"restartPolicy\": \"Never\",\r\n            \"volumes\": [\r\n               {\r\n                  \"name\": \"var-lib-etcd\",\r\n                  \"hostPath\": {\r\n                     \"path\": \"/var/lib/etcd\"\r\n                  }\r\n               },\r\n               {\r\n                  \"name\": \"var-lib-kubelet\",\r\n                  \"hostPath\": {\r\n                     \"path\": \"/var/lib/kubelet\"\r\n                  }\r\n               },\r\n               {\r\n                  \"name\": \"etc-systemd\",\r\n                  \"hostPath\": {\r\n                     \"path\": \"/etc/systemd\"\r\n                  }\r\n               },\r\n               {\r\n                  \"name\": \"etc-kubernetes\",\r\n                  \"hostPath\": {\r\n                     \"path\": \"/etc/kubernetes\"\r\n                  }\r\n               },\r\n               {\r\n                  \"name\": \"usr-bin\",\r\n                  \"hostPath\": {\r\n                     \"path\": \"/usr/bin\"\r\n                  }\r\n               }\r\n            ]\r\n         }\r\n      }\r\n   }\r\n}"
	kubeBenchEKSJobManifest = "{\r\n   \"apiVersion\": \"batch/v1\",\r\n   \"kind\": \"Job\",\r\n   \"metadata\": {\r\n      \"name\": \"kube-bench\"\r\n   },\r\n   \"spec\": {\r\n      \"template\": {\r\n         \"spec\": {\r\n            \"hostPID\": true,\r\n            \"containers\": [\r\n               {\r\n                  \"name\": \"kube-bench\",\r\n                  \"image\": \"aquasec/kube-bench:latest\",\r\n                  \"command\": [\r\n                     \"kube-bench\",\r\n                     \"node\",\r\n                     \"--benchmark\",\r\n                     \"eks-1.0\"\r\n                  ],\r\n                  \"volumeMounts\": [\r\n                     {\r\n                        \"name\": \"var-lib-kubelet\",\r\n                        \"mountPath\": \"/var/lib/kubelet\",\r\n                        \"readOnly\": true\r\n                     },\r\n                     {\r\n                        \"name\": \"etc-systemd\",\r\n                        \"mountPath\": \"/etc/systemd\",\r\n                        \"readOnly\": true\r\n                     },\r\n                     {\r\n                        \"name\": \"etc-kubernetes\",\r\n                        \"mountPath\": \"/etc/kubernetes\",\r\n                        \"readOnly\": true\r\n                     }\r\n                  ]\r\n               }\r\n            ],\r\n            \"restartPolicy\": \"Never\",\r\n            \"volumes\": [\r\n               {\r\n                  \"name\": \"var-lib-kubelet\",\r\n                  \"hostPath\": {\r\n                     \"path\": \"/var/lib/kubelet\"\r\n                  }\r\n               },\r\n               {\r\n                  \"name\": \"etc-systemd\",\r\n                  \"hostPath\": {\r\n                     \"path\": \"/etc/systemd\"\r\n                  }\r\n               },\r\n               {\r\n                  \"name\": \"etc-kubernetes\",\r\n                  \"hostPath\": {\r\n                     \"path\": \"/etc/kubernetes\"\r\n                  }\r\n               }\r\n            ]\r\n         }\r\n      }\r\n   }\r\n}"
	kubeBenchGKEJobManifest = "{\r\n   \"apiVersion\": \"batch/v1\",\r\n   \"kind\": \"Job\",\r\n   \"metadata\": {\r\n      \"name\": \"kube-bench\"\r\n   },\r\n   \"spec\": {\r\n      \"template\": {\r\n         \"spec\": {\r\n            \"hostPID\": true,\r\n            \"containers\": [\r\n               {\r\n                  \"name\": \"kube-bench\",\r\n                  \"image\": \"aquasec/kube-bench:latest\",\r\n                  \"command\": [\r\n                     \"kube-bench\",\r\n                     \"--benchmark\",\r\n                     \"gke-1.0\",\r\n                     \"run\",\r\n                     \"--targets\",\r\n                     \"node,policies,managedservices\"\r\n                  ],\r\n                  \"volumeMounts\": [\r\n                     {\r\n                        \"name\": \"var-lib-kubelet\",\r\n                        \"mountPath\": \"/var/lib/kubelet\"\r\n                     },\r\n                     {\r\n                        \"name\": \"etc-systemd\",\r\n                        \"mountPath\": \"/etc/systemd\"\r\n                     },\r\n                     {\r\n                        \"name\": \"etc-kubernetes\",\r\n                        \"mountPath\": \"/etc/kubernetes\"\r\n                     }\r\n                  ]\r\n               }\r\n            ],\r\n            \"restartPolicy\": \"Never\",\r\n            \"volumes\": [\r\n               {\r\n                  \"name\": \"var-lib-kubelet\",\r\n                  \"hostPath\": {\r\n                     \"path\": \"/var/lib/kubelet\"\r\n                  }\r\n               },\r\n               {\r\n                  \"name\": \"etc-systemd\",\r\n                  \"hostPath\": {\r\n                     \"path\": \"/etc/systemd\"\r\n                  }\r\n               },\r\n               {\r\n                  \"name\": \"etc-kubernetes\",\r\n                  \"hostPath\": {\r\n                     \"path\": \"/etc/kubernetes\"\r\n                  }\r\n               }\r\n            ]\r\n         }\r\n      }\r\n   }\r\n}"
)

var SonobuoyMasterLabel = map[string]string{"run": "sonobuoy-master"}

func (c *Controller) handleClusterScans(ctx context.Context) error {
	scans := c.cisFactory.Cis().V1().ClusterScan()
	jobs := c.batchFactory.Batch().V1().Job()
	configmaps := c.coreFactory.Core().V1().ConfigMap()
	services := c.coreFactory.Core().V1().Service()

	cisctlv1.RegisterClusterScanGeneratingHandler(ctx, scans, c.apply.WithCacheTypes(configmaps, services).WithGVK(jobs.GroupVersionKind()).WithDynamicLookup().WithNoDelete(), "", c.Name,
		func(obj *v1.ClusterScan, status v1.ClusterScanStatus) (objects []runtime.Object, _ v1.ClusterScanStatus, _ error) {
			if obj == nil || obj.DeletionTimestamp != nil {
				return objects, status, nil
			}
			logrus.Infof("ClusterScan GENERATING HANDLER: scan=%s/%s@%s, %v, status=%+v", obj.Namespace, obj.Name, obj.Spec.ScanProfileName, obj.ResourceVersion, status.LastRunTimestamp)

			if obj.Status.LastRunTimestamp == "" && !v1.ClusterScanConditionCreated.IsTrue(obj) {
				if err := c.isRunnerPodPresent(); err != nil {
					v1.ClusterScanConditionReconciling.True(obj)
					return objects, obj.Status, fmt.Errorf("Retrying ClusterScan %v since got error: %v ", obj.Name, err)
				}

				//launch new on demand scan
				c.mu.Lock()
				defer c.mu.Unlock()

				profile, err := c.getClusterScanProfile(obj)
				if err != nil {
					v1.ClusterScanConditionStalled.True(obj)
					logrus.Errorf("Error validating ClusterScanProfile %v, error: %v", obj.Spec.ScanProfileName, err)
					return objects, obj.Status, nil
				}
				logrus.Infof("Launching a new on demand Job to run cis using profile %v", profile.Name)
				configmaps, err := ciscore.NewConfigMaps(obj, profile, c.Name, c.ImageConfig)
				if err != nil {
					v1.ClusterScanConditionReconciling.True(obj)
					return objects, obj.Status, fmt.Errorf("Error when creating ConfigMaps: %v", err)
				}
				service, err := ciscore.NewService(obj, profile, c.Name)
				if err != nil {
					v1.ClusterScanConditionReconciling.True(obj)
					return objects, obj.Status, fmt.Errorf("Error when creating Service: %v", err)
				}

				objects = append(objects, cisjob.New(obj, profile, c.Name, c.ImageConfig), configmaps[0], configmaps[1], configmaps[2], service)
				obj.Status.LastRunTimestamp = time.Now().String()
				v1.ClusterScanConditionCreated.True(obj)
				v1.ClusterScanConditionRunCompleted.Unknown(obj)

				return objects, obj.Status, nil
			}
			return objects, obj.Status, nil
		},
		&generic.GeneratingHandlerOptions{
			AllowClusterScoped: true,
		},
	)
	return nil
}
func (c *Controller) getClusterScanProfile(scan *v1.ClusterScan) (*v1.ClusterScanProfile, error) {
	var profileName string
	clusterscanprofiles := c.cisFactory.Cis().V1().ClusterScanProfile()

	if scan.Spec.ScanProfileName != "" {
		profileName = scan.Spec.ScanProfileName
	} else {
		//pick the default profile by checking the cluster provider
		profileName = c.getDefaultClusterScanProfile(c.ClusterProvider)
	}
	profile, err := clusterscanprofiles.Get(profileName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	err = c.validateClusterScanProfile(profile)
	if err != nil {
		return nil, err
	}
	return profile, nil
}

func (c Controller) getDefaultClusterScanProfile(clusterprovider string) string {
	var profileName string
	//load clusterScan
	switch clusterprovider {
	case v1.ClusterProviderRKE:
		profileName = "rke-profile-permissive"
	case v1.ClusterProviderEKS:
		profileName = "eks-profile"
	case v1.ClusterProviderGKE:
		profileName = "gke-profile"
	default:
		profileName = "cis-1.5-profile"
	}
	return profileName
}

func (c Controller) validateClusterScanProfile(profile *v1.ClusterScanProfile) error {

	// validate benchmarkVersion is valid and is applicable to this cluster
	clusterscanbmks := c.cisFactory.Cis().V1().ClusterScanBenchmark()
	benchmark, err := clusterscanbmks.Get(profile.Spec.BenchmarkVersion, metav1.GetOptions{})
	if err != nil {
		return err
	}

	// validate benchmark's provider matches the cluster
	if benchmark.Spec.ClusterProvider != "" {
		if !strings.EqualFold(benchmark.Spec.ClusterProvider, c.ClusterProvider) {
			return fmt.Errorf("ClusterProvider mismatch, ClusterScanProfile %v is not valid for this cluster's provider %v", profile.Name, c.ClusterProvider)
		}
	}

	// validate cluster's k8s version matches the benchmark's k8s version range
	clusterK8sToMatch, err := semver.Make(c.KubernetesVersion[1:])
	if err != nil {
		return fmt.Errorf("Cluster's k8sVersion is not sem-ver %s %v", c.KubernetesVersion, err)
	}
	var k8sRange string
	if benchmark.Spec.MinKubernetesVersion != "" {
		k8sRange = ">=" + benchmark.Spec.MinKubernetesVersion
	}
	if benchmark.Spec.MaxKubernetesVersion != "" {
		k8sRange = k8sRange + " <=" + benchmark.Spec.MaxKubernetesVersion
	}
	if k8sRange != "" {
		benchmarkK8sRange, err := semver.ParseRange(k8sRange)
		if err != nil {
			return fmt.Errorf("Range for Benchmark %s not sem-ver %v, error: %v", benchmark.Name, k8sRange, err)
		}
		if !benchmarkK8sRange(clusterK8sToMatch) {
			return fmt.Errorf("Kubernetes version mismatch, ClusterScanProfile %v is not valid for this cluster's K8s version %v", profile.Name, c.KubernetesVersion)
		}
	}

	return nil
}

func (c Controller) isRunnerPodPresent() error {
	v2Pods, err := c.listRunnerPods(v1.ClusterScanNS)
	if err != nil {
		return fmt.Errorf("error listing pods: %v", err)
	}
	if v2Pods != 0 {
		return fmt.Errorf("A rancher-cis-benchmark runner pod is already running")
	}

	v1Pods, err := c.listRunnerPods(v1.CISV1NS)
	if err != nil {
		return fmt.Errorf("error listing pods: %v", err)
	}
	if v1Pods != 0 {
		return fmt.Errorf("A CIS v1 rancher-cis-benchmark runner pod is already running")
	}

	return nil
}

func (c Controller) listRunnerPods(namespace string) (int, error) {
	pods := c.coreFactory.Core().V1().Pod()
	podList, err := pods.Cache().List(namespace, labels.Set(SonobuoyMasterLabel).AsSelector())
	if err != nil {
		return 0, fmt.Errorf("error listing pods: %v", err)
	}
	return len(podList), nil
}