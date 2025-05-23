package service

import (
	"context"
	"errors"
	"go-graphql-blog/graph/database"
	"go-graphql-blog/graph/model"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type BlogService struct{}

const BLOG_COLLECTION = "blog"

func (b *BlogService) GetAllBlogs() []*model.Blog {

	// create a query to get all blog data
	var query primitive.D = bson.D{{}}

	// create a find option to order blog data by createdAt
	var findOptions *options.FindOptions = options.Find()

	findOptions.SetSort(bson.D{{Key: "createdAt", Value: -1}})

	cursor, err := database.GetCollection(BLOG_COLLECTION).
		Find(context.TODO(), query, findOptions)

	if err != nil {
		return []*model.Blog{}
	}

	var blogs []*model.Blog = make([]*model.Blog, 0)

	if err := cursor.All(context.TODO(), &blogs); err != nil {
		return []*model.Blog{}
	}

	return blogs
}

func (b *BlogService) GetBlogByID(id string) (*model.Blog, error) {
	blogID, err := primitive.ObjectIDFromHex(id)

	if err != nil {
		return &model.Blog{}, errors.New("id is invalid")
	}

	var query primitive.D = bson.D{{Key: "_id", Value: blogID}}

	var collection *mongo.Collection = database.GetCollection(BLOG_COLLECTION)

	var blogData *mongo.SingleResult = collection.
		FindOne(context.TODO(), query)

	if blogData.Err() != nil {
		return &model.Blog{}, errors.New("blog not found")
	}

	var blog *model.Blog = &model.Blog{}

	blogData.Decode(blog)

	return blog, nil
}

func (b *BlogService) CreateBlog(input model.NewBlog, user model.User) (*model.Blog, error) {
	var blog model.Blog = model.Blog{
		Title:     input.Title,
		Content:   input.Content,
		Author:    &user,
		CreatedAt: time.Now(),
	}

	var collection *mongo.Collection = database.GetCollection(BLOG_COLLECTION)

	result, err := collection.InsertOne(context.TODO(), blog)

	if err != nil {
		return &model.Blog{}, errors.New("create blog failed")
	}

	var filter primitive.D = bson.D{{Key: "_id", Value: result.InsertedID}}

	var createdRecord *mongo.SingleResult = collection.FindOne(context.TODO(), filter)

	var createdBlog *model.Blog = &model.Blog{}

	createdRecord.Decode(createdBlog)

	return createdBlog, nil
}

func (b *BlogService) EditBlog(input model.EditBlog, user model.User) (*model.Blog, error) {
	blogId, err := primitive.ObjectIDFromHex(input.BlogID)

	if err != nil {
		return &model.Blog{}, errors.New("id is invalid")
	}

	var query primitive.D = bson.D{
		{Key: "_id", Value: blogId},
		{Key: "author._id", Value: user.ID},
	}

	var update primitive.D = bson.D{{
		Key: "$set",
		Value: bson.D{
			{Key: "title", Value: input.Title},
			{Key: "content", Value: input.Content},
			{Key: "updatedAt", Value: time.Now()},
		},
	}}

	var collection *mongo.Collection = database.GetCollection(BLOG_COLLECTION)

	var updateResult *mongo.SingleResult = collection.FindOneAndUpdate(
		context.TODO(),
		query,
		update,
		options.FindOneAndUpdate().SetReturnDocument(options.After),
	)

	if err := updateResult.Err(); err != nil {
		if err == mongo.ErrNoDocuments {
			return &model.Blog{}, errors.New("bog not found")
		}
		return &model.Blog{}, errors.New("update blog failed")
	}

	var editedBlog *model.Blog = &model.Blog{}

	updateResult.Decode(editedBlog)

	return editedBlog, nil
}

func (b *BlogService) DeleteBlog(input model.DeleteBlog, user model.User) bool {
	blogID, err := primitive.ObjectIDFromHex(input.BlogID)

	if err != nil {
		return false
	}

	var query primitive.D = bson.D{
		{Key: "_id", Value: blogID},
		{Key: "author._id", Value: user.ID},
	}

	var collection *mongo.Collection = database.GetCollection(BLOG_COLLECTION)

	result, err := collection.DeleteOne(context.TODO(), query)

	var isFailed bool = err != nil || result.DeletedCount < 1

	return !isFailed

}
