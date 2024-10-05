package mqcode

import (
	"context"
	"encoding/json"
	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/libreofficedocker/unoserver-rest-api/typeall"
	"log"
)

type Producer struct {
	ProducerCli golang.Producer
}

const (
	Topic     = "print_filetopdf_send_msg_r"
	GroupName = "print_filetopdf_send_msg_r_group"
	Endpoint  = "192.168.12.12:9081"
	AccessKey = ""
	SecretKey = ""
)

func CreateProducerCli() golang.Producer {

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
	log.Println("producer is running")
	return producer
}

func CreateProducer() *Producer {
	return &Producer{
		ProducerCli: CreateProducerCli(),
	}
}

func (p *Producer) Send(message []byte) error {
	if p.ProducerCli == nil {
		p.ProducerCli = CreateProducerCli()
	}
	resp, err := p.ProducerCli.Send(context.TODO(), &golang.Message{
		Topic: Topic,
		Body:  message,
	})
	if err != nil {
		return err
	}
	for i := 0; i < len(resp); i++ {
		log.Printf("%#v\n", resp[i])
	}
	return nil
}

func (p *Producer) SendError(id *string, message string) {
	log.Printf("发送异常消息:id:{%s},message:{%s}", id, message)
	if p.ProducerCli == nil {
		p.ProducerCli = CreateProducerCli()
	}
	status := new(int)
	*status = 0
	errorMsgObject := typeall.PrintDataFromPDFResp{
		Id:      id,
		Status:  status,
		Message: &message,
	}
	errorMsg, err := json.Marshal(errorMsgObject)

	resp, err := p.ProducerCli.Send(context.TODO(), &golang.Message{
		Topic: Topic,
		Body:  errorMsg,
	})
	if err != nil {
		log.Fatal(err)
	}
	for i := 0; i < len(resp); i++ {
		log.Printf("%#v\n", resp[i])
	}
}

func (p *Producer) Close() error {
	if p.ProducerCli != nil {
		err := p.ProducerCli.GracefulStop()
		if err != nil {
			return err
		}
	}
	return nil
}
