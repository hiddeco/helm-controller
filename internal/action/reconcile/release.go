package reconcile

import (
	"context"
	"errors"
	intrelease "github.com/fluxcd/helm-controller/internal/release"
	"github.com/fluxcd/helm-controller/internal/storage"
	helmaction "helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/release"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/helm-controller/internal/action"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
)

var errAbort = errors.New("abort")

type releaseAction func(ctx context.Context, release *Release) error

type Release struct {
	getter genericclioptions.RESTClientGetter
	object *helmv2.HelmRelease
	chart  *chart.Chart
	values chartutil.Values
}

func (r *Release) Reconcile(ctx context.Context) error {
	for {
		a, err := nextAction(r)
		if errors.Is(err, errAbort) {
			return nil
		}
		if err = a(ctx, r); err != nil {
			return err
		}
	}
}

func (r *Release) newConfig(log helmaction.DebugLog, obs ...storage.ObserveFunc) *action.Configuration {
	return &action.Configuration{
		Getter:           r.getter,
		StorageNamespace: r.object.GetStorageNamespace(),
		Observers:        obs,
		Log:              log,
	}
}

// nextAction determines the action that should be performed for the release
// by verifying the integrity of the Helm storage and further state of the
// release, and comparing the Release.chart and Release.values to the active
// deployment.
func nextAction(r *Release) (releaseAction, error) {
	rls, err := action.VerifyStorage(r.newConfig(nil), r.object)
	if err != nil {
		switch err {
		case action.ErrReleaseNotFound, action.ErrReleaseDisappeared:
			return install, nil
		case action.ErrReleaseNotObserved, action.ErrReleaseDigest:
			return upgrade, nil
		default:
			return nil, err
		}
	}

	if rls.Info.Status.IsPending() {
		return unlock, nil
	}

	switch rls.Info.Status {
	case release.StatusDeployed:
		if testSpec := r.object.Spec.GetTest(); testSpec.Enable {
			if !intrelease.HasBeenTested(rls) {
				return test, nil
			}
			if !testSpec.IgnoreFailures && intrelease.HasFailedTests(rls) {
				if r.object.Status.Previous != nil {
					return rollback, nil
				}
				return uninstall, nil
			}
		}
	case release.StatusFailed:
		if r.object.Status.Previous != nil {
			return rollback, nil
		}
		return uninstall, nil
	case release.StatusUninstalled:
		return install, nil
	case release.StatusSuperseded:
		return install, nil
	}

	if err = action.VerifyRelease(rls, r.object, r.chart.Metadata, r.values); err != nil {
		switch err {
		case action.ErrChartChanged:
			return upgrade, nil
		case action.ErrConfigDigest:
			return upgrade, nil
		default:
			return nil, err
		}
	}
	return nil, errAbort
}
