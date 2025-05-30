// Copyright The Shipwright Contributors
//
// SPDX-License-Identifier: Apache-2.0

package utils

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/onsi/ginkgo/v2"
	tektonClient "github.com/tektoncd/pipeline/pkg/client/clientset/versioned"

	"k8s.io/client-go/kubernetes"
	// from https://github.com/kubernetes/client-go/issues/345
	_ "k8s.io/client-go/plugin/pkg/client/auth/oidc"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"

	buildClient "github.com/shipwright-io/build/pkg/client/clientset/versioned"
	test "github.com/shipwright-io/build/test/v1beta1_samples"
)

var (
	namespaceCounter int32
)

// TestBuild wraps all required clients to run integration
// tests and also the namespace and operator channel used
// per each test case
type TestBuild struct {
	// TODO: Adding specific field for polling here, interval and timeout
	// but I think we need a small refactoring to make them global for all
	// tests under /test dir
	Interval                time.Duration
	TimeOut                 time.Duration
	KubeConfig              *rest.Config
	Clientset               *kubernetes.Clientset
	Namespace               string
	BuildClientSet          *buildClient.Clientset
	PipelineClientSet       *tektonClient.Clientset
	ControllerRuntimeClient client.Client
	Catalog                 test.Catalog
	Context                 context.Context
}

// NewTestBuild returns an initialized instance of TestBuild
func NewTestBuild() (*TestBuild, error) {
	namespaceID := ginkgo.GinkgoParallelProcess()*200 + int(atomic.AddInt32(&namespaceCounter, 1))
	testNamespace := "test-build-" + strconv.Itoa(namespaceID)

	kubeConfig, restConfig, err := KubeConfig()
	if err != nil {
		return nil, err
	}

	// clientSet is required to communicate with our CRDs objects
	// see https://www.openshift.com/blog/kubernetes-deep-dive-code-generation-customresources
	buildClientSet, err := buildClient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	pipelineClientSet, err := tektonClient.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}

	controllerRuntimeClient, err := client.New(restConfig, client.Options{})
	if err != nil {
		return nil, err
	}

	return &TestBuild{
		// TODO: interval and timeout can be configured via ENV vars
		Interval:                time.Second * 3,
		TimeOut:                 time.Second * 300,
		KubeConfig:              restConfig,
		Clientset:               kubeConfig,
		Namespace:               testNamespace,
		BuildClientSet:          buildClientSet,
		PipelineClientSet:       pipelineClientSet,
		ControllerRuntimeClient: controllerRuntimeClient,
		Context:                 context.Background(),
	}, nil
}

// KubeConfig returns all required clients to speak with
// the k8s API
func KubeConfig() (*kubernetes.Clientset, *rest.Config, error) {
	location := os.Getenv("KUBECONFIG")
	if location == "" {
		location = filepath.Join(os.Getenv("HOME"), ".kube", "config")
	}

	config, err := clientcmd.BuildConfigFromFlags("", location)
	if err != nil {
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, nil, err
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, nil, err
	}

	return clientset, config, nil
}
