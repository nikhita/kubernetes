/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package quota

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"k8s.io/api/core/v1"
	"k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/informers"
	kubeclientset "k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/record"
	internalclientset "k8s.io/kubernetes/pkg/client/clientset_generated/internalclientset"
	internalinformers "k8s.io/kubernetes/pkg/client/informers/informers_generated/internalversion"
	"k8s.io/kubernetes/pkg/controller"
	replicationcontroller "k8s.io/kubernetes/pkg/controller/replication"
	resourcequotacontroller "k8s.io/kubernetes/pkg/controller/resourcequota"
	"k8s.io/kubernetes/pkg/quota/generic"
	quotainstall "k8s.io/kubernetes/pkg/quota/install"
	"k8s.io/kubernetes/plugin/pkg/admission/resourcequota"
	resourcequotaapi "k8s.io/kubernetes/plugin/pkg/admission/resourcequota/apis/resourcequota"
	"k8s.io/kubernetes/test/integration/framework"
)

// 1.2 code gets:
// 	quota_test.go:95: Took 4.218619579s to scale up without quota
// 	quota_test.go:199: unexpected error: timed out waiting for the condition, ended with 342 pods (1 minute)
// 1.3+ code gets:
// 	quota_test.go:100: Took 4.196205966s to scale up without quota
// 	quota_test.go:115: Took 12.021640372s to scale up with quota
func TestQuota(t *testing.T) {
	// Set up a master
	h := &framework.MasterHolder{Initialized: make(chan struct{})}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		<-h.Initialized
		h.M.GenericAPIServer.Handler.ServeHTTP(w, req)
	}))

	admissionCh := make(chan struct{})
	clientset := clientset.NewForConfigOrDie(&restclient.Config{QPS: -1, Host: s.URL, ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}}})
	internalClientset := internalclientset.NewForConfigOrDie(&restclient.Config{QPS: -1, Host: s.URL, ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}}})
	config := &resourcequotaapi.Configuration{}
	admission, err := resourcequota.NewResourceQuota(config, 5, admissionCh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	admission.SetInternalKubeClientSet(internalClientset)
	internalInformers := internalinformers.NewSharedInformerFactory(internalClientset, controller.NoResyncPeriodFunc())
	admission.SetInternalKubeInformerFactory(internalInformers)
	qca := quotainstall.NewQuotaConfigurationForAdmission()
	admission.SetQuotaConfiguration(qca)
	defer close(admissionCh)

	masterConfig := framework.NewIntegrationTestMasterConfig()
	masterConfig.GenericConfig.AdmissionControl = admission
	_, _, closeFn := framework.RunAMasterUsingServer(masterConfig, s, h)
	defer closeFn()

	ns := framework.CreateTestingNamespace("quotaed", s, t)
	defer framework.DeleteTestingNamespace(ns, s, t)
	ns2 := framework.CreateTestingNamespace("non-quotaed", s, t)
	defer framework.DeleteTestingNamespace(ns2, s, t)

	controllerCh := make(chan struct{})
	defer close(controllerCh)

	informers := informers.NewSharedInformerFactory(clientset, controller.NoResyncPeriodFunc())
	rm := replicationcontroller.NewReplicationManager(
		informers.Core().V1().Pods(),
		informers.Core().V1().ReplicationControllers(),
		clientset,
		replicationcontroller.BurstReplicas,
	)
	rm.SetEventRecorder(&record.FakeRecorder{})
	go rm.Run(3, controllerCh)

	discoveryFunc := clientset.Discovery().ServerPreferredNamespacedResources
	listerFuncForResource := generic.ListerFuncForResourceFunc(informers.ForResource)
	qc := quotainstall.NewQuotaConfigurationForControllers(listerFuncForResource)
	informersStarted := make(chan struct{})
	resourceQuotaControllerOptions := &resourcequotacontroller.ResourceQuotaControllerOptions{
		QuotaClient:               clientset.Core(),
		ResourceQuotaInformer:     informers.Core().V1().ResourceQuotas(),
		ResyncPeriod:              controller.NoResyncPeriodFunc,
		InformerFactory:           informers,
		ReplenishmentResyncPeriod: controller.NoResyncPeriodFunc,
		DiscoveryFunc:             discoveryFunc,
		IgnoredResourcesFunc:      qc.IgnoredResources,
		InformersStarted:          informersStarted,
		Registry:                  generic.NewRegistry(qc.Evaluators()),
	}
	resourceQuotaController, err := resourcequotacontroller.NewResourceQuotaController(resourceQuotaControllerOptions)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	go resourceQuotaController.Run(2, controllerCh)

	// Periodically the quota controller to detect new resource types
	go resourceQuotaController.Sync(discoveryFunc, 30*time.Second, controllerCh)

	internalInformers.Start(controllerCh)
	informers.Start(controllerCh)
	close(informersStarted)

	startTime := time.Now()
	scale(t, ns2.Name, clientset)
	endTime := time.Now()
	t.Logf("Took %v to scale up without quota", endTime.Sub(startTime))

	quota := &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "quota",
			Namespace: ns.Name,
		},
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				v1.ResourcePods: resource.MustParse("1000"),
			},
		},
	}
	waitForQuota(t, quota, clientset)

	startTime = time.Now()
	scale(t, "quotaed", clientset)
	endTime = time.Now()
	t.Logf("Took %v to scale up with quota", endTime.Sub(startTime))
}

func waitForQuota(t *testing.T, quota *v1.ResourceQuota, clientset *kubeclientset.Clientset) {
	w, err := clientset.Core().ResourceQuotas(quota.Namespace).Watch(metav1.SingleObject(metav1.ObjectMeta{Name: quota.Name}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := clientset.Core().ResourceQuotas(quota.Namespace).Create(quota); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = watch.Until(1*time.Minute, w, func(event watch.Event) (bool, error) {
		switch event.Type {
		case watch.Modified:
		default:
			return false, nil
		}
		switch cast := event.Object.(type) {
		case *v1.ResourceQuota:
			if len(cast.Status.Hard) > 0 {
				return true, nil
			}
		}

		return false, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func scale(t *testing.T, namespace string, clientset *kubeclientset.Clientset) {
	target := int32(100)
	rc := &v1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: namespace,
		},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &target,
			Selector: map[string]string{"foo": "bar"},
			Template: &v1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"foo": "bar",
					},
				},
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "container",
							Image: "busybox",
						},
					},
				},
			},
		},
	}

	w, err := clientset.Core().ReplicationControllers(namespace).Watch(metav1.SingleObject(metav1.ObjectMeta{Name: rc.Name}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := clientset.Core().ReplicationControllers(namespace).Create(rc); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = watch.Until(3*time.Minute, w, func(event watch.Event) (bool, error) {
		switch event.Type {
		case watch.Modified:
		default:
			return false, nil
		}

		switch cast := event.Object.(type) {
		case *v1.ReplicationController:
			fmt.Printf("Found %v of %v replicas\n", int(cast.Status.Replicas), target)
			if cast.Status.Replicas == target {
				return true, nil
			}
		}

		return false, nil
	})
	if err != nil {
		pods, _ := clientset.Core().Pods(namespace).List(metav1.ListOptions{LabelSelector: labels.Everything().String(), FieldSelector: fields.Everything().String()})
		t.Fatalf("unexpected error: %v, ended with %v pods", err, len(pods.Items))
	}
}

func TestQuotaLimitedResourceDenial(t *testing.T) {
	// Set up a master
	h := &framework.MasterHolder{Initialized: make(chan struct{})}
	s := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		<-h.Initialized
		h.M.GenericAPIServer.Handler.ServeHTTP(w, req)
	}))

	admissionCh := make(chan struct{})
	clientset := clientset.NewForConfigOrDie(&restclient.Config{QPS: -1, Host: s.URL, ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}}})
	internalClientset := internalclientset.NewForConfigOrDie(&restclient.Config{QPS: -1, Host: s.URL, ContentConfig: restclient.ContentConfig{GroupVersion: &schema.GroupVersion{Group: "", Version: "v1"}}})

	// stop creation of a pod resource unless there is a quota
	config := &resourcequotaapi.Configuration{
		LimitedResources: []resourcequotaapi.LimitedResource{
			{
				Resource:      "pods",
				MatchContains: []string{"pods"},
			},
		},
	}
	qca := quotainstall.NewQuotaConfigurationForAdmission()
	admission, err := resourcequota.NewResourceQuota(config, 5, admissionCh)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	admission.SetInternalKubeClientSet(internalClientset)
	internalInformers := internalinformers.NewSharedInformerFactory(internalClientset, controller.NoResyncPeriodFunc())
	admission.SetInternalKubeInformerFactory(internalInformers)
	admission.SetQuotaConfiguration(qca)
	defer close(admissionCh)

	masterConfig := framework.NewIntegrationTestMasterConfig()
	masterConfig.GenericConfig.AdmissionControl = admission
	_, _, closeFn := framework.RunAMasterUsingServer(masterConfig, s, h)
	defer closeFn()

	ns := framework.CreateTestingNamespace("quota", s, t)
	defer framework.DeleteTestingNamespace(ns, s, t)

	controllerCh := make(chan struct{})
	defer close(controllerCh)

	informers := informers.NewSharedInformerFactory(clientset, controller.NoResyncPeriodFunc())
	rm := replicationcontroller.NewReplicationManager(
		informers.Core().V1().Pods(),
		informers.Core().V1().ReplicationControllers(),
		clientset,
		replicationcontroller.BurstReplicas,
	)
	rm.SetEventRecorder(&record.FakeRecorder{})
	go rm.Run(3, controllerCh)

	discoveryFunc := clientset.Discovery().ServerPreferredNamespacedResources
	listerFuncForResource := generic.ListerFuncForResourceFunc(informers.ForResource)
	qc := quotainstall.NewQuotaConfigurationForControllers(listerFuncForResource)
	informersStarted := make(chan struct{})
	resourceQuotaControllerOptions := &resourcequotacontroller.ResourceQuotaControllerOptions{
		QuotaClient:               clientset.Core(),
		ResourceQuotaInformer:     informers.Core().V1().ResourceQuotas(),
		ResyncPeriod:              controller.NoResyncPeriodFunc,
		InformerFactory:           informers,
		ReplenishmentResyncPeriod: controller.NoResyncPeriodFunc,
		DiscoveryFunc:             discoveryFunc,
		IgnoredResourcesFunc:      qc.IgnoredResources,
		InformersStarted:          informersStarted,
		Registry:                  generic.NewRegistry(qc.Evaluators()),
	}
	resourceQuotaController, err := resourcequotacontroller.NewResourceQuotaController(resourceQuotaControllerOptions)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	go resourceQuotaController.Run(2, controllerCh)

	// Periodically the quota controller to detect new resource types
	go resourceQuotaController.Sync(discoveryFunc, 30*time.Second, controllerCh)

	internalInformers.Start(controllerCh)
	informers.Start(controllerCh)
	close(informersStarted)

	// try to create a pod
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "foo",
			Namespace: ns.Name,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:  "container",
					Image: "busybox",
				},
			},
		},
	}
	if _, err := clientset.Core().Pods(ns.Name).Create(pod); err == nil {
		t.Fatalf("expected error for insufficient quota")
	}

	// now create a covering quota
	// note: limited resource does a matchContains, so we now have "pods" matching "pods" and "count/pods"
	quota := &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "quota",
			Namespace: ns.Name,
		},
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				v1.ResourcePods:               resource.MustParse("1000"),
				v1.ResourceName("count/pods"): resource.MustParse("1000"),
			},
		},
	}
	waitForQuota(t, quota, clientset)

	// attempt to create a new pod once the quota is propagated
	err = wait.PollImmediate(5*time.Second, time.Minute, func() (bool, error) {
		// retry until we succeed (to allow time for all changes to propagate)
		if _, err := clientset.Core().Pods(ns.Name).Create(pod); err == nil {
			return true, nil
		}
		return false, nil
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

/*
func mockDiscoveryFunc() ([]*metav1.APIResourceList, error) {
	return []*metav1.APIResourceList{}, nil
}

type testContext struct {
	tearDown           func()
	qc                 *resourcequotacontroller.ResourceQuotaController
	clientSet          kubeclientset.Interface
	apiExtensionClient apiextensionsclientset.Interface
	dynamicClient      dynamic.Interface
	startQC            func(workers int)
	// syncPeriod is how often the QC started with startQC will be resynced.
	syncPeriod time.Duration
}

func setup(t *testing.T, workerCount int) *testContext {
	return setupWithServer(t, kubeapiservertesting.StartTestServerOrDie(t, nil, nil, framework.SharedEtcd()), workerCount)
}

func setupWithServer(t *testing.T, result *kubeapiservertesting.TestServer, workerCount int) *testContext {
	clientSet, err := clientset.NewForConfig(result.ClientConfig)
	if err != nil {
		t.Fatalf("error creating clientset: %v", err)
	}

	// Helpful stuff for testing CRD.
	apiExtensionClient, err := apiextensionsclientset.NewForConfig(result.ClientConfig)
	if err != nil {
		t.Fatalf("error creating extension clientset: %v", err)
	}
	// CreateNewCustomResourceDefinition wants to use this namespace for verifying
	// namespace-scoped CRD creation.
	createNamespaceOrDie("aval", clientSet, t)

	discoveryClient := cacheddiscovery.NewMemCacheClient(clientSet.Discovery())
	restMapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	restMapper.Reset()

	quotableResources := resourcequotacontroller.GetQuotableResources(discoveryClient)
	config := *result.ClientConfig
	dynamicClient, err := dynamic.NewForConfig(&config)
	if err != nil {
		t.Fatalf("failed to create dynamicClient: %v", err)
	}
	sharedInformers := informers.NewSharedInformerFactory(clientSet, 0)
	alwaysStarted := make(chan struct{})
	close(alwaysStarted)

	gvr := schema.GroupVersionResource{Group: "example.com", Version: "v1", Resource: "foobars"}
	resourceQuotaControllerOptions := &resourcequotacontroller.ResourceQuotaControllerOptions{
		QuotaClient:               clientset.CoreV1(),
		ResourceQuotaInformer:     sharedInformers.Core().V1().ResourceQuotas(),
		ResyncPeriod:              controller.NoResyncPeriodFunc,
		ReplenishmentResyncPeriod: controller.NoResyncPeriodFunc,
		IgnoredResourcesFunc:      quotainstall.DefaultIgnoredResources(),
		DiscoveryFunc:             mockDiscoveryFunc,
		Registry:                  generic.NewRegistry(quotaConfiguration.Evaluators()),
		InformersStarted:          alwaysStarted,
		DynamicClient:             dynamicClient,
		Mapper:                    rm,
		QuotableResources:         twoResources,
		SharedInformerFactory:     sharedInformers,
	}

	gc, err := garbagecollector.NewGarbageCollector(
		dynamicClient,
		restMapper,
		deletableResources,
		garbagecollector.DefaultIgnoredResources(),
		sharedInformers,
		alwaysStarted,
	)
	if err != nil {
		t.Fatalf("failed to create garbage collector: %v", err)
	}

	stopCh := make(chan struct{})
	tearDown := func() {
		close(stopCh)
		result.TearDownFn()
	}
	syncPeriod := 5 * time.Second
	startGC := func(workers int) {
		go wait.Until(func() {
			// Resetting the REST mapper will also invalidate the underlying discovery
			// client. This is a leaky abstraction and assumes behavior about the REST
			// mapper, but we'll deal with it for now.
			restMapper.Reset()
		}, syncPeriod, stopCh)
		go gc.Run(workers, stopCh)
		go gc.Sync(clientSet.Discovery(), syncPeriod, stopCh)
	}

	if workerCount > 0 {
		startGC(workerCount)
	}

	return &testContext{
		tearDown:           tearDown,
		gc:                 gc,
		clientSet:          clientSet,
		apiExtensionClient: apiExtensionClient,
		dynamicClient:      dynamicClient,
		startGC:            startGC,
		syncPeriod:         syncPeriod,
	}
}

func createNamespaceOrDie(name string, c clientset.Interface, t *testing.T) *v1.Namespace {
	ns := &v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: name}}
	if _, err := c.CoreV1().Namespaces().Create(ns); err != nil {
		t.Fatalf("failed to create namespace: %v", err)
	}
	falseVar := false
	_, err := c.CoreV1().ServiceAccounts(ns.Name).Create(&v1.ServiceAccount{
		ObjectMeta:                   metav1.ObjectMeta{Name: "default"},
		AutomountServiceAccountToken: &falseVar,
	})
	if err != nil {
		t.Fatalf("failed to create service account: %v", err)
	}
	return ns
}

func createRandomCustomResourceDefinition(
	t *testing.T, apiExtensionClient apiextensionsclientset.Interface,
	dynamicClient dynamic.Interface,
	namespace string,
) (*apiextensionsv1beta1.CustomResourceDefinition, dynamic.ResourceInterface) {
	// Create a random custom resource definition and ensure it's available for
	// use.
	definition := apiextensionstestserver.NewRandomNameCustomResourceDefinition(apiextensionsv1beta1.NamespaceScoped)

	err := apiextensionstestserver.CreateNewCustomResourceDefinition(definition, apiExtensionClient, dynamicClient)
	if err != nil {
		t.Fatalf("failed to create CustomResourceDefinition: %v", err)
	}

	// Get a client for the custom resource.
	gvr := schema.GroupVersionResource{Group: definition.Spec.Group, Version: definition.Spec.Version, Resource: definition.Spec.Names.Plural}

	resourceClient := dynamicClient.Resource(gvr).Namespace(namespace)

	return definition, resourceClient
}

func newResourceQuota(name, namespace string) *v1.ResourceQuota {
	quota := &v1.ResourceQuota{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: namespace,
			Name:      name,
		},
		Spec: v1.ResourceQuotaSpec{
			Hard: v1.ResourceList{
				"count/foobar.example.com": resource.MustParse("4"),
			},
		},
		Status: v1.ResourceQuotaStatus{
			Hard: v1.ResourceList{
				"count/foobar.example.com": resource.MustParse("4"),
			},
		},
	}
	return quota
}

type testContext struct {
	tearDown           func()
	gc                 *garbagecollector.GarbageCollector
	clientSet          clientset.Interface
	apiExtensionClient apiextensionsclientset.Interface
	dynamicClient      dynamic.Interface
	startGC            func(workers int)
	// syncPeriod is how often the GC started with startGC will be resynced.
	syncPeriod time.Duration
}

// if workerCount > 0, will start the GC, otherwise it's up to the caller to Run() the GC.
func setup(t *testing.T, workerCount int) *testContext {
	return setupWithServer(t, kubeapiservertesting.StartTestServerOrDie(t, nil, nil, framework.SharedEtcd()), workerCount)
}

func setupWithServer(t *testing.T, result *kubeapiservertesting.TestServer, workerCount int) *testContext {
	clientSet, err := clientset.NewForConfig(result.ClientConfig)
	if err != nil {
		t.Fatalf("error creating clientset: %v", err)
	}

	// Helpful stuff for testing CRD.
	apiExtensionClient, err := apiextensionsclientset.NewForConfig(result.ClientConfig)
	if err != nil {
		t.Fatalf("error creating extension clientset: %v", err)
	}
	// CreateNewCustomResourceDefinition wants to use this namespace for verifying
	// namespace-scoped CRD creation.
	createNamespaceOrDie("aval", clientSet, t)

	discoveryClient := cacheddiscovery.NewMemCacheClient(clientSet.Discovery())
	restMapper := restmapper.NewDeferredDiscoveryRESTMapper(discoveryClient)
	restMapper.Reset()
	deletableResources := resourcequotacontroller.GetQuotableResources(discoveryClient)
	config := *result.ClientConfig
	dynamicClient, err := dynamic.NewForConfig(&config)
	if err != nil {
		t.Fatalf("failed to create dynamicClient: %v", err)
	}
	sharedInformers := informers.NewSharedInformerFactory(clientSet, 0)
	alwaysStarted := make(chan struct{})
	close(alwaysStarted)
	rq, err := resourcequotacontroller.NewResourceQuotaController(
		dynamicClient,
		restMapper,
	)
	gc, err := garbagecollector.NewGarbageCollector(
		dynamicClient,
		restMapper,
		deletableResources,
		garbagecollector.DefaultIgnoredResources(),
		sharedInformers,
		alwaysStarted,
	)
	if err != nil {
		t.Fatalf("failed to create garbage collector: %v", err)
	}

	stopCh := make(chan struct{})
	tearDown := func() {
		close(stopCh)
		result.TearDownFn()
	}
	syncPeriod := 5 * time.Second
	startGC := func(workers int) {
		go wait.Until(func() {
			// Resetting the REST mapper will also invalidate the underlying discovery
			// client. This is a leaky abstraction and assumes behavior about the REST
			// mapper, but we'll deal with it for now.
			restMapper.Reset()
		}, syncPeriod, stopCh)
		go gc.Run(workers, stopCh)
		go gc.Sync(clientSet.Discovery(), syncPeriod, stopCh)
	}

	if workerCount > 0 {
		startGC(workerCount)
	}

	return &testContext{
		tearDown:           tearDown,
		gc:                 gc,
		clientSet:          clientSet,
		apiExtensionClient: apiExtensionClient,
		dynamicClient:      dynamicClient,
		startGC:            startGC,
		syncPeriod:         syncPeriod,
	}
}

func newCRDInstance(definition *apiextensionsv1beta1.CustomResourceDefinition, namespace, name string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       definition.Spec.Names.Kind,
			"apiVersion": definition.Spec.Group + "/" + definition.Spec.Version,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": namespace,
			},
		},
	}
}

func TestQuotaForCustomResource(t *testing.T) {
	ctx := setup(t, 5)
	defer ctx.tearDown()

	clientSet, apiExtensionClient, dynamicClient := ctx.clientSet, ctx.apiExtensionClient, ctx.dynamicClient

	ns := createNamespaceOrDie("crd-quota", clientSet, t)

	quotaClient := clientSet.CoreV1().ResourceQuotas(ns.Name)

	definition, resourceClient := createRandomCustomResourceDefinition(t, apiExtensionClient, dynamicClient, ns.Name)

	// Create a custom resource
	owner := newCRDInstance(definition, ns.Name, names.SimpleNameGenerator.GenerateName("foo1"))
	owner, err := resourceClient.Create(owner)
	if err != nil {
		t.Fatalf("failed to create owner resource %q: %v", owner.GetName(), err)
	}
	t.Logf("created a custom resource %q", owner.GetName())

	fooQuota := newResourceQuota("fooQuota", ns.Name)
	createdQuota, err := quotaClient.Create(fooQuota)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(fooQuota)

}
*/
