package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/apache/rocketmq-clients/golang/v5"
	_ "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/libreofficedocker/unoserver-rest-api/typeall"
	"log"
	"os"
)

const (
	Topic     = "asdfwe_123213"
	GroupName = "asdfwe_123213_group"
	Endpoint  = "192.168.12.12:9081"
	AccessKey = ""
	SecretKey = ""
)

func main() {
	os.Setenv("mq.consoleAppender.enabled", "true")
	golang.ResetLogger()
	// new producer instance
	producer, err := golang.NewProducer(&golang.Config{
		Endpoint:      Endpoint,
		ConsumerGroup: GroupName,
		Credentials: &credentials.SessionCredentials{
			AccessKey:    AccessKey,
			AccessSecret: SecretKey,
		},
	},
		golang.WithTopics(Topic),
	)
	if err != nil {
		log.Fatal(err)
	}
	// start producer
	err = producer.Start()
	if err != nil {
		log.Fatal(err)
	}
	// gracefule stop producer
	defer producer.GracefulStop()

	for i := 0; i < 100; i++ {
		tsss := typeall.TestS{
			Name:  "test",
			Type:  1,
			Fsa:   false,
			Index: i,
		}
		jsonStudent, err := json.Marshal(tsss)

		// new a message
		msg := &golang.Message{
			Topic: Topic,
			Body:  jsonStudent,
		}
		// set keys and tag
		msg.SetKeys("a", "b")
		msg.SetTag("ab")
		// send message in sync
		resp, err := producer.Send(context.TODO(), msg)
		if err != nil {
			log.Fatal(err)
		}
		for i := 0; i < len(resp); i++ {
			fmt.Printf("%#v\n", resp[i])
		}

	}
}
