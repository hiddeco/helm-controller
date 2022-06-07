package action

import (
	"context"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
)

func Uninstall(ctx context.Context, config *Configuration, obj *helmv2.HelmRelease) (*release.UninstallReleaseResponse, error) {
	cfg, err := NewActionConfig(config)
	if err != nil {
		return nil, err
	}
	uninstall := newUninstall(cfg, obj)
	return uninstall.Run(obj.GetReleaseName())
}

func newUninstall(config *action.Configuration, obj *helmv2.HelmRelease) *action.Uninstall {
	uninstall := action.NewUninstall(config)

	uninstall.Timeout = obj.Spec.GetUninstall().GetTimeout(obj.GetTimeout()).Duration
	uninstall.DisableHooks = obj.Spec.GetUninstall().DisableHooks
	uninstall.KeepHistory = obj.Spec.GetUninstall().KeepHistory
	uninstall.Wait = !obj.Spec.GetUninstall().DisableWait

	return uninstall
}
