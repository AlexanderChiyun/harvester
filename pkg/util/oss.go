package util

import (
    "context"
    minio "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
    "io"
    "net/url"
    "strings"
)

const (
    DefaultBucket = "images"
    bucketLocation = "us-east-1"
)

var gOSSCli *minio.Client

var (
    accessKey = "minioadmin"
    secretKey = "minioadmin"
    endpoint = "http://10.12.21.66:9000"
)

func initCli()  error {
    if gOSSCli != nil {
        return nil
    }
    withSSL := false
    u, err := url.Parse(endpoint)
    if err != nil {
        return err
    }
    if strings.Compare(u.Scheme, "https") == 0{
        withSSL = true
    }
    ep := u.Host

    ret, err := minio.New(ep,
        &minio.Options{
            Creds: credentials.NewStaticV4(accessKey, secretKey, ""),
            Secure: withSSL,
        })
    if err != nil {
        gOSSCli = nil
        return err
    } else {
        gOSSCli = ret
        return nil
    }
}

func InitBucket(bucket string) error {
    if err := initCli();err != nil {
        return err
    }
    existed, err := gOSSCli.BucketExists(context.Background(), bucket)
    if err == nil && existed {
        return nil
    }
    return gOSSCli.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{Region: bucketLocation})
}

func PutObject(bucket string, object string, reader io.Reader) (minio.UploadInfo, error) {
    if err := initCli();err != nil {
        return minio.UploadInfo{}, err
    }
    info, err := gOSSCli.PutObject(context.Background(), bucket, object, reader, -1, minio.PutObjectOptions{})
    if err != nil {
        return minio.UploadInfo{}, err
    }
    return info, nil
}

func GetObject(bucket string, object string) (*minio.Object, error) {
    if err := initCli();err != nil {
        return nil, err
    }
    return gOSSCli.GetObject(context.Background(), bucket, object, minio.GetObjectOptions{})
}

func RemoveObject(bucket string, object string) error {
    if err := initCli();err != nil {
        return err
    }
    return gOSSCli.RemoveObject(context.Background(), bucket, object, minio.RemoveObjectOptions{})
}
