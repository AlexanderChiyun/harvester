package image

import (
	"io"
	"fmt"
        "time"
	"net/http"
	"reflect"
	"github.com/gorilla/mux"
	lhv1beta1 "github.com/longhorn/longhorn-manager/k8s/pkg/apis/longhorn/v1beta1"
	"github.com/pkg/errors"
	"github.com/rancher/apiserver/pkg/apierror"
	"github.com/rancher/apiserver/pkg/types"
	"github.com/rancher/wrangler/pkg/schemas/validation"
	"github.com/sirupsen/logrus"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apisv1beta1 "github.com/harvester/harvester/pkg/apis/harvesterhci.io/v1beta1"
	"github.com/harvester/harvester/pkg/generated/controllers/harvesterhci.io/v1beta1"
	ctllhv1beta1 "github.com/harvester/harvester/pkg/generated/controllers/longhorn.io/v1beta1"
	"github.com/harvester/harvester/pkg/util"
	"mime/multipart"
)

const (
	actionUpload   = "upload"
	actionDownload = "download"
)

func Formatter(request *types.APIRequest, resource *types.RawResource) {
	resource.Actions = make(map[string]string, 1)
	if request.AccessControl.CanUpdate(request, resource.APIObject, resource.Schema) != nil {
		return
	}

	if resource.APIObject.Data().String("spec", "sourceType") == apisv1beta1.VirtualMachineImageSourceTypeUpload {
		resource.AddAction(request, actionUpload)
	}
}

type ImageHandler struct {
	httpClient                  http.Client
	Images                      v1beta1.VirtualMachineImageClient
	ImageCache                  v1beta1.VirtualMachineImageCache
	BackingImageDataSources     ctllhv1beta1.BackingImageDataSourceClient
	BackingImageDataSourceCache ctllhv1beta1.BackingImageDataSourceCache
}

func (h ImageHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if err := h.do(rw, req); err != nil {
		status := http.StatusInternalServerError
		if e, ok := err.(*apierror.APIError); ok {
			status = e.Code.Status
		}
		rw.WriteHeader(status)
		_, _ = rw.Write([]byte(err.Error()))
		return
	}
	rw.WriteHeader(http.StatusOK)
}

func (h ImageHandler) do(rw http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	if req.Method == http.MethodGet {
		return h.doGet(vars["link"], rw, req)
	} else if req.Method == http.MethodPost {
		return h.doPost(vars["action"], rw, req)
	}

	return apierror.NewAPIError(validation.InvalidAction, fmt.Sprintf("Unsupported method %s", req.Method))
}

func (h ImageHandler) doGet(link string, rw http.ResponseWriter, req *http.Request) error {
	switch link {
	case actionDownload:
		return h.downloadImage(rw, req)
	default:
		return apierror.NewAPIError(validation.InvalidAction, fmt.Sprintf("Unsupported GET action %s", link))
	}
}

func (h ImageHandler) doPost(action string, rw http.ResponseWriter, req *http.Request) error {
	switch action {
	case actionUpload:
		return h.uploadImage(rw, req)
	default:
		return apierror.NewAPIError(validation.InvalidAction, fmt.Sprintf("Unsupported POST action %s", action))
	}
}

func (h ImageHandler) downloadImage(rw http.ResponseWriter, req *http.Request) error {
	vars := mux.Vars(req)
	namespace := vars["namespace"]
	name := vars["name"]
	vmImage, err := h.Images.Get(namespace, name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get VMImage with name(%s), ns(%s), error: %w", name, namespace, err)
	}

	imgName := vmImage.Annotations["harvesterhci.io/image-name"]
	objName := fmt.Sprintf("%s-%s-%s", namespace, name, imgName)

	obj, err := util.GetObject(util.DefaultBucket, objName)
    if err != nil {
		logrus.Debug("Download image error:", err)
		return err
	}
    defer obj.Close()

	rw.Header().Set("Content-Disposition", "attachment; filename="+imgName)
	rw.Header().Set("Content-Type", "multipart/form-data")

	if _, err := io.Copy(rw, obj); err != nil {
		return fmt.Errorf("failed to copy download content to target(%s), err: %w", objName, err)
	}

	return nil
}

func (h ImageHandler) uploadImage(rw http.ResponseWriter, req *http.Request) error {
	// action:upload name:image-xxxxx namespace:default type:harvesterhci.io.virtualmachineimages
	vars := mux.Vars(req)
	namespace := vars["namespace"]
	name := vars["name"]
	image, err := h.Images.Get(namespace, name, metav1.GetOptions{})

	reader, err := req.MultipartReader()
	if err != nil {
		return err
	}

	var imgPart *multipart.Part

	for {
		if imgPart, err = reader.NextPart(); err != nil {
			return err
		}
		break
	}

	objName := fmt.Sprintf("%s-%s-%s", namespace, name, imgPart.FileName()) // default|image-7wtcd|cirros-qcow2.img
	info, err := util.PutObject(util.DefaultBucket, objName, imgPart)
	if err != nil {
		logrus.Debug("Upload image error:", err)
		return err
	}
	logrus.Debug(info)

	err = h.updateStatusOnConflict(image, 100, info.Size, info.Location)
	if err != nil {
		logrus.Debug("Upate status err:", err)
		return err
	}

	storageClass := image.Annotations["harvesterhci.io/storageClassName"]
	dvName := fmt.Sprintf("%s-%s", namespace, name)

	if err = util.CreateDataVolume(namespace, dvName, storageClass, info.Location);err != nil {
		logrus.Debug("Create DV error:", err)
		return err
	}
	return nil
}

func (h ImageHandler) updateStatusOnConflict(image *apisv1beta1.VirtualMachineImage, progress int, size int64, location string) error {
	retry := 3
	for i := 0; i < retry; i++ {
		current, err := h.ImageCache.Get(image.Namespace, image.Name)
		if err != nil {
			return err
		}
		if current.DeletionTimestamp != nil {
			return nil
		}

		toUpdate := current.DeepCopy()
		toUpdate.Status.Progress = progress
		toUpdate.Status.Size = size
		toUpdate.Annotations["harvesterhci.io/imageLocation"] = location

		apisv1beta1.ImageImported.SetStatusBool(toUpdate, true)
		apisv1beta1.ImageImported.Reason(toUpdate, "Imported")

		if reflect.DeepEqual(current, toUpdate) {
			return nil
		}
		_, err = h.Images.Update(toUpdate)
		if err == nil || !apierrors.IsConflict(err) {
			return err
		}
		time.Sleep(2 * time.Second)
	}
	return errors.New("failed to update image uploaded status, max retries exceeded")
}

func (h ImageHandler) waitForBackingImageDataSourceReady(name string) error {
	retry := 30
	for i := 0; i < retry; i++ {
		ds, err := h.BackingImageDataSources.Get(util.LonghornSystemNamespaceName, name, metav1.GetOptions{})
		if err != nil && !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed waiting for backing image data source to be ready: %w", err)
		}
		if err == nil {
			if ds.Status.CurrentState == lhv1beta1.BackingImageStatePending {
				return nil
			}
			if ds.Status.CurrentState == lhv1beta1.BackingImageStateFailed {
				return errors.New(ds.Status.Message)
			}
		}
		time.Sleep(2 * time.Second)
	}
	return errors.New("timeout waiting for backing image data source to be ready")
}

func (h ImageHandler) updateImportedConditionOnConflict(image *apisv1beta1.VirtualMachineImage,
	status, reason, message string) error {
	retry := 3
	for i := 0; i < retry; i++ {
		current, err := h.ImageCache.Get(image.Namespace, image.Name)
		if err != nil {
			return err
		}
		if current.DeletionTimestamp != nil {
			return nil
		}
		toUpdate := current.DeepCopy()
		apisv1beta1.ImageImported.SetStatus(toUpdate, status)
		apisv1beta1.ImageImported.Reason(toUpdate, reason)
		apisv1beta1.ImageImported.Message(toUpdate, message)
		if reflect.DeepEqual(current, toUpdate) {
			return nil
		}
		_, err = h.Images.Update(toUpdate)
		if err == nil || !apierrors.IsConflict(err) {
			return err
		}
		time.Sleep(2 * time.Second)
	}
	return errors.New("failed to update image uploaded condition, max retries exceeded")
}

