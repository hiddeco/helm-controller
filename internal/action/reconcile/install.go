package reconcile

import (
	"context"
	ctrl "sigs.k8s.io/controller-runtime"

	"helm.sh/helm/v3/pkg/release"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/helm-controller/internal/action"
	intrelease "github.com/fluxcd/helm-controller/internal/release"
	"github.com/fluxcd/helm-controller/internal/storage"
)

func install(ctx context.Context, r *Release) error {
	curRel := r.object.Status.Current
	logBuf := action.NewLogBuffer(action.DebugLogr(ctrl.LoggerFrom(ctx)), 10)
	_, err := action.Install(ctx, r.newConfig(logBuf.Log, observeInstall(r.object)), r.object, r.chart, r.values)
	if err != nil {
		// Set object failure cond
		if r.object.Status.Current == curRel {
			return err
		}
		return nil
	}
	// Set object success cond
	return nil
}

func observeInstall(obj *helmv2.HelmRelease) storage.ObserveFunc {
	return func(rls *release.Release) {
		cur := obj.Status.Current
		if cur != nil && cur.Name == rls.Name && cur.Version < rls.Version {
			// Add current to previous when we observe the first write of a
			// newer release.
			obj.Status.Previous = obj.Status.Current
		}
		if cur == nil || obj.Status.Current.Version <= rls.Version {
			// Overwrite current with newer release, or update it.
			obj.Status.Current = intrelease.ObservedToInfo(intrelease.ObserveRelease(rls))
		}
		if prev := obj.Status.Previous; prev != nil && obj.Status.Previous.Version == rls.Version {
			// Write latest state of previous (e.g. status updates) to status.
			obj.Status.Previous = intrelease.ObservedToInfo(intrelease.ObserveRelease(rls))
		}
	}
}
