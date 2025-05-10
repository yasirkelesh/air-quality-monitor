package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/yasirkelesh/notification/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UserRepository tanımını arayüz (interface) üzerinden gerçekleştiriyoruz.
type UserRepository interface {
	CreateUser(ctx context.Context, user *domain.User) error
	Close() error
	FindUsersByRegion(regionGeohash string) ([]*domain.User, error)
	SaveNotification(notification *domain.Notification) error
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
	// Önce kullanıcının var olup olmadığını kontrol et
	filter := bson.M{"email": user.Email}
	existingUser := &domain.User{}
	err := r.collection.FindOne(ctx, filter).Decode(existingUser)
	if err == nil {
		return fmt.Errorf("bu email adresi ile kayıtlı kullanıcı zaten mevcut")
	}
	if err != mongo.ErrNoDocuments {
		return fmt.Errorf("kullanıcı kontrolü sırasında hata: %v", err)
	}

	now := time.Now().UTC()
	user.CreatedAt = now
	user.UpdatedAt = now

	_, err = r.collection.InsertOne(ctx, user)
	return err
}

func (r *userRepository) FindUsersByRegion(regionGeohash string) ([]*domain.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{
		"geohash": bson.M{
			"$regex":   "^" + regionGeohash,
			"$options": "i",
		},
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []*domain.User
	if err = cursor.All(ctx, &users); err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) SaveNotification(notification *domain.Notification) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, notification)
	return err
}
