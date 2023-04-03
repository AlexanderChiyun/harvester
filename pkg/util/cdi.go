package util

import (
    "context"
    "github.com/sirupsen/logrus"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
    cdi "kubevirt.io/client-go/generated/containerized-data-importer/clientset/versioned"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/client-go/rest"
)

const (
    DefaultSize = "10Gi"
)

var DefaultMod = corev1.PersistentVolumeBlock

func initCDICli() (*cdi.Clientset, error) {
    config, err := rest.InClusterConfig()
    if err != nil {
        logrus.Errorf("InClusterConfig failed: %v", err)
        return nil, err
    }
    client, err := cdi.NewForConfig(config)
    if err != nil {
        logrus.Errorf("NewForConfig failed: %v", err)
        return nil, err
    }
    return client, nil
}

func CreateDataVolume(namespace, name, storageclass, imgurl string) error {
    newdv := &cdiv1.DataVolume{
        ObjectMeta: metav1.ObjectMeta{
            Name: name,
        },
        Spec: cdiv1.DataVolumeSpec{
            Source: &cdiv1.DataVolumeSource{
                HTTP: &cdiv1.DataVolumeSourceHTTP{
                    URL: imgurl,
                },
            },
            PVC: &corev1.PersistentVolumeClaimSpec{
                AccessModes: []corev1.PersistentVolumeAccessMode{
                    corev1.ReadWriteMany,
                },
                Resources: corev1.ResourceRequirements{
                    Requests: corev1.ResourceList{
                        corev1.ResourceStorage: resource.MustParse(DefaultSize),
                    },
                },
                StorageClassName: &storageclass,
                VolumeMode: &DefaultMod,
            },
        },
    }

    cli, err := initCDICli()
    if err != nil {
        logrus.Errorf("Init CDI client failed.")
        return err
    }
    _, err = cli.CdiV1beta1().DataVolumes(namespace).Create(context.Background(), newdv, metav1.CreateOptions{})
    return err
}

func DeleteDataVolume(namespace string, name string, storageclass string) error {
    cli, err := initCDICli()
    if err != nil {
        logrus.Errorf("Init CDI client failed.")
        return err
    }
    return cli.CdiV1beta1().DataVolumes(namespace).Delete(context.Background(), name, metav1.DeleteOptions{})
}

