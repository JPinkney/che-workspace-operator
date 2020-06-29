//
// Copyright (c) 2019-2020 Red Hat, Inc.
// This program and the accompanying materials are made
// available under the terms of the Eclipse Public License 2.0
// which is available at https://www.eclipse.org/legal/epl-2.0/
//
// SPDX-License-Identifier: EPL-2.0
//
// Contributors:
//   Red Hat, Inc. - initial API and implementation
//

package main

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"log"

	devconfig "github.com/devfile/devworkspace-operator/pkg/config"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)


var (
	OP_GVR = schema.GroupVersionResource{
		Group: "operators.coreos.com",
		Version: "v1",
		Resource: "subscription",
	}
)

func main() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal("Failed when attempting to retrieve in cluster config: ", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal("Failed when attempting to retrieve in cluster config: ", err)
	}

	dynamic, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Fatal("Failed when attempting to retrieve in cluster config: ", err)
	}

	crdClient := dynamic.Resource(OP_GVR)
	crd, errcrd := crdClient.Get("eclipse-che", metav1.GetOptions{})
	if errcrd != nil {
		log.Fatal("Failed when attempting to get dynamic resource: ", err)
	}
	log.Println(crd.GetName())
	log.Println(crd)


	crdConfig := *config
	crdConfig.ContentConfig.GroupVersion = &schema.GroupVersion{Group: "operators.coreos.com", Version: "v1"}
	crdConfig.APIPath = "/apis"
	crdConfig.NegotiatedSerializer = serializer.NewCodecFactory(scheme.Scheme)
	crdConfig.UserAgent = rest.DefaultKubernetesUserAgent()

	exampleRestClient, err := rest.UnversionedRESTClientFor(&crdConfig)
	if err != nil {
		log.Fatal("Failed trying to get unversioned rest client: ", err)
	}

	b, err := exampleRestClient.Get().Resource("pods").Do().Raw()
	if err != nil {
		log.Fatal("Failed when trying to get subscription: ", err)
	}
	log.Println(string(b))

	/**
	 * My guess is that when the installplan is deleted or the crds are deleted then its not upgrading?
	 * We need to make sure that it doesn't delete all the workspaces when upgrading but does when its actually uninstalling
	 *
	 */

	clientset.AppsV1()

	log.Println("Attempting to delete all the deployments with label 2: " + devconfig.WorkspaceIDLabel)
	deployments, err := clientset.AppsV1().Deployments("che-workspace-controller").List(metav1.ListOptions{LabelSelector: "controller.devfile.io/workspace_id"})
	if err != nil {
		log.Fatal("Failed when attempting to delete all deployments with label: "+devconfig.WorkspaceIDLabel, err)
	}
	for _, s := range deployments.Items {
		log.Println("Attempting to delete pod: " + s.Name)
		clientset.AppsV1().Deployments("che-workspace-controller").Delete(s.Name, &metav1.DeleteOptions{})
	}
	log.Println("Deleted all the deployments with label: " + devconfig.WorkspaceIDLabel)
}
