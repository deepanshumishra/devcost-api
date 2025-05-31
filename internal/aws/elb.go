package aws

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	// "github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/deepanshumishra/devcost-api/internal/config"
	"github.com/deepanshumishra/devcost-api/internal/models"
)

// ListUnusedLoadBalancers identifies load balancers with no registered targets.
func ListUnusedLoadBalancers(ctx context.Context, cfg *config.Config) ([]models.UnusedResource, error) {
	client := elasticloadbalancingv2.NewFromConfig(cfg.AWSConfig)
	var unusedLBs []models.UnusedResource

	// List load balancers
	lbInput := &elasticloadbalancingv2.DescribeLoadBalancersInput{}
	lbResult, err := client.DescribeLoadBalancers(ctx, lbInput)
	if err != nil {
		return nil, err
	}

	for _, lb := range lbResult.LoadBalancers {
		// Get target groups for the load balancer
		tgInput := &elasticloadbalancingv2.DescribeTargetGroupsInput{
			LoadBalancerArn: lb.LoadBalancerArn,
		}
		tgResult, err := client.DescribeTargetGroups(ctx, tgInput)
		if err != nil {
			return nil, err
		}

		hasTargets := false
		for _, tg := range tgResult.TargetGroups {
			// Check target health
			healthInput := &elasticloadbalancingv2.DescribeTargetHealthInput{
				TargetGroupArn: tg.TargetGroupArn,
			}
			healthResult, err := client.DescribeTargetHealth(ctx, healthInput)
			if err != nil {
				return nil, err
			}
			if len(healthResult.TargetHealthDescriptions) > 0 {
				hasTargets = true
				break
			}
		}

		if !hasTargets {
			unusedLBs = append(unusedLBs, models.UnusedResource{
				ResourceType: "elasticloadbalancing:loadbalancer",
				ResourceID:   aws.ToString(lb.LoadBalancerArn),
				Reason:       "No registered targets",
			})
		}
	}

	return unusedLBs, nil
}