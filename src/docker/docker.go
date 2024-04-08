package docker

import (
	"fmt"

	"github.com/pulumi/pulumi-command/sdk/go/command/local"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func BridgeClusters(ctx *pulumi.Context, clusterOne, clusterTwo string, finalCluster pulumi.Resource) error {
	_, err := local.NewCommand(ctx, fmt.Sprintf("bridge-%s-%s", clusterOne, clusterTwo), &local.CommandArgs{
		Create: pulumi.Sprintf("docker network connect k3d-%s k3d-%s-server-0 && docker network connect k3d-%s k3d-%s-server-0", clusterOne, clusterTwo, clusterTwo, clusterOne),
	}, pulumi.DependsOn([]pulumi.Resource{finalCluster}))
	if err != nil {
		return err
	}
	if err = ctx.Log.Info(fmt.Sprintf("Bridged cluster %s => %s networks", clusterOne, clusterTwo), nil); err != nil {
		return err
	}
	return nil
}
