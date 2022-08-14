package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type (
	S3Config struct {
		AwsDefaultRegion string
		EndPoint         string
		BucketName       string
	}
	S3Client struct {
		clientCfg *S3Config
		sdkCfg    aws.Config
	}
)

// 処理を開始するには、localstack_s3ディレクトリで下記コマンドを実行する
// docker compose exec golang go run ./go_app  
func main() {
	fmt.Println("処理開始")
	// S3の設定を環境変数(docker-composeのenvironmentで設定したもの)から取得して設定する
	cfg := S3Config{
		AwsDefaultRegion: os.Getenv("AWS_DEFAULT_REGION"),
		EndPoint:         os.Getenv("S3_ENDPOINT"),
		BucketName:       os.Getenv("S3_BUCKET_NAME"),
	}

	// S3クライアントの作成
	s3client := NewS3Client(&cfg)

	// データ送信フォームの作成
	api := s3.NewFromConfig(s3client.sdkCfg, func(options *s3.Options) {
		options.UsePathStyle = true
	})

	// 送信するファイル名とファイルの配置先を設定
	csvFileName := "sample.csv"
	filePath := "/app/go_app"

	// ファイルを開く
	var csvFile io.ReadSeekCloser
	csvFile, err := os.Open(filePath + "/" + csvFileName)
	if err != nil {
		fmt.Printf("FileOpenError: %s", err.Error())
		return
	}

	// ファイルの保存先を設定
	s3Location := fmt.Sprintf("%s/%s", "work", csvFileName)

	// S3のタイムアウトの時間
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(3)*time.Second)
	defer cancel()

	// S3バケットにファイルを送信する
	_, err = api.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s3client.clientCfg.BucketName),
		Key:    aws.String(s3Location),
		Body:   csvFile,
	})
	if err != nil {
		fmt.Printf("FileUploadFileError: %s", err.Error())
		return
	}

	// S3バケットからファイルを取得したい場合に使用
	// api.GetObject(ctx, &s3.GetObjectInput{
	// 	Bucket: aws.String(s3client.clientCfg.BucketName),
	// 	Key:    aws.String(s3Location),
	// })

	fmt.Println("処理終了")
}

// S3クライアントの作成
func NewS3Client(cfg *S3Config) *S3Client {
	loadOptions := []func(*config.LoadOptions) error{config.WithRegion(cfg.AwsDefaultRegion)}

	if cfg.EndPoint != "" {
		endpoint := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
			return aws.Endpoint{
				URL: cfg.EndPoint,
			}, nil
		})
		loadOptions = append(loadOptions, config.WithEndpointResolverWithOptions(endpoint))
	}

	sdkCfg, err := config.LoadDefaultConfig(context.TODO(), loadOptions...)
	if err != nil {
		panic(err)
	}

	return &S3Client{cfg, sdkCfg}
}
