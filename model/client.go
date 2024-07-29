package model

import (
	"context"
	"fmt"
	"log"
	"os"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

var (
	redisClient     *redis.Client
	storageClient   *storage.Client
	firestoreClient *firestore.Client

	GcpProjectID        = os.Getenv("GCP_PROJECT_ID")
	FirestoreDatabaseID = os.Getenv("FIRESTORE_DATABASE_ID")
	FirestoreCollection = os.Getenv("FIRESTORE_COLLECTION")
	RedisUri            = os.Getenv("REDIS_URI")
	RedisPort           = os.Getenv("REDIS_PORT")
	BucketName          = os.Getenv("BUCKET_NAME")
)

func init() {
	ctx := context.Background()
	log.Printf("init redis")
	redisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", RedisUri, RedisPort),
		Password: "", // no password set
		DB:       11,
	})
	if err := redisotel.InstrumentTracing(redisClient); err != nil {
		panic(err)
	}

	log.Printf("init storage")
	var err error
	storageClient, err = storage.NewClient(ctx)
	if err != nil {
		log.Printf("storage.NewClient err: %+v", err)
		panic(err)
	}

	log.Printf("init firestore")
	firestoreClient, err = firestore.NewClientWithDatabase(ctx, GcpProjectID, FirestoreDatabaseID)
	if err != nil {
		log.Printf("firestore.NewClient err: %+v", err)
		panic(err)
	}
}
