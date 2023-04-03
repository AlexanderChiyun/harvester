package util

import (
    "io"
    "os"
    "net/url"
    "context"
    minio "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

const (
    DefaultBucket = "images"
    bucketLocation = "us-east-1"
    accessKeyEnvKey = "OSS_ACCESS_KEY"
    secretKeyEnvKey = "OSS_SECRET_KEY"
    endpointEnvKey  = "OSS_ENDPOINT"
)
/*
var (
    accessKey = "minioadmin"
    secretKey = "minioadmin"
    endpoint = "http://10.12.21.66:9000"
)
*/
func initCli() (*minio.Client, error) {
    accessKey := os.Getenv(accessKeyEnvKey)
    secretKey := os.Getenv(secretKeyEnvKey)
    endpoint := os.Getenv(endpointEnvKey)

    u, err := url.Parse(endpoint)
    if err != nil {
        return nil, err
    }
    return minio.New(u.Host, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: u.Scheme == "https",
    })
}
/*
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
*/

func InitBucket(bucket string) error {
    cli, err := initCli()

    if err != nil {
        return err
    }
    existed, err := cli.BucketExists(context.Background(), bucket)
    if err == nil && existed {
        return nil
    }
    return cli.MakeBucket(context.Background(), bucket, minio.MakeBucketOptions{Region: bucketLocation})
}

func PutObject(bucket string, object string, reader io.Reader) (minio.UploadInfo, error) {
    cli, err := initCli()

    if err != nil {
        return minio.UploadInfo{}, err
    }
    info, err := cli.PutObject(context.Background(), bucket, object, reader, -1, minio.PutObjectOptions{})
    if err != nil {
        return minio.UploadInfo{}, err
    }
    return info, nil
}

func GetObject(bucket string, object string) (*minio.Object, error) {
    cli, err := initCli()

    if err != nil {
        return nil, err
    }
    return cli.GetObject(context.Background(), bucket, object, minio.GetObjectOptions{})
}

func RemoveObject(bucket string, object string) error {
    cli, err := initCli()

    if err != nil {
        return err
    }
    return cli.RemoveObject(context.Background(), bucket, object, minio.RemoveObjectOptions{})
}

