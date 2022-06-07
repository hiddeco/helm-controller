package action

import (
	"context"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"helm.sh/helm/v3/pkg/action"
)

func Rollback(ctx context.Context, config *Configuration, obj *helmv2.HelmRelease) error {
	cfg, err := NewActionConfig(config)
	if err != nil {
		return err
	}
	rollback := newRollback(cfg, obj)
	return rollback.Run(obj.GetReleaseName())
}

func newRollback(config *action.Configuration, rel *helmv2.HelmRelease) *action.Rollback {
	rollback := action.NewRollback(config)

	rollback.Timeout = rel.Spec.GetRollback().GetTimeout(rel.GetTimeout()).Duration
	rollback.Wait = !rel.Spec.GetRollback().DisableWait
	rollback.WaitForJobs = !rel.Spec.GetRollback().DisableWaitForJobs
	rollback.DisableHooks = rel.Spec.GetRollback().DisableHooks
	rollback.Force = rel.Spec.GetRollback().Force
	rollback.Recreate = rel.Spec.GetRollback().Recreate
	rollback.CleanupOnFail = rel.Spec.GetRollback().CleanupOnFail

	if rel.Status.Previous != nil {
		rollback.Version = rel.Status.Previous.Version
	}

	return rollback
}
