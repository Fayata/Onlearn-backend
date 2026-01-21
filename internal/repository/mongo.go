package repository

import (
	"context"
	"errors"
	"onlearn-backend/internal/domain"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type moduleRepo struct {
	db *mongo.Database
}

func NewModuleRepository(db *mongo.Database) domain.ModuleRepository {
	return &moduleRepo{db}
}

func (r *moduleRepo) Create(ctx context.Context, module *domain.Module) error {
	collection := r.db.Collection("modules")

	// Set created_at if not set
	if module.CreatedAt.IsZero() {
		module.CreatedAt = time.Now()
	}

	// Generate ObjectID if not exists
	if module.ID == "" {
		result, err := collection.InsertOne(ctx, module)
		if err != nil {
			return err
		}
		module.ID = result.InsertedID.(primitive.ObjectID).Hex()
		return nil
	}

	_, err := collection.InsertOne(ctx, module)
	return err
}

func (r *moduleRepo) GetByCourseID(ctx context.Context, courseID uint) ([]domain.Module, error) {
	collection := r.db.Collection("modules")
	filter := bson.M{"course_id": courseID}

	// Sort by order ascending
	opts := options.Find().SetSort(bson.D{{Key: "order", Value: 1}})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var modules []domain.Module
	if err := cursor.All(ctx, &modules); err != nil {
		return nil, err
	}

	// Convert ObjectID to string for each module
	for i := range modules {
		if modules[i].ID == "" && cursor.Current != nil {
			if id, ok := cursor.Current.Lookup("_id").ObjectIDOK(); ok {
				modules[i].ID = id.Hex()
			}
		}
	}

	return modules, nil
}

func (r *moduleRepo) GetByID(ctx context.Context, id string) (*domain.Module, error) {
	collection := r.db.Collection("modules")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid module ID")
	}

	filter := bson.M{"_id": objID}

	var module domain.Module
	err = collection.FindOne(ctx, filter).Decode(&module)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("module not found")
		}
		return nil, err
	}

	module.ID = objID.Hex()
	return &module, nil
}

func (r *moduleRepo) Update(ctx context.Context, module *domain.Module) error {
	collection := r.db.Collection("modules")

	objID, err := primitive.ObjectIDFromHex(module.ID)
	if err != nil {
		return errors.New("invalid module ID")
	}

	filter := bson.M{"_id": objID}
	update := bson.M{
		"$set": bson.M{
			"title":       module.Title,
			"type":        module.Type,
			"content_url": module.ContentURL,
			"quiz_link":   module.QuizLink,
			"description": module.Description,
			"order":       module.Order,
		},
	}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("module not found")
	}

	return nil
}

func (r *moduleRepo) Delete(ctx context.Context, id string) error {
	collection := r.db.Collection("modules")

	objID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid module ID")
	}

	filter := bson.M{"_id": objID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return errors.New("module not found")
	}

	return nil
}
