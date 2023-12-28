package service

import (
	"context"
	"fmt"
	"gin-chat-demo/conf"
	"gin-chat-demo/model/ws"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

func InsertMsg(database, id string, content string, read uint, expire int64) error {
	// 插入MongoDB中
	collection := conf.MongoDBClient.Database(database).Collection(id)
	comment := ws.Trainer{
		Content:   content,
		StartTime: time.Now().Unix(),
		EndTime:   time.Now().Unix() + expire,
		Read:      read,
	}
	_, err := collection.InsertOne(context.TODO(), comment)
	return err
}
func FindMany(database, sendID, id string, pageSize int) (results []ws.Result, err error) {
	var resultMe []ws.Trainer
	var resultYou []ws.Trainer
	sendIDCollection := conf.MongoDBClient.Database(database).Collection(sendID)
	idCollection := conf.MongoDBClient.Database(database).Collection(id)
	sendIDTimeCurcor, err := sendIDCollection.Find(context.TODO(),
		options.Find().SetSort(bson.D{{"startTime", -1}}),
		options.Find().SetLimit(int64(pageSize)),
	)
	idTimeCurcor, err := idCollection.Find(context.TODO(),
		options.Find().SetSort(bson.D{{"startTime", -1}}),
		options.Find().SetLimit(int64(pageSize)),
	)
	err = sendIDTimeCurcor.All(context.TODO(), &resultYou)
	err = idTimeCurcor.All(context.TODO(), &resultMe)

	results, _ = AppendAndSort(resultMe, resultYou)
	return
}

type SendSortMsg struct {
	Content  string `json:"content"`
	Read     uint   `json:"read"`
	CreateAt int64  `json:"create_at"`
}

func AppendAndSort(resultMe, resultYou []ws.Trainer) (results []ws.Result, err error) {
	for _, r := range resultMe {
		// 构造函数返回msg
		sendSortMsg := SendSortMsg{
			Content:  r.Content,
			Read:     r.Read,
			CreateAt: r.StartTime,
		}
		// 构造函数返回所有的内容
		result := ws.Result{
			StartTime: r.StartTime,
			Msg:       fmt.Sprintf("%v", sendSortMsg),
			From:      "me",
		}
		results = append(results, result)
	}
	for _, r := range resultYou {
		// 构造函数返回msg
		sendSortMsg := SendSortMsg{
			Content:  r.Content,
			Read:     r.Read,
			CreateAt: r.StartTime,
		}
		// 构造函数返回所有的内容
		result := ws.Result{
			StartTime: r.StartTime,
			Msg:       fmt.Sprintf("%v", sendSortMsg),
			From:      "you",
		}
		results = append(results, result)
	}
	return results, nil
}
