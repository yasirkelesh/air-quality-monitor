// repository/mongodb.go
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/yasirkelesh/data-collector/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// PollutionRepository veri saklama arayüzü
type PollutionRepository interface {
	Save(ctx context.Context, data *domain.PollutionData) (string, error)
	Close() error
	CountData(ctx context.Context) (int64, error)
	FindAll(ctx context.Context, page, pageSize int) ([]*domain.PollutionData, error)
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

// MogoDB'den veriyi alır

// Close MongoDB bağlantısını kapatır
func (r *MongoRepository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.client.Disconnect(ctx)
}

func (r *MongoRepository) FindAll(ctx context.Context, page, pageSize int) ([]*domain.PollutionData, error) {
	coll := r.client.Database(r.database).Collection(r.collection)

	if page < 1 {
		page = 1
	}

	if pageSize < 1 || pageSize > 100 {
		pageSize = 20 // Varsayılan sayfa boyutu
	}

	// Sayfalama seçenekleri
	skip := (page - 1) * pageSize

	findOptions := options.Find().
		SetSkip(int64(skip)).
		SetLimit(int64(pageSize)).
		SetSort(bson.D{{Key: "timestamp", Value: -1}}) // En yeniden en eskiye sırala

	cursor, err := coll.Find(ctx, bson.M{}, findOptions)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []*domain.PollutionData
	if err = cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	return results, nil
}

// toplam kayit sayısını döndür
func (r *MongoRepository) CountData(ctx context.Context) (int64, error) {
	coll := r.client.Database(r.database).Collection(r.collection)

	// Filtreleme işlemi
	count, err := coll.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, err
	}

	return count, nil
}
