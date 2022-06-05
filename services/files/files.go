package main

import (
	"context"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/gtsatsis/harvester"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/tkanos/gonfig"
	"gitlab.com/george/shoya-go/config"
	pb "gitlab.com/george/shoya-go/gen/v1/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var MinioClient *minio.Core

type server struct {
	pb.UnimplementedFileServer
}

func main() {
	initializeConfig()
	initializeRedis()
	initializeApiConfig()
	initMinioClient()

	lis, err := net.Listen("tcp", config.RuntimeConfig.Files.ListenAddress)
	if err != nil {
		panic(err)
	}

	s := grpc.NewServer()
	pb.RegisterFileServer(s, &server{})

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// initializeConfig reads the config.json file and initializes the runtime config
func initializeConfig() {
	err := gonfig.GetConf("config.json", &config.RuntimeConfig)
	if err != nil {
		panic("error reading config file")
	}

	if config.RuntimeConfig.Files == nil {
		panic("error reading config file: RuntimeConfig.Api was nil")
	}
}

// initializeRedis initializes the redis clients
func initializeRedis() {
	config.HarvestRedisClient = redis.NewClient(&redis.Options{
		Addr:     config.RuntimeConfig.Files.Redis.Host,
		Password: config.RuntimeConfig.Files.Redis.Password,
		DB:       config.RuntimeConfig.Files.Redis.Database,
	})

	_, err := config.HarvestRedisClient.Ping(context.Background()).Result()
	if err != nil {
		panic(err)
	}
}

// initializeApiConfig initializes harvester client used to configure the API
func initializeApiConfig() {
	h, err := harvester.New(&config.ApiConfiguration).
		WithRedisSeed(config.HarvestRedisClient).
		WithRedisMonitor(config.HarvestRedisClient, 50*time.Millisecond).
		Create()
	if err != nil {
		panic(fmt.Errorf("failed to set up configuration harvester: %v", err))
	}

	err = h.Harvest(context.Background())
	if err != nil {
		panic(fmt.Errorf("failed to harvest configuration: %v", err))
	}
}

func (s *server) GetFile(ctx context.Context, in *pb.GetFileRequest) (*pb.GetFileResponse, error) {
	headers := http.Header{}
	f, err := MinioClient.PresignHeader(context.TODO(), http.MethodGet, config.ApiConfiguration.FilesS3Bucket.Get(), in.GetName(), time.Minute*5, make(url.Values), headers)
	if err != nil {
		log.Printf("[%v] [GetFile] [ERROR]: %v", time.Now(), err)
		return nil, err
	}

	fileUrl := f.String()
	return &pb.GetFileResponse{Url: &fileUrl}, nil
}

func (s *server) CreateFile(ctx context.Context, in *pb.CreateFileRequest) (*pb.CreateFileResponse, error) {
	headers := http.Header{}
	headers.Add("Content-MD5", in.GetMd5())
	u, err := MinioClient.PresignHeader(context.TODO(), http.MethodPut, config.ApiConfiguration.FilesS3Bucket.Get(), in.GetName(), time.Hour*3, url.Values{}, headers)
	if err != nil {
		log.Printf("[%v] [CreateFile] [ERROR]: %v", time.Now(), err)
		return nil, err
	}
	uploadUrl := u.String()
	return &pb.CreateFileResponse{Url: &uploadUrl}, nil
}

func (s *server) HealthCheck(ctx context.Context, in *pb.HealthCheckRequest) (*pb.HealthCheckResponse, error) {
	ok := true
	return &pb.HealthCheckResponse{Ok: &ok}, nil
}

func initMinioClient() {
	var endpointUrl = config.ApiConfiguration.FilesS3Endpoint.Get()
	var endpointIsSecure = strings.Contains(endpointUrl, "https://")
	var endpoint = strings.ReplaceAll(strings.ReplaceAll(endpointUrl, "https://", ""), "http://", "")
	var err error
	// Initialize minio client object.
	MinioClient, err = minio.NewCore(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.ApiConfiguration.FilesS3AccessKey.Get(), config.ApiConfiguration.FilesS3SecretKey.Get(), ""),
		Secure: endpointIsSecure,
	})
	if err != nil {
		log.Fatalln(err)
	}
}
