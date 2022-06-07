package reconcile

import (
	"context"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	intrelease "github.com/fluxcd/helm-controller/internal/release"
	"github.com/fluxcd/helm-controller/internal/storage"
	"helm.sh/helm/v3/pkg/release"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/fluxcd/helm-controller/internal/action"
)

func uninstall(ctx context.Context, r *Release) error {
	curRel := r.object.Status.Current
	logBuf := action.NewLogBuffer(action.DebugLogr(ctrl.LoggerFrom(ctx)), 10)
	_, err := action.Uninstall(ctx, r.newConfig(logBuf.Log, observeUninstall(r.object)), r.object)
	if err != nil {
		// Set object failure cond
		if r.object.Status.Current == curRel {
			return err
		}
		return nil
	}
	// Set test success cond
	return nil
}

func observeUninstall(obj *helmv2.HelmRelease) storage.ObserveFunc {
	return func(rls *release.Release) {
		if cur := obj.Status.Current; cur != nil {
			if cur.Version == rls.Version && cur.Name == rls.Name {
				obj.Status.Current = intrelease.ObservedToInfo(intrelease.ObserveRelease(rls))
			}
		}
	}
}
