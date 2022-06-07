package action

import (
	"errors"
	intrelease "github.com/fluxcd/helm-controller/internal/release"
	"helm.sh/helm/v3/pkg/release"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	intchartutil "github.com/fluxcd/helm-controller/internal/chartutil"
	"github.com/opencontainers/go-digest"
	"helm.sh/helm/v3/pkg/chart"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/storage/driver"
)

var (
	ErrReleaseDisappeared = errors.New("observed release disappeared from storage")
	ErrReleaseNotFound    = errors.New("no release found")
	ErrReleaseNotObserved = errors.New("release not observed to be made by controller")
	ErrReleaseDigest      = errors.New("release digest verification error")
	ErrChartChanged       = errors.New("release chart changed")
	ErrConfigDigest       = errors.New("release config verification error")
)

func VerifyStorage(config *Configuration, obj *helmv2.HelmRelease) (*release.Release, error) {
	cfg, err := NewActionConfig(config)
	if err != nil {
		return nil, err
	}

	curRel := obj.Status.Current
	rls, err := cfg.Releases.Last(obj.GetReleaseName())
	if err != nil {
		if errors.Is(err, driver.ErrReleaseNotFound) {
			if curRel != nil {
				return nil, ErrReleaseDisappeared
			}
			return nil, ErrReleaseNotFound
		}
		return nil, err
	}
	if curRel == nil {
		return rls, ErrReleaseNotObserved
	}

	obs := intrelease.ObserveRelease(rls, intrelease.IgnoreHookTestEvents)
	relDig, err := digest.Parse(obj.Status.Current.Digest)
	if err != nil {
		return rls, ErrReleaseDigest
	}
	verifier := relDig.Verifier()
	if err := obs.Encode(verifier); err != nil {
		// We are expected to be able to encode valid JSON, error out without a
		// typed error assuming malfunction to signal to e.g. retry.
		return nil, err
	}
	if !verifier.Verified() {
		return rls, ErrReleaseNotObserved
	}
	return rls, nil
}

func VerifyRelease(rls *release.Release, obj *helmv2.HelmRelease, chrt *chart.Metadata, vals chartutil.Values) error {
	if rls == nil {
		return ErrReleaseNotFound
	}

	if chrt != nil {
		if _, eq := intchartutil.DiffMeta(*rls.Chart.Metadata, *chrt); !eq {
			return ErrChartChanged
		}
	}

	configDig, err := digest.Parse(obj.Status.Current.ConfigDigest)
	if err != nil {
		return ErrConfigDigest
	}
	if !intchartutil.VerifyValues(configDig, vals) {
		return ErrConfigDigest
	}
	return nil
}
