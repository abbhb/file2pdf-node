package mqcode

import (
	"context"
	"encoding/json"
	"fmt"
	rmq_client "github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/libreofficedocker/unoserver-rest-api/depot"
	"github.com/libreofficedocker/unoserver-rest-api/typeall"
	"github.com/libreofficedocker/unoserver-rest-api/unoconvert"
	model "github.com/unidoc/unipdf/v3/model"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
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

func (consumer *Consumer) StartConsumer() {
	// log to console
	os.Setenv("mq.consoleAppender.enabled", "true")
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
	defer simpleConsumer.GracefulStop()
	// Create a channel to signal stopping
	consumer.stopChan = make(chan struct{})
	var wg sync.WaitGroup

	go func() {
		for {
			select {
			case <-consumer.stopChan:
				break
			default:
				fmt.Println("start recevie message")
				mvs, err := simpleConsumer.Receive(context.TODO(), consumer.maxMessageNum, consumer.invisibleDuration)
				if err != nil {
					fmt.Println(err)
				}
				// ack message
				for _, mv := range mvs {
					// 每条消息
					simpleConsumer.Ack(context.TODO(), mv)
					// 实例化结构体
					needToPDFObject := typeall.PrintDataFileToPDFReq{}
					err := json.Unmarshal(mv.GetBody(), &needToPDFObject)
					if err != nil {
						log.Fatal(err)
						// 统一异常生产消息
						// 结构体无法解析json
						log.Printf("无法解析json:%s", string(mv.GetBody()))
						continue
					}
					if needToPDFObject.Id == nil {
						log.Fatal("任务id不存在")
						continue
					}
					if needToPDFObject.FileUrl == nil {
						log.Fatal("任务FileUrl不存在")
						consumer.producer.SendError(needToPDFObject.Id, "任务FileUrl不存在")
						continue
					}
					if needToPDFObject.FilePDFUrl == nil {
						log.Fatal("任务FilePDFUrl不存在")
						consumer.producer.SendError(needToPDFObject.Id, "任务FilePDFUrl不存在")
						continue
					}
					if needToPDFObject.FilePDFUploadUrl == nil {
						log.Fatal("任务FilePDFUploadUrl不存在")
						consumer.producer.SendError(needToPDFObject.Id, "任务FilePDFUploadUrl不存在")
						continue
					}
					log.Printf("json是正确的")
					filename := filepath.Base(*needToPDFObject.FileUrl)
					// 创建临时文件
					tempFile, err := os.CreateTemp(depot.WorkDir, filename)
					if err != nil {
						fmt.Printf("Failed to create temp file: %v\n", err)
						consumer.producer.SendError(needToPDFObject.Id, "Failed to create temp file")
						return
					}
					defer tempFile.Close()

					// 获取 HTTP 响应
					resp, err := http.Get(*needToPDFObject.FileUrl)
					if err != nil {
						fmt.Printf("Failed to download file: %v\n", err)
						consumer.producer.SendError(needToPDFObject.Id, "Failed to download file")
						return
					}
					defer resp.Body.Close()

					// 将响应体写入临时文件
					_, err = io.Copy(tempFile, resp.Body)
					if err != nil {
						consumer.producer.SendError(needToPDFObject.Id, "Failed to save file")
						fmt.Printf("Failed to save file: %v\n", err)
						return
					}
					// 获取临时文件的路径
					tempFilePath := tempFile.Name()
					fmt.Printf("File downloaded to: %s\n", tempFilePath)
					defer func() {
						err := os.Remove(tempFilePath)
						if err != nil {
							log.Println("Delege temp file failed", err)
						}
					}()
					// Prepare output file path
					outFile, err := os.CreateTemp(depot.WorkDir, filename+".pdf")
					if err != nil {
						consumer.producer.SendError(needToPDFObject.Id, "Create temp file failed2")
						log.Println("Create temp file failed2", err)
						return
					}
					defer func() {
						err := os.Remove(outFile.Name())
						if err != nil {
							log.Println("Delege temp file failed", err)
						}
					}()
					// Run unoconvert command with options
					// If context timeout is 0s run without timeout
					if unoconvert.ContextTimeout == 0 {
						err = unoconvert.Run(tempFilePath, outFile.Name())
					} else {
						err = unoconvert.RunContext(context.Background(), tempFilePath, outFile.Name())
					}

					log.Printf("Processing: %s %s", tempFilePath, outFile.Name())
					if err != nil {
						consumer.producer.SendError(needToPDFObject.Id, "unoconvert error")
						log.Printf("unoconvert error: %s", err)
						return
					}
					// 到此处应该是成功了
					resultChan := make(chan typeall.PageCountResult)
					// 启动一个新的协程来处理 PDF 文件
					wg.Add(1)
					go func() {
						defer wg.Done()

						// 打开 PDF 文件
						reader, file, err := model.NewPdfReaderFromFile(outFile.Name(), &model.ReaderOpts{})
						if err != nil {
							resultChan <- typeall.PageCountResult{Err: err}
							return
						}
						defer func(file *os.File) {
							err := file.Close()
							if err != nil {

							}
						}(file)

						// 获取 PDF 文件的页数
						numPages, err := reader.GetNumPages()
						if err != nil {
							resultChan <- typeall.PageCountResult{Err: err}
							return
						}

						// 发送结果到通道
						resultChan <- typeall.PageCountResult{NumPages: numPages}
					}()

					// 打开文件
					file, err := os.Open(outFile.Name())
					if err != nil {
						log.Printf("failed to open file2: %v", err)
						consumer.producer.SendError(needToPDFObject.Id, "failed to open file2")

						return
					}
					defer file.Close()

					// 获取文件信息
					_, err = file.Stat()
					if err != nil {
						log.Printf("failed to get file info: %v", err)
						consumer.producer.SendError(needToPDFObject.Id, "failed to get file info:")
						return
					}

					// 创建 HTTP 请求
					request, err := http.NewRequest("PUT", *needToPDFObject.FilePDFUploadUrl, file)
					if err != nil {
						log.Printf("failed to create request: %v", err)
						consumer.producer.SendError(needToPDFObject.Id, "failed to create request:")
						return
					}
					// 设置 Content-Type
					request.Header.Set("Content-Type", "application/octet-stream")

					// 发送请求
					client := &http.Client{}
					response, err := client.Do(request)
					if err != nil {
						log.Printf("failed to send request: %v", err)
						consumer.producer.SendError(needToPDFObject.Id, "failed to send request")
						return
					}
					defer response.Body.Close()
					// 检查响应状态
					if response.StatusCode != http.StatusOK {
						log.Printf("failed to upload file: %s", response.Status)
						consumer.producer.SendError(needToPDFObject.Id, "failed to upload file")
						return
					}
					log.Printf("File uploaded successfully")
					// Send成功消息
					// 等待协程完成
					wg.Wait()
					// 关闭通道
					defer close(resultChan)
					result := <-resultChan
					if result.Err != nil {
						log.Printf("Not Get Pages")
					}
					status := new(int)
					*status = 1
					// 发送消息失败的直接丢掉
					printDataFromPDFResp := &typeall.PrintDataFromPDFResp{
						Id:         needToPDFObject.Id,
						FilePDFUrl: needToPDFObject.FilePDFUrl,
						PageNums:   &result.NumPages,
						Status:     status,
						Message:    nil,
					}
					printData, err := json.Marshal(printDataFromPDFResp)
					err = consumer.producer.Send(printData)
					if err != nil {
						log.Printf("failed to send print data: %v", err)
						consumer.producer.SendError(needToPDFObject.Id, "failed to send print data")
						return
					}
				}
			}
		}
	}()
	// run for a while
	// Block until stop is called
	<-consumer.stopChan
}

// 添加停止消费的方法
func (consumer *Consumer) stopConsumer() {
	close(consumer.stopChan)
}
