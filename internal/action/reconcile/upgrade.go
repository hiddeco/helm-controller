package reconcile

import (
	"context"
	v2 "github.com/fluxcd/helm-controller/api/v2beta1"
	intrelease "github.com/fluxcd/helm-controller/internal/release"
	"github.com/fluxcd/helm-controller/internal/storage"
	"helm.sh/helm/v3/pkg/release"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/fluxcd/helm-controller/internal/action"
)

func upgrade(ctx context.Context, r *Release) error {
	curRel := r.object.Status.Current
	logBuf := action.NewLogBuffer(action.DebugLogr(ctrl.LoggerFrom(ctx)), 10)
	_, err := action.Upgrade(ctx, r.newConfig(logBuf.Log, observeUpgrade(r.object)), r.object, r.chart, r.values)
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

func observeUpgrade(obj *v2.HelmRelease) storage.ObserveFunc {
	return func(rls *release.Release) {
		if cur := obj.Status.Current; cur != nil {
			if cur.Name == rls.Name && cur.Version < rls.Version {
				obj.Status.Previous = obj.Status.Current
			}
		}
		if obj.Status.Current == nil || obj.Status.Current.Version <= rls.Version {
			obj.Status.Current = intrelease.ObservedToInfo(intrelease.ObserveRelease(rls))
		}
		if obj.Status.Previous != nil && obj.Status.Previous.Version == rls.Version {
			obj.Status.Previous = intrelease.ObservedToInfo(intrelease.ObserveRelease(rls))
		}
	}
}
