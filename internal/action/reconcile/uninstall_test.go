package reconcile

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"helm.sh/helm/v3/pkg/release"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/helm-controller/internal/storage"
)

func Test_uninstall(t *testing.T) {
	type args struct {
		ctx context.Context
		r   *Release
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := uninstall(tt.args.ctx, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("uninstall() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_observeUninstall(t *testing.T) {
	tests := []struct {
		name string
		rls *release.Release
		obj *helmv2.HelmRelease
		want helmv2.HelmReleaseStatus
	}{
		{
			name: "uninstall of current release",
			rls: &release.Release{}
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			observeUninstall(tt.obj)(tt.rls)
			g.Expect(tt.obj.Status).To(Equal(tt.want))
		})
	}
}
