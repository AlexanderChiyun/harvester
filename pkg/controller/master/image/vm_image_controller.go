package image

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"net/http"

	harvesterv1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	ctlharvesterv1 "github.com/harvester/harvester/pkg/generated/controllers/harvesterhci.io/v1beta1"
	lhv1beta1 "github.com/harvester/harvester/pkg/generated/controllers/longhorn.io/v1beta1"
	"github.com/harvester/harvester/pkg/util"
	ctlcorev1 "github.com/rancher/wrangler/pkg/generated/controllers/core/v1"
	ctlstoragev1 "github.com/rancher/wrangler/pkg/generated/controllers/storage/v1"
	"k8s.io/apimachinery/pkg/api/errors"
)

// vmImageHandler syncs status on vm image changes, and manage a storageclass & a backingimage per vm image
type vmImageHandler struct {
	httpClient     http.Client
	storageClasses ctlstoragev1.StorageClassClient
	images         ctlharvesterv1.VirtualMachineImageClient
	backingImages  lhv1beta1.BackingImageClient
	pvcCache       ctlcorev1.PersistentVolumeClaimCache
}

func (h *vmImageHandler) OnChanged(_ string, image *harvesterv1.VirtualMachineImage) (*harvesterv1.VirtualMachineImage, error) {
	if image == nil || image.DeletionTimestamp != nil {
		return image, nil
	}

	if harvesterv1.ImageInitialized.GetStatus(image) == "" {
		return h.initialize(image)
	}//else if image.Spec.URL != image.Status.AppliedURL {
		// URL is changed, recreate the storageclass and backingimage
		//return h.initialize(image)

	//}

	// sync display_name to labels in order to list by labelSelector
	if image.Spec.DisplayName != image.Labels[util.LabelImageDisplayName] {
		toUpdate := image.DeepCopy()
		if toUpdate.Labels == nil {
			toUpdate.Labels = map[string]string{}
		}
		toUpdate.Labels[util.LabelImageDisplayName] = image.Spec.DisplayName
		return h.images.Update(toUpdate)
	}

	return image, nil
}

func (h *vmImageHandler) OnRemove(_ string, image *harvesterv1.VirtualMachineImage) (*harvesterv1.VirtualMachineImage, error) {
	if image == nil {
		return nil, nil
	}
	namespace := image.Namespace
	name := image.Name
	storageClass := image.Annotations[util.AnnotationStorageClassName]
	dvName := fmt.Sprintf("%s-%s", namespace, name)
	if err := util.DeleteDataVolume(namespace, dvName, storageClass); err != nil {
		if !errors.IsNotFound(err) {
			logrus.Debug("Delete DV error:", err)
			return nil, err
		}
	}

	imageName := image.Annotations["harvesterhci.io/image-name"]
	objName := fmt.Sprintf("%s-%s-%s", namespace, name, imageName)
	if err := util.RemoveObject(util.DefaultBucket, objName); err != nil {
		if !errors.IsNotFound(err) {
			logrus.Debug("Delete image object error:", err)
			return nil, err
		}
	}
	return image, nil
}

func (h *vmImageHandler) initialize(image *harvesterv1.VirtualMachineImage) (*harvesterv1.VirtualMachineImage, error) {

	toUpdate := image.DeepCopy()
	toUpdate.Status.AppliedURL = toUpdate.Spec.URL
	//toUpdate.Status.StorageClassName = util.GetImageStorageClassName(image.Name)

	if image.Spec.SourceType == harvesterv1.VirtualMachineImageSourceTypeDownload {
		resp, err := h.httpClient.Head(image.Spec.URL)
		if err != nil {
			harvesterv1.ImageInitialized.False(toUpdate)
			harvesterv1.ImageInitialized.Message(toUpdate, err.Error())
			return h.images.Update(toUpdate)
		}
		defer resp.Body.Close()

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
			harvesterv1.ImageInitialized.False(toUpdate)
			harvesterv1.ImageInitialized.Message(toUpdate, fmt.Sprintf("got %d status code from %s", resp.StatusCode, image.Spec.URL))
			return h.images.Update(toUpdate)
		}

		if resp.ContentLength > 0 {
			toUpdate.Status.Size = resp.ContentLength
		}
	} else {
		toUpdate.Status.Progress = 0
	}

	harvesterv1.ImageImported.Unknown(toUpdate)
	harvesterv1.ImageImported.Reason(toUpdate, "Importing")
	harvesterv1.ImageInitialized.True(toUpdate)
	harvesterv1.ImageInitialized.Reason(toUpdate, "Initialized")

	return h.images.Update(toUpdate)
}

/*
func (h *vmImageHandler) initialize(image *harvesterv1.VirtualMachineImage) (*harvesterv1.VirtualMachineImage, error) {
	if err := h.createBackingImage(image); err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}
	if err := h.createStorageClass(image); err != nil && !errors.IsAlreadyExists(err) {
		return nil, err
	}

	toUpdate := image.DeepCopy()
	toUpdate.Status.AppliedURL = toUpdate.Spec.URL
	toUpdate.Status.StorageClassName = util.GetImageStorageClassName(image.Name)

	if image.Spec.SourceType == harvesterv1.VirtualMachineImageSourceTypeDownload {
		resp, err := h.httpClient.Head(image.Spec.URL)
		if err != nil {
			harvesterv1.ImageInitialized.False(toUpdate)
			harvesterv1.ImageInitialized.Message(toUpdate, err.Error())
			return h.images.Update(toUpdate)
		}
		defer resp.Body.Close()

		if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusBadRequest {
			harvesterv1.ImageInitialized.False(toUpdate)
			harvesterv1.ImageInitialized.Message(toUpdate, fmt.Sprintf("got %d status code from %s", resp.StatusCode, image.Spec.URL))
			return h.images.Update(toUpdate)
		}

		if resp.ContentLength > 0 {
			toUpdate.Status.Size = resp.ContentLength
		}
	} else {
		toUpdate.Status.Progress = 0
	}

	harvesterv1.ImageImported.Unknown(toUpdate)
	harvesterv1.ImageImported.Reason(toUpdate, "Importing")
	harvesterv1.ImageInitialized.True(toUpdate)
	harvesterv1.ImageInitialized.Reason(toUpdate, "Initialized")

	return h.images.Update(toUpdate)
}


func (h *vmImageHandler) createBackingImage(image *harvesterv1.VirtualMachineImage) error {
	bi := &v1beta1.BackingImage{
		ObjectMeta: metav1.ObjectMeta{
			Name:      util.GetBackingImageName(image),
			Namespace: util.LonghornSystemNamespaceName,
			Annotations: map[string]string{
				util.AnnotationImageID: ref.Construct(image.Namespace, image.Name),
			},
		},
		Spec: v1beta1.BackingImageSpec{
			SourceType:       v1beta1.BackingImageDataSourceType(image.Spec.SourceType),
			SourceParameters: map[string]string{},
			Checksum:         image.Spec.Checksum,
		},
	}
	if image.Spec.SourceType == harvesterv1.VirtualMachineImageSourceTypeDownload {
		bi.Spec.SourceParameters[v1beta1.DataSourceTypeDownloadParameterURL] = image.Spec.URL
	}

	if image.Spec.SourceType == harvesterv1.VirtualMachineImageSourceTypeExportVolume {
		pvc, err := h.pvcCache.Get(image.Spec.PVCNamespace, image.Spec.PVCName)
		if err != nil {
			return fmt.Errorf("failed to get pvc %s/%s, error: %s", image.Spec.PVCName, image.Namespace, err.Error())
		}

		bi.Spec.SourceParameters[lhcontroller.DataSourceTypeExportFromVolumeParameterVolumeName] = pvc.Spec.VolumeName
		bi.Spec.SourceParameters[lhmanager.DataSourceTypeExportFromVolumeParameterExportType] = lhmanager.DataSourceTypeExportFromVolumeParameterExportTypeRAW
	}

	_, err := h.backingImages.Create(bi)
	return err
}

func (h *vmImageHandler) createStorageClass(image *harvesterv1.VirtualMachineImage) error {
	reclaimPolicy := corev1.PersistentVolumeReclaimDelete
	volumeBindingMode := storagev1.VolumeBindingImmediate
	sc := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: util.GetImageStorageClassName(image.Name),
		},
		Provisioner:          types.LonghornDriverName,
		ReclaimPolicy:        &reclaimPolicy,
		AllowVolumeExpansion: pointer.BoolPtr(true),
		VolumeBindingMode:    &volumeBindingMode,
		Parameters:           util.GetImageStorageClassParameters(image),
	}

	_, err := h.storageClasses.Create(sc)
	return err
}

func (h *vmImageHandler) deleteBackingImage(image *harvesterv1.VirtualMachineImage) error {
	return h.backingImages.Delete(util.LonghornSystemNamespaceName, util.GetBackingImageName(image), &metav1.DeleteOptions{})
}

func (h *vmImageHandler) deleteStorageClass(image *harvesterv1.VirtualMachineImage) error {
	return h.storageClasses.Delete(util.GetImageStorageClassName(image.Name), &metav1.DeleteOptions{})
}

func (h *vmImageHandler) deleteBackingImageAndStorageClass(image *harvesterv1.VirtualMachineImage) error {
	if err := h.deleteBackingImage(image); err != nil && !errors.IsNotFound(err) {
		return err
	}
	if err := h.deleteStorageClass(image); err != nil && !errors.IsNotFound(err) {
		return err
	}
	return nil
}
*/
