package mongoauth

import (
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	_ "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"ssov2/internal/domain/models"
	"time"
)

type User struct {
	Email      string `bson:"email"`
	PasswdHash []byte `bson:"passwd_hash"`
	isAdmin    bool   `bson:"is_admin"`
}

type ProvideUser struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Email      string             `bson:"email"`
	PasswdHash []byte             `bson:"passwd_hash"`
	isAdmin    bool               `bson:"is_admin"`
}

const collectionName = "users"
const dbName = "sso"

type Storage struct {
	collection *mongo.Collection
}

func New(uri string) (*Storage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	collection := client.Database(dbName).Collection(collectionName)
	return &Storage{collection: collection}, nil
}

func (s *Storage) SaveUser(ctx context.Context, email string, passwdHash []byte) (userID int64, err error) {
	user := User{
		Email:      email,
		PasswdHash: passwdHash,
	}

	_, err = s.collection.InsertOne(ctx, user)
	if err != nil {
		return 0, fmt.Errorf("failed to save user: %v", err)
	}
	return 1, err
}

func (s *Storage) User(ctx context.Context, email string) (models.User, error) {
	const op = "storage.sqlite.User"

	filter := bson.M{"email": email}
	var user ProvideUser
	err := s.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return models.User{}, nil
		}
		return models.User{}, fmt.Errorf("%s: %w", op, err)
	}

	return models.User{
		ID:       user.ID.Hex(),
		Email:    user.Email,
		PassHash: user.PasswdHash,
	}, nil
}
