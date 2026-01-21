package repository

import (
	"context"
	"onlearn-backend/internal/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type moduleRepo struct {
	db *mongo.Database
}

func NewModuleRepository(db *mongo.Database) domain.ModuleRepository {
	return &moduleRepo{db}
}

func (r *moduleRepo) Create(ctx context.Context, module *domain.Module) error {
	collection := r.db.Collection("modules")
	_, err := collection.InsertOne(ctx, module)
	return err
}

func (r *moduleRepo) GetByCourseID(ctx context.Context, courseID uint) ([]domain.Module, error) {
	collection := r.db.Collection("modules")
	// MongoDB menyimpan course_id sebagai int/int64, kita pastikan query nya benar
	filter := bson.M{"course_id": courseID}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var modules []domain.Module
	if err := cursor.All(ctx, &modules); err != nil {
		return nil, err
	}
	return modules, nil
}

func (r *moduleRepo) Delete(ctx context.Context, id string) error {
	objID, _ := primitive.ObjectIDFromHex(id)
	_, err := r.db.Collection("modules").DeleteOne(ctx, bson.M{"_id": objID})
	return err
}
