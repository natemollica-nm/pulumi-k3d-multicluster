package clusters

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func CreateRegistry(ctx *pulumi.Context, registryName string, registryPort int) (*local.Command, error) {
	return local.NewCommand(ctx, "createRegistry", &local.CommandArgs{
		Create: pulumi.Sprintf("k3d registry create --port %d %s", registryPort, registryName),
		Delete: pulumi.String("k3d registry delete " + registryName),
	})
}

func CreateCluster(ctx *pulumi.Context, clusterName string, registryPort int, registry pulumi.Resource) (*local.Command, error) {
	return local.NewCommand(ctx, "cluster-manage-"+clusterName, &local.CommandArgs{
		Create: pulumi.String(fmt.Sprintf("k3d cluster create %s --registry-use k3d-registry.localhost:%d", clusterName, registryPort)),
		Delete: pulumi.String(fmt.Sprintf("k3d cluster delete %s", clusterName)),
	}, pulumi.DependsOn([]pulumi.Resource{registry}))
}

func CreateNodeLabel(ctx *pulumi.Context, clusterName string, region string, cluster pulumi.Resource, k8sProvider *kubernetes.Provider) (*local.Command, error) {
	return local.NewCommand(ctx, "labelNode-"+clusterName, &local.CommandArgs{
		Create: pulumi.String(fmt.Sprintf("kubectl --context=k3d-%s label nodes --overwrite=true k3d-%s-server-0 topology.kubernetes.io/region=%s", clusterName, clusterName, region)),
	}, pulumi.DependsOn([]pulumi.Resource{cluster}), pulumi.Provider(k8sProvider))
}
