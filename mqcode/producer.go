package mqcode

import (
	"context"
	"encoding/json"
	"github.com/abbhb/filel2pdf-node/typeall"
	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
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

// send 一定要带tag，同一个消费者组要求监听的tag一致
func (p *Producer) Send(message *typeall.PrintDataFromPDFResp) error {
	if p.ProducerCli == nil {
		p.ProducerCli = CreateProducerCli()
	}
	printData, err := json.Marshal(message)
	log.Printf("正确处理")
	tag := new(string)
	*tag = "resp"
	resp, err := p.ProducerCli.Send(context.TODO(), &golang.Message{
		Topic: Topic,
		Body:  printData,
		Tag:   tag,
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
	tag := new(string)
	*tag = "resp"
	resp, err := p.ProducerCli.Send(context.TODO(), &golang.Message{
		Topic: Topic,
		Body:  errorMsg,
		Tag:   tag,
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
