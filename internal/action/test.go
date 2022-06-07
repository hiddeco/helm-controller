package action

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
)

func Test(ctx context.Context, config *Configuration, obj *helmv2.HelmRelease) (*release.Release, error) {
	cfg, err := NewActionConfig(config)
	if err != nil {
		return nil, err
	}

	test := newTest(cfg, obj)
	return test.Run(obj.GetReleaseName())
}

func newTest(config *action.Configuration, obj *helmv2.HelmRelease) *action.ReleaseTesting {
	test := action.NewReleaseTesting(config)

	test.Namespace = obj.GetReleaseNamespace()
	test.Timeout = obj.Spec.GetTest().GetTimeout(obj.GetTimeout()).Duration

	return test
}
