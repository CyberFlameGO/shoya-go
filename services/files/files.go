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
	"net/url"
	"time"
)

var MinioClient *minio.Client

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

	if config.RuntimeConfig.Api == nil {
		panic("error reading config file: RuntimeConfig.Api was nil")
	}
}

// initializeRedis initializes the redis clients
func initializeRedis() {
	config.HarvestRedisClient = redis.NewClient(&redis.Options{
		Addr:     config.RuntimeConfig.Api.Redis.Host,
		Password: config.RuntimeConfig.Api.Redis.Password,
		DB:       config.RuntimeConfig.Api.Redis.Database,
	})

	_, err := config.RedisClient.Ping(context.Background()).Result()
	_, err2 := config.HarvestRedisClient.Ping(context.Background()).Result()
	if err != nil || err2 != nil {
		panic(err)
	}
}

// initializeApiConfig initializes harvester client used to configure the API
func initializeApiConfig() {
	h, err := harvester.New(&config.ApiConfiguration).
		WithRedisSeed(config.HarvestRedisClient).
		WithRedisMonitor(config.HarvestRedisClient, time.Duration(config.RuntimeConfig.Api.ApiConfigRefreshRateMs)*time.Millisecond).
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
	f, err := MinioClient.PresignedGetObject(context.TODO(), "shoya-test", in.GetName(), time.Minute*5, make(url.Values))
	if err != nil {
		log.Printf("[%v] [GetFile] [ERROR]: %v", time.Now(), err)
		return nil, err
	}

	fileUrl := f.String()
	return &pb.GetFileResponse{Url: &fileUrl}, nil
}

func (s *server) CreateFile(ctx context.Context, in *pb.CreateFileRequest) (*pb.CreateFileResponse, error) {
	log.Printf("Received: %v", in.GetName())
	u, err := MinioClient.PresignedPutObject(context.TODO(), "shoya-test", in.GetName(), time.Hour*3)
	if err != nil {
		log.Printf("[%v] [CreateFile] [ERROR]: %v", time.Now(), err)
		return nil, err
	}
	uploadUrl := u.String()
	return &pb.CreateFileResponse{Url: &uploadUrl}, nil
}

func initMinioClient() {
	var err error
	// Initialize minio client object.
	MinioClient, err = minio.New(config.ApiConfiguration.FilesS3Endpoint.Get(), &minio.Options{
		Creds:  credentials.NewStaticV4(config.ApiConfiguration.FilesAccessKey.Get(), config.ApiConfiguration.FilesSecretKey.Get(), ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalln(err)
	}
}
