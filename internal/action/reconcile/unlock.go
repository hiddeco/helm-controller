package reconcile

import (
	"context"
	"fmt"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/helm-controller/internal/action"
	intrelease "github.com/fluxcd/helm-controller/internal/release"
	"github.com/fluxcd/helm-controller/internal/storage"
	"helm.sh/helm/v3/pkg/release"
	ctrl "sigs.k8s.io/controller-runtime"
)

func unlock(ctx context.Context, r *Release) error {
	cfg, err := action.NewActionConfig(r.newConfig(action.DebugLogr(ctrl.LoggerFrom(ctx)), observeUnlock(r.object)))
	if err != nil {
		return err
	}
	rls, err := cfg.Releases.Last(r.object.GetReleaseName())
	if err != nil {
		return err
	}
	if status := rls.Info.Status; status.IsPending() {
		rls.SetStatus(release.StatusFailed, fmt.Sprintf("Release unlocked from stale '%s' state", status))
	}
	if err = cfg.Releases.Update(rls); err != nil {
		return err
	}
	return nil
}

func observeUnlock(obj *helmv2.HelmRelease) storage.ObserveFunc {
	return func(rls *release.Release) {
		if cur := obj.Status.Current; cur != nil {
			if cur.Version == rls.Version && cur.Name == rls.Name {
				obj.Status.Current = intrelease.ObservedToInfo(intrelease.ObserveRelease(rls))
			}
		}
	}
}
