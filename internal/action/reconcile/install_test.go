package reconcile

import (
	"context"
	"testing"
)

func Test_install(t *testing.T) {
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
			if err := install(tt.args.ctx, tt.args.r); (err != nil) != tt.wantErr {
				t.Errorf("install() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
