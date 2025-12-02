package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Client struct {
	client   *mongo.Client
	database *mongo.Database
	dbName   string
}

type Config struct {
	URI         string
	Database    string
	Timeout     time.Duration
	MaxPoolSize uint64
}

func NewClient(config Config) (*Client, error) {
	ctx, cancel := context.WithTimeout(context.Background(), config.Timeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.URI)

	if config.MaxPoolSize > 0 {
		clientOptions.SetMaxPoolSize(config.MaxPoolSize)
	}

	clientOptions.SetServerSelectionTimeout(5 * time.Second)
	clientOptions.SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	database := client.Database(config.Database)

	log.Printf("Connected to MongoDB database: %s", config.Database)

	return &Client{
		client:   client,
		database: database,
		dbName:   config.Database,
	}, nil
}

func (c *Client) GetDatabase() *mongo.Database {
	return c.database
}

func (c *Client) GetCollection(name string) *mongo.Collection {
	return c.database.Collection(name)
}

func (c *Client) GetClient() *mongo.Client {
	return c.client
}

func (c *Client) DatabaseName() string {
	return c.dbName
}

func (c *Client) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := c.client.Disconnect(ctx); err != nil {
		return fmt.Errorf("failed to close MongoDB connection: %w", err)
	}

	log.Println("MongoDB connection closed")
	return nil
}

func (c *Client) Ping(ctx context.Context) error {
	return c.client.Ping(ctx, nil)
}

func (c *Client) Find(ctx context.Context, collection string, filter interface{}, opts ...*options.FindOptions) ([]map[string]interface{}, error) {
	coll := c.GetCollection(collection)

	cursor, err := coll.Find(ctx, filter, opts...)
	if err != nil {
		return nil, fmt.Errorf("find failed: %w", err)
	}
	defer cursor.Close(ctx)

	var results []map[string]interface{}
	if err := cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("cursor decode failed: %w", err)
	}

	return results, nil
}

func (c *Client) FindOne(ctx context.Context, collection string, filter interface{}) (map[string]interface{}, error) {
	coll := c.GetCollection(collection)

	var result map[string]interface{}
	if err := coll.FindOne(ctx, filter).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("document not found")
		}
		return nil, fmt.Errorf("findone failed: %w", err)
	}

	return result, nil
}

func (c *Client) Count(ctx context.Context, collection string, filter interface{}) (int64, error) {
	coll := c.GetCollection(collection)

	count, err := coll.CountDocuments(ctx, filter)
	if err != nil {
		return 0, fmt.Errorf("count failed: %w", err)
	}

	return count, nil
}

func DefaultConfig() Config {
	return Config{
		URI:         "mongodb://localhost:27017",
		Database:    "dnd",
		Timeout:     10 * time.Second,
		MaxPoolSize: 100,
	}
}
