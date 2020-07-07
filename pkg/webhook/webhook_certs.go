package webhook

import (
	"context"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	crclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func SetupWebhookCerts(client crclient.Client, ctx context.Context, namespace string) error {
	log.Info("Attempting to create the secure service")
	err := createSecureService(client, ctx, namespace)
	if err != nil {
		log.Info("Failed creating the secure service")
		return err
	}

	log.Info("Attempting to create the secure configmap")
	err = createConfigMap(client, ctx, namespace)
	if err != nil {
		log.Info("Failed creating the secure configmap")
		return err
	}
	return nil
}

func createSecureService(client crclient.Client, ctx context.Context, namespace string) error {
	label := map[string]string{
		"app.kubernetes.io/name": "devworkspace-controller",
		"app.kubernetes.io/part-of": "devworkspace-operator",
	}

	port := int32(443)
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      SecureServiceName,
			Namespace: namespace,
			Labels:    label,
			Annotations: map[string]string{
				"service.beta.openshift.io/serving-cert-secret-name": CertSecretName,
			},
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Port:       port,
					Protocol:   "TCP",
					TargetPort: intstr.FromString(WebhookServerName),
				},
			},
			Selector: label,
		},
	}

	if err := client.Create(ctx, service); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		existingCfg, err := getClusterService(ctx, namespace, client)
		if err != nil {
			return err
		}

		// Cannot naively copy spec, as clusterIP is unmodifiable
		clusterIP := existingCfg.Spec.ClusterIP
		service.Spec = existingCfg.Spec
		service.Spec.ClusterIP = clusterIP
		service.ResourceVersion = existingCfg.ResourceVersion

		err = client.Update(ctx, service)
		if err != nil {
			return err
		}
		log.Info("Updating webhook server secure cert service")
	} else {
		log.Info("Updating webhook server secure cert service")
	}
	return nil
}

func createConfigMap(client crclient.Client, ctx context.Context, namespace string) error {
	configMapData := make(map[string]string, 0)
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      CertConfigMapName,
			Namespace: namespace,
			Annotations: map[string]string{
				"service.beta.openshift.io/inject-cabundle": "true",
			},
		},
		Data: configMapData,
	}

	if err := client.Create(ctx, configMap); err != nil {
		if !apierrors.IsAlreadyExists(err) {
			return err
		}
		existingCfg, err := getClusterConfigMap(ctx, namespace, client)
		if err != nil {
			return err
		}

		configMap.ResourceVersion = existingCfg.ResourceVersion
		err = client.Update(ctx, configMap)
		if err != nil {
			return err
		}
		log.Info("Updating webhook server secure cert configmap")
	} else {
		log.Info("Creating webhook server secure cert configmap")
	}

	return nil
}

func getClusterConfigMap(ctx context.Context, namespace string, client crclient.Client) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	namespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      CertConfigMapName,
	}
	err := client.Get(ctx, namespacedName, configMap)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return configMap, nil
}

func getClusterService(ctx context.Context, namespace string, client crclient.Client) (*corev1.Service, error) {
	service := &corev1.Service{}
	namespacedName := types.NamespacedName{
		Namespace: namespace,
		Name:      SecureServiceName,
	}
	err := client.Get(ctx, namespacedName, service)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, err
	}
	return service, nil
}
