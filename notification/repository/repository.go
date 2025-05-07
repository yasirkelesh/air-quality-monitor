package repository

import (
	"context"
	"time"

	"github.com/yasirkelesh/notification/domain"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository tanımını arayüz (interface) üzerinden gerçekleştiriyoruz.
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	Close() error
}

// userRepository, UserRepository interface'ini uygulayan yapı olarak tanımlanır.
type userRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

// Close, mongo.Client ile yapılan bağlantıyı kapatır.
func (r *userRepository) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return r.client.Disconnect(ctx)
}

// NewUserRepository, verilen URI, veritabanı ve koleksiyon adıyla yeni bir userRepository oluşturur.
func NewUserRepository(uri, database, collectionName string) (UserRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	if err = client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	// İlgili koleksiyonun referansını alıyoruz.
	collection := client.Database(database).Collection(collectionName)

	return &userRepository{
		client:     client,
		collection: collection,
	}, nil
}

// CreateUser, domain.User tipindeki kullanıcıyı oluşturur.
// Burada, CreatedAt ve UpdatedAt alanlarını mevcut zamana ayarlıyoruz.
func (r *userRepository) CreateUser(ctx context.Context, user *domain.User) error {
	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err := r.collection.InsertOne(ctx, user)
	return err
}
