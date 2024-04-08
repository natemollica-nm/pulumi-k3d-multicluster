package main

import (
	"os"
	"path/filepath"

	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"

	"pulumi-k3d-multicluster/src/clusters"
	"pulumi-k3d-multicluster/src/docker"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		var err error
		var k8sProvider *kubernetes.Provider

		// Initialize configuration
		conf := config.New(ctx, "")
		singleCluster := conf.GetBool("singleCluster")

		// Create local registry
		var registry pulumi.Resource
		registryName := "registry.localhost"
		registryPort := 5000
		registry, err = clusters.CreateRegistry(ctx, registryName, registryPort)
		if err != nil {
			return err
		}

		// Define clusters
		clustersMap := map[string]string{"c1": "us-east-1"}
		if !singleCluster {
			clustersMap = map[string]string{
				"c1": "us-east-1",
				"c2": "us-east-2",
				"c3": "us-west-1",
				"c4": "us-west-2",
			}
		}

		// Assume kubeconfig is saved at a standard path or set KUBECONFIG environment variable
		kubeconfigPath := filepath.Join(os.Getenv("HOME"), ".kube", "config")
		k8sProvider, err = kubernetes.NewProvider(ctx, "k3d-pulumi-provider", &kubernetes.ProviderArgs{
			Kubeconfig: pulumi.String(kubeconfigPath),
		})
		if err != nil {
			return err
		}

		// Create/Delete clusters
		var finalCluster pulumi.Resource
		for clusterName, region := range clustersMap {
			var cluster pulumi.Resource
			// Simplified cluster creation
			cluster, err = clusters.CreateCluster(ctx, clusterName, registryPort, registry)
			if err != nil {
				return err
			}
			// Apply labels with a dependency on the cluster creation
			_, err = clusters.CreateNodeLabel(ctx, clusterName, region, cluster, k8sProvider)
			if err != nil {
				return err
			}

			if clusterName == "c4" {
				finalCluster = cluster
			}
		}
		// Interconnect multicluster environment
		if !singleCluster {
			clusterPairs := [][2]string{
				{"c1", "c2"},
				{"c1", "c3"},
				{"c1", "c4"},
				{"c2", "c3"},
				{"c2", "c4"},
				{"c3", "c4"},
			}

			for _, pair := range clusterPairs {
				err = docker.BridgeClusters(ctx, pair[0], pair[1], finalCluster)
				if err != nil {
					return err
				}
			}
		}
		return nil
	})
}
