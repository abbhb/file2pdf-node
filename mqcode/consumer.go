package mqcode

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/abbhb/filel2pdf-node/depot"
	"github.com/abbhb/filel2pdf-node/typeall"
	"github.com/abbhb/filel2pdf-node/unoconvert"
	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"

	model "github.com/unidoc/unipdf/v3/model"
	"gopkg.in/resty.v1"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

type Consumer struct {
	Topic             string
	GroupName         string
	Endpoint          string
	AccessKey         string
	SecretKey         string
	awaitDuration     time.Duration
	maxMessageNum     int32
	invisibleDuration time.Duration
	stopChan          chan struct{} // 将 stopChan 添加到 Consumer 结构体中
	producer          *Producer
}

func CreateConsumer(producer *Producer) *Consumer {
	return &Consumer{
		Topic:             "print_filetopdf_send_msg",
		GroupName:         "print_filetopdf_send_msg_group",
		Endpoint:          "192.168.12.12:9081",
		AccessKey:         "",
		SecretKey:         "",
		awaitDuration:     time.Second * 20,
		maxMessageNum:     1,
		invisibleDuration: time.Second * 120, // 最大一条消息2分钟执行完，没见过什么文件2分钟转换不完
		producer:          producer,
	}
}
func (consumer *Consumer) handel(needToPDFObject *typeall.PrintDataFileToPDFReq) (*typeall.PrintDataFromPDFResp, error) {
	if needToPDFObject.Id == nil {
		return nil, errors.New("任务id不存在")
	}
	if needToPDFObject.FileUrl == nil {
		return nil, errors.New("任务FileUrl不存在")
	}
	if needToPDFObject.FilePDFUrl == nil {
		return nil, errors.New("任务FilePDFUrl不存在")
	}
	if needToPDFObject.FilePDFUploadUrl == nil {
		return nil, errors.New("任务FilePDFUploadUrl不存在")
	}
	log.Printf("json是正确的")
	filename := filepath.Base(*needToPDFObject.FileUrl)
	// 创建临时文件
	tempFile, err := os.CreateTemp(depot.WorkDir, "file*"+filename)
	if err != nil {
		// todo:bug 文件名长到一定程度会报错
		return nil, errors.New("Failed to create temp file")
	}
	defer func(tempFile *os.File) {
		err := tempFile.Close()
		if err != nil {

		}
	}(tempFile)
	// 获取 HTTP 响应
	resp, err := http.Get(*needToPDFObject.FileUrl)
	if err != nil {
		fmt.Printf("Failed to download file: %v\n", err)
		return nil, errors.New("Failed to download file")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	// 将响应体写入临时文件
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		fmt.Printf("Failed to save file: %v\n", err)
		return nil, errors.New("Failed to save file")
	}

	// 获取临时文件的路径
	tempFilePath := tempFile.Name()
	fmt.Printf("File downloaded to: %s\n", tempFilePath)
	// Prepare output file path
	outFile, err := os.CreateTemp(depot.WorkDir, "file*"+filename+".pdf")
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(tempFilePath)

	if err != nil {
		log.Println("Create temp file failed2", err)
		return nil, errors.New("Create temp file failed2")
	}
	outFileName := outFile.Name()
	defer func(name string) {
		err := os.Remove(name)
		if err != nil {

		}
	}(outFileName)
	// Run unoconvert command with options
	// If context timeout is 0s run without timeout
	log.Printf("路径")
	err = unoconvert.Run(tempFilePath, outFileName)
	if err != nil {
		log.Printf("unoconvert error: %s", err)
		return nil, errors.New("unoconvert error")
	}
	// 到此处应该是成功了

	// 打开 PDF 文件
	reader, file2, err := model.NewPdfReaderFromFile(outFileName, &model.ReaderOpts{})
	if err != nil {
		return nil, errors.New("create reader failed")
	}

	// 获取 PDF 文件的页数
	numPages, err := reader.GetNumPages()
	if err != nil {
		return nil, errors.New("get num pages failed")
	}
	log.Printf("协程运行结束")
	err = file2.Close()
	if err != nil {
		return nil, err
	}

	// 打开文件
	file, err := os.ReadFile(outFileName)
	if err != nil {
		log.Printf("failed to open file2: %v", err)
		return nil, errors.New("failed to open file2")
	}

	// Create a Resty Client
	client := resty.New()
	// Request goes as JSON content type
	// No need to set auth token, error, if you have client level settings
	response, err := client.R().
		SetBody(file).
		Put(*needToPDFObject.FilePDFUploadUrl)
	if err != nil {
		log.Printf("failed to send request: %v", err)
		return nil, errors.New("failed to send request")
	}

	// 检查响应状态
	if response.StatusCode() != http.StatusOK {
		return nil, errors.New("failed to upload file")
	}

	log.Printf("File uploaded successfully")
	// Send成功消息

	status := new(int)
	*status = 1
	// 发送消息失败的直接丢掉
	printDataFromPDFResp := &typeall.PrintDataFromPDFResp{
		Id:         needToPDFObject.Id,
		FilePDFUrl: needToPDFObject.FilePDFUrl,
		PageNums:   &numPages,
		Status:     status,
		Message:    nil,
	}

	return printDataFromPDFResp, nil
}
func (consumer *Consumer) StartConsumer() {
	// log to console
	err := os.Setenv("mq.consoleAppender.enabled", "true")
	if err != nil {
		return
	}
	rmq_client.ResetLogger()
	// new simpleConsumer instance
	simpleConsumer, err := rmq_client.NewSimpleConsumer(&rmq_client.Config{
		Endpoint:      consumer.Endpoint,
		ConsumerGroup: consumer.GroupName,
		Credentials: &credentials.SessionCredentials{
			AccessKey:    consumer.AccessKey,
			AccessSecret: consumer.SecretKey,
		},
	},
		rmq_client.WithAwaitDuration(consumer.awaitDuration), // 每20s拉取一次
		rmq_client.WithSubscriptionExpressions(map[string]*rmq_client.FilterExpression{
			consumer.Topic: rmq_client.NewFilterExpression("req"),
		}),
	)
	if err != nil {
		log.Fatal(err)
	}
	// start simpleConsumer
	err = simpleConsumer.Start()
	if err != nil {
		log.Fatal(err)
	}
	// gracefule stop simpleConsumer
	defer func(simpleConsumer rmq_client.SimpleConsumer) {
		err := simpleConsumer.GracefulStop()
		if err != nil {

		}
	}(simpleConsumer)
	// Create a channel to signal stopping

	for {
		fmt.Println("start recevie message")
		mvs, err := simpleConsumer.Receive(context.TODO(), consumer.maxMessageNum, consumer.invisibleDuration)
		if err != nil {
			fmt.Println(err)
		}
		// ack message
		for _, mv := range mvs {
			// 每条消息
			err := simpleConsumer.Ack(context.TODO(), mv)
			if err != nil {
				log.Printf("消息确认失败")
				log.Printf(err.Error())
				continue
			}
			// 实例化结构体
			needToPDFObject := typeall.PrintDataFileToPDFReq{}
			err = json.Unmarshal(mv.GetBody(), &needToPDFObject)
			if err != nil {
				log.Printf(err.Error())
				// 统一异常生产消息
				// 结构体无法解析json
				log.Printf("无法解析json:%s", string(mv.GetBody()))
				continue
			}
			printDataFromPDFResp, err := consumer.handel(&needToPDFObject)
			if err != nil {
				if needToPDFObject.Id != nil {
					consumer.producer.SendError(needToPDFObject.Id, err.Error())
					log.Printf("统一异常处理  异常:%s", err.Error())

				}
				continue
			}

			err = consumer.producer.Send(printDataFromPDFResp)
			log.Printf("消息发送成功")
			if err != nil {
				log.Printf("failed to send print data: %v", err)
				consumer.producer.SendError(needToPDFObject.Id, "failed to send print data")
				continue
			}

		}
	}
	// run for a while
	// Block until stop is called

}
