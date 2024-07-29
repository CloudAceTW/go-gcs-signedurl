package model

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/storage"
	"github.com/CloudAceTW/go-gcs-signedurl/helper"
	"go.opentelemetry.io/otel"
	otelCode "go.opentelemetry.io/otel/codes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	tracer = otel.Tracer("signurl-model")
)

type SignURL struct {
	Id          string          `json:"id" firestore:"-"`
	PostSignURL string          `json:"postSignURL" firestore:"-"`
	SignURL     string          `json:"signURL" firestore:"-"`
	FileName    string          `json:"-" firestore:"gcs_file_name,omitempty"`
	RedisKey    string          `json:"key" firestore:"-"`
	Err         []error         `json:"-" firestore:"-"`
	Ctx         context.Context `json:"-" firestore:"-"`
}

func NewSignURL(ctx context.Context) (signURL SignURL) {
	ctx, span := tracer.Start(ctx, helper.FuncName())
	defer span.End()

	signURL.Ctx = ctx
	signURL.makeFileNmae()
	signURL.makeRedisKey()
	c := make(chan helper.RespChannel, 2)
	go signURL.makeGcsSignURL(c)
	go signURL.setKeyInRedis(ctx, c)

	for i := 0; i < 2; i++ {
		rc := <-c
		if rc.Err != nil {
			span.SetStatus(otelCode.Error, fmt.Sprintf("err: %+v", signURL.Err))
			signURL.Err = append(signURL.Err, rc.Err)
			return
		}
	}
	span.SetStatus(otelCode.Ok, "OK")
	return
}

func (signURL *SignURL) makeFileNmae() {
	_, span := tracer.Start(signURL.Ctx, helper.FuncName())
	defer span.End()
	signURL.FileName = fmt.Sprintf("%s/%s", time.Now().UTC().Format("2006-01-02"), helper.RandStringBytesMaskImprSrcSB(36))
}

func (signURL *SignURL) makeRedisKey() {
	_, span := tracer.Start(signURL.Ctx, helper.FuncName())
	defer span.End()
	signURL.RedisKey = fmt.Sprintf("signURL:%s", helper.RandStringBytesMaskImprSrcSB(36))
}

func (signURL *SignURL) makeGcsSignURL(c chan helper.RespChannel) {
	_, span := tracer.Start(signURL.Ctx, helper.FuncName())
	defer span.End()
	log.Printf("ready to makeGcsSignURL")
	rc := helper.RespChannel{}
	err := signURL.MakeGcsSignURL("PUT", 15*time.Minute)
	if err != nil {
		rc.Err = err
		c <- rc
		return
	}
	c <- rc
}

func (signURL *SignURL) MakeGcsSignURL(httpMethod string, expireTime time.Duration) error {
	_, span := tracer.Start(signURL.Ctx, helper.FuncName())
	defer span.End()
	optsPut := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  httpMethod,
		Expires: time.Now().Add(expireTime),
	}

	postSignURL, err := storageClient.Bucket(BucketName).SignedURL(signURL.FileName, optsPut)
	if err != nil {
		log.Printf("client.SignedURL err: %+v", err)
		return err
	}
	switch httpMethod {
	case "PUT":
		signURL.PostSignURL = postSignURL
	case "GET":
		signURL.SignURL = postSignURL
	}

	return nil
}

func (signURL *SignURL) setKeyInRedis(ctx context.Context, c chan helper.RespChannel) {
	_, span := tracer.Start(signURL.Ctx, helper.FuncName())
	defer span.End()
	log.Printf("ready to setKeyInRedis")
	rc := helper.RespChannel{}

	err := signURL.SetKeyInRedis(ctx)
	if err != nil {
		log.Printf("redisClient.Set err: %+v", err)
		rc.Err = err
		c <- rc
		return
	}

	c <- rc
}

func (signURL *SignURL) SetKeyInRedis(ctx context.Context) error {
	_, span := tracer.Start(signURL.Ctx, helper.FuncName())
	defer span.End()
	err := redisClient.Set(ctx, signURL.RedisKey, signURL.FileName, 14*time.Minute).Err()
	if err != nil {
		log.Printf("redisClient.Set err: %+v", err)
		return err
	}
	return nil
}

func (signURL *SignURL) GetFileNameFromRedis(ctx context.Context) {
	_, span := tracer.Start(signURL.Ctx, helper.FuncName())
	defer span.End()
	fileName, err := redisClient.Get(ctx, signURL.RedisKey).Result()
	if err != nil {
		log.Printf("redisClient.Get err: %+v", err)
		signURL.Err = append(signURL.Err, err)
		return
	}

	signURL.FileName = fileName
}

func (signURL *SignURL) SetGCSPathToFirestore(ctx context.Context) {
	_, span := tracer.Start(signURL.Ctx, helper.FuncName())
	defer span.End()
	signURL.setID(ctx)
	ref := firestoreClient.Collection(FirestoreCollection).Doc(signURL.Id)
	_, err := ref.Set(ctx, signURL)
	if err != nil {
		log.Printf("firestoreClient.Collection(%s).Doc(%s).Set err: %+v", FirestoreCollection, signURL.Id, err)
		signURL.Err = append(signURL.Err, err)
	}
}

func (signURL *SignURL) GetGCSPathFromFirestore(ctx context.Context) {
	_, span := tracer.Start(signURL.Ctx, helper.FuncName())
	defer span.End()
	ref := firestoreClient.Collection(FirestoreCollection).Doc(signURL.Id)
	docsnap, err := ref.Get(ctx)
	if status.Code(err) == codes.NotFound {
		signURL.FileName = ""
		return
	}
	err = docsnap.DataTo(signURL)
	if err != nil {
		signURL.FileName = ""
		return
	}
}

func (signURL *SignURL) setID(ctx context.Context) {
	_, span := tracer.Start(signURL.Ctx, helper.FuncName())
	defer span.End()
	var id string
	for {
		id = helper.RandStringBytesMaskImprSrcSB(7)
		if _, err := firestoreClient.Collection(FirestoreCollection).Doc(id).Get(ctx); err != nil {
			break
		}
	}
	signURL.Id = id
}
