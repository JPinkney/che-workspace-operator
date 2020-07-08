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

package prestop

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func RemoveExistingCustomResources() {
	config, err := rest.InClusterConfig()
	if err != nil {
		log.Error(err, "Failed when attempting to retrieve in cluster config")
	}

	name := "web-terminal"
	namespace := "openshift-operators"
	dynamic, err := dynamic.NewForConfig(config)
	if err != nil {
		log.Error(err, "Failed when trying to get the dynamic config")
	}
	isAvailable, err := isSubscriptionAvailable(dynamic, name, namespace)
	if err != nil {
		log.Error(err, "Failed when trying to find out if the subscription is available")
	}

	if isAvailable {
		log.Info("Subscription is available")
		err = deleteCustomResources(dynamic, namespace)
		if err != nil {
			log.Error(err, "Failed when trying to delete custom resources")
		}
	} else {
		log.Info("Subscription is unavailable")
	}
}

/**
 * Check if the subscription to web-terminal is available.
 * If an error is found return that it's available so that custom resources don't get cleaned up.
 * Only return false when the get request to the resource results in an isNotFound error
 */
func isSubscriptionAvailable(config dynamic.Interface, name, namespace string) (bool, error) {
	OpGvr := schema.GroupVersionResource{
		Group: "operators.coreos.com",
		Version: "v1alpha1",
		Resource: "subscriptions",
	}

	crdClient := config.Resource(OpGvr)


	_, err := crdClient.Namespace(namespace).List(metav1.ListOptions{})
	log.Error(err, "")

	_, err = crdClient.Namespace(namespace).Get(name, metav1.GetOptions{})
	if err != nil {
		if apierrors.IsNotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

/**
 * Clean up all the devworkspace custom resources
 */
func deleteCustomResources(config dynamic.Interface, namespace string) error {
	OpGvr := schema.GroupVersionResource{
		Group: "workspace.devfile.io",
		Version: "v1alpha1",
		Resource: "devworkspaces",
	}

	crdClient := config.Resource(OpGvr)
	devworkspaces, err := crdClient.Namespace(namespace).List(metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, devworkspace := range devworkspaces.Items {
		err = crdClient.Namespace(namespace).Delete(devworkspace.GetName(), &metav1.DeleteOptions{})
		if err != nil {
			return err
		}
	}
	log.Info("Deleted all the devworkspaces")
	return nil
}
