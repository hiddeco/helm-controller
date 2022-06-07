package release

import "helm.sh/helm/v3/pkg/release"

func HasStatus(rls *release.Release, status release.Status) bool {
	return rls.Info.Status == status
}

func GetTestHooks(rls *release.Release) map[string]*release.Hook {
	th := make(map[string]*release.Hook)
	for _, h := range rls.Hooks {
		for _, e := range h.Events {
			if e == release.HookTest {
				th[h.Name] = h
			}
		}
	}
	return th
}

func HasBeenTested(rls *release.Release) bool {
	for _, h := range rls.Hooks {
		for _, e := range h.Events {
			if e == release.HookTest {
				if !h.LastRun.StartedAt.IsZero() {
					return true
				}
			}
		}
	}
	return false
}

func HasFailedTests(rls *release.Release) bool {
	for _, h := range rls.Hooks {
		for _, e := range h.Events {
			if e == release.HookTest {
				if h.LastRun.Phase == release.HookPhaseFailed {
					return true
				}
			}
		}
	}
	return false
}
