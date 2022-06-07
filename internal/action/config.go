package action

import (
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/fluxcd/helm-controller/internal/kube"
	"github.com/fluxcd/pkg/runtime/client"
	"helm.sh/helm/v3/pkg/action"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/fluxcd/helm-controller/internal/storage"
)

const (
	DefaultStorageDriver = "secret"
)

type Configuration struct {
	Getter           genericclioptions.RESTClientGetter
	StorageDriver    string
	StorageNamespace string
	Observers        []storage.ObserveFunc
	Log              action.DebugLog
}

func MakeConfiguration(obj *helmv2.HelmRelease, secret *corev1.Secret, clientOpts client.Options, kubeCfgOpts client.KubeConfigOptions) (*Configuration, error) {
	opts := []kube.ClientGetterOption{kube.WithClientOptions(clientOpts)}
	if obj.Spec.ServiceAccountName != "" {
		opts = append(opts, kube.WithImpersonate(obj.Spec.ServiceAccountName))
	}
	if secret != nil {
		kubeConfig, err := kube.ConfigFromSecret(secret, obj.Spec.KubeConfig.SecretRef.Key)
		if err != nil {
			return nil, err
		}
		opts = append(opts, kube.WithKubeConfig(kubeConfig, kubeCfgOpts))
	}
	getter, err := kube.BuildClientGetter(obj.GetReleaseNamespace(), opts...)
	if err != nil {
		return nil, err
	}
	return &Configuration{
		Getter:           getter,
		StorageDriver:    DefaultStorageDriver,
		StorageNamespace: obj.GetStorageNamespace(),
	}, nil
}

func NewActionConfig(cfg *Configuration) (*action.Configuration, error) {
	config := new(action.Configuration)

	storageDriver := cfg.StorageDriver
	if storageDriver == "" {
		storageDriver = DefaultStorageDriver
	}

	if err := config.Init(cfg.Getter, cfg.StorageNamespace, storageDriver, cfg.Log); err != nil {
		return nil, err
	}

	if len(cfg.Observers) > 0 {
		observer := storage.NewObserver(config.Releases.Driver, cfg.Observers...)
		config.Releases.Driver = observer
	}

	return config, nil
}
