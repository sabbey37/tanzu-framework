package utils

import (
	"fmt"
	"github.com/vmware-tanzu/tanzu-framework/addons/pinniped/config-controller/constants"
	tkgconstants "github.com/vmware-tanzu/tanzu-framework/pkg/v1/tkg/constants"
	corev1 "k8s.io/api/core/v1"
	clusterapiv1beta1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

// IsAddonType returns true if the secret is type `tkg.tanzu.vmware.com/addon`
func IsAddonType(secret *corev1.Secret) bool {
	return secret.Type == constants.TKGAddonType
}

// HasAddonLabel returns true if the `tkg.tanzu.vmware.com/addon` label matches the parameter we pass in
func HasAddonLabel(secret *corev1.Secret, label string) bool {
	return secret.Labels[constants.TKGAddonLabel] == label
}
// IsManagementCluster returns true if the cluster has the "cluster-role.tkg.tanzu.vmware.com/management" label
func IsManagementCluster(cluster clusterapiv1beta1.Cluster) bool {
	_, labelExists := cluster.GetLabels()[constants.TKGManagementLabel]
	return labelExists
}

// GetInfraProvider get infrastructure kind from cluster spec
func GetInfraProvider(cluster clusterapiv1beta1.Cluster) (string, error) {
	var infraProvider string

	infrastructureRef := cluster.Spec.InfrastructureRef
	if infrastructureRef == nil {
		return "", fmt.Errorf("cluster.Spec.InfrastructureRef is not set for cluster '%s", cluster.Name)
	}

	switch infrastructureRef.Kind {
	case tkgconstants.InfrastructureRefVSphere:
		infraProvider = tkgconstants.InfrastructureProviderVSphere
	case tkgconstants.InfrastructureRefAWS:
		infraProvider = tkgconstants.InfrastructureProviderAWS
	case tkgconstants.InfrastructureRefAzure:
		infraProvider = tkgconstants.InfrastructureProviderAzure
	case constants.InfrastructureRefDocker:
		infraProvider = tkgconstants.InfrastructureProviderDocker
	default:
		return "", fmt.Errorf("unknown cluster.Spec.InfrastructureRef.Kind is set for cluster '%s", cluster.Name)
	}

	return infraProvider, nil
}