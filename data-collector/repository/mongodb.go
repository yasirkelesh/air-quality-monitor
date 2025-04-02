// repository/mongodb.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/yasirkelesh/data-collector/domain"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PollutionRepository veri saklama arayüzü
type PollutionRepository interface {
	Save(ctx context.Context, data *domain.PollutionData) (string, error)
	Close() error
}

// MongoRepository MongoDB'ye erişim sağlayan repository
type MongoRepository struct {
	client     *mongo.Client
	database   string
	collection string
}

// NewMongoRepository yeni bir MongoDB repository oluşturur
func NewMongoRepository(uri, database, collection string) (PollutionRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// MongoDB bağlantı ayarları
	clientOptions := options.Client().ApplyURI(uri)

	// MongoDB'ye bağlan
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Bağlantıyı test et
	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	return &MongoRepository{
		client:     client,
		database:   database,
		collection: collection,
	}, nil
}

// Save veriyi MongoDB'ye kaydeder
func (r *MongoRepository) Save(ctx context.Context, data *domain.PollutionData) (string, error) {
	// Zaman damgası yoksa ekle
	if data.Timestamp.IsZero() {
		data.Timestamp = time.Now().UTC()
	}

	// Koleksiyonu al
	coll := r.client.Database(r.database).Collection(r.collection)

	// MongoDB'ye kaydet
	result, err := coll.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}

	// ObjectID'yi string olarak dön
	oid, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return "", errors.New("cannot convert InsertedID to ObjectID")
	}

	return oid.Hex(), nil
}

// Close MongoDB bağlantısını kapatır
func (r *MongoRepository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.client.Disconnect(ctx)
}
