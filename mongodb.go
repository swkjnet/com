//数据库
package com

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type H_Mgo struct {
	client *mongo.Client
	ctx    context.Context
	cancel context.CancelFunc
}

//初始化mgo句柄(不支持并发)
func InitMgoDB(dburl string) (*H_Mgo, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(dburl))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return &H_Mgo{client, ctx, cancel}, nil
}

//获取表句柄
func (this *H_Mgo) getCollection(db, col string) *mongo.Collection {
	return this.client.Database(db).Collection(col)
}

//插入数据
func (this *H_Mgo) Insert(db string, col string, v interface{}) error {
	_, err := this.getCollection(db, col).InsertOne(this.ctx, v)
	return err
}

//根据条件查询一条记录
func (this *H_Mgo) FindOne(db string, col string, filter interface{}, result interface{}) error {
	return this.getCollection(db, col).FindOne(this.ctx, filter).Decode(&result)
}

//根据条件查询所有记录
func (this *H_Mgo) FindAll(db string, col string, filter interface{}, result interface{}) error {
	cursor, err := this.getCollection(db, col).Find(this.ctx, filter)
	if err != nil {
		return err
	}
	return cursor.All(this.ctx, &result)
}

//根据条件查询一条记录
func (this *H_Mgo) FindId(db string, col string, id interface{}, result interface{}) error {
	return this.FindOne(db, col, bson.M{"_id": id}, result)
}

//创建索引
func (this *H_Mgo) CreateIndex(dbname, colname string, indexname string, keys bson.D, unique bool, dropDups bool) error {
	optionindex := options.Index()
	optionindex.SetBackground(true)
	optionindex.SetUnique(unique)
	optionindex.SetName(indexname)
	model := mongo.IndexModel{Keys: keys, Options: optionindex}
	_, err := this.getCollection(dbname, colname).Indexes().CreateOne(this.ctx, model)
	return err
}

//更新数据$set
func (this *H_Mgo) UpdateManySet(dbname, colname string, filter interface{}, v interface{}, opts ...*options.UpdateOptions) error {
	_, err := this.getCollection(dbname, colname).UpdateMany(this.ctx, filter, bson.D{{"$set", v}}, opts...)
	return err
}

//更新数据，不限制操作($inc等)
func (this *H_Mgo) UpdateManyAll(dbname, colname string, filter interface{}, v interface{}, opts ...*options.UpdateOptions) error {
	_, err := this.getCollection(dbname, colname).UpdateMany(this.ctx, filter, v, opts...)
	return err
}

//根据id更新数据
func (this *H_Mgo) UpdateById(dbname, colname string, id interface{}, v interface{}, opts ...*options.UpdateOptions) error {
	_, err := this.getCollection(dbname, colname).UpdateOne(this.ctx, bson.D{{"_id", id}}, v, opts...)
	return err
}

//查询记录数
func (this *H_Mgo) FindCount(dbname, colname string, filter interface{}) (int64, error) {
	return this.getCollection(dbname, colname).CountDocuments(this.ctx, filter)
}
