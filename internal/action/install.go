package action

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/release"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/helm-controller/internal/postrender"
)

func Install(ctx context.Context, config *Configuration, obj *helmv2.HelmRelease, chrt *chart.Chart, vals chartutil.Values) (*release.Release, error) {
	cfg, err := NewActionConfig(config)
	if err != nil {
		return nil, err
	}

	install, err := newInstall(cfg, obj)
	if err != nil {
		return nil, err
	}

	if err := installCRDs(cfg, obj, chrt, install); err != nil {
		return nil, err
	}

	return install.RunWithContext(ctx, chrt, vals.AsMap())
}

func newInstall(config *action.Configuration, obj *helmv2.HelmRelease) (*action.Install, error) {
	install := action.NewInstall(config)

	install.ReleaseName = obj.GetReleaseName()
	install.Namespace = obj.GetReleaseNamespace()
	install.Timeout = obj.Spec.GetInstall().GetTimeout(obj.GetTimeout()).Duration
	install.Wait = !obj.Spec.GetInstall().DisableWait
	install.WaitForJobs = !obj.Spec.GetInstall().DisableWaitForJobs
	install.DisableHooks = obj.Spec.GetInstall().DisableHooks
	install.DisableOpenAPIValidation = obj.Spec.GetInstall().DisableOpenAPIValidation
	install.Replace = obj.Spec.GetInstall().Replace
	install.Devel = true

	if obj.Spec.TargetNamespace != "" {
		install.CreateNamespace = obj.Spec.GetInstall().CreateNamespace
	}

	renderer, err := postrender.BuildPostRenderers(obj)
	if err != nil {
		return nil, err
	}
	install.PostRenderer = renderer

	return install, nil
}

func installCRDs(config *action.Configuration, obj *helmv2.HelmRelease, chrt *chart.Chart, install *action.Install) error {
	policy, err := crdPolicyOrDefault(obj.Spec.GetInstall().CRDs)
	if err != nil {
		return err
	}
	if policy == helmv2.Skip || policy == helmv2.CreateReplace {
		install.SkipCRDs = true
	}
	if policy == helmv2.CreateReplace {
		crds := chrt.CRDObjects()
		if len(crds) > 0 {
			if err := applyCRDs(config, policy, chrt); err != nil {
				return err
			}
		}
	}
	return nil
}
