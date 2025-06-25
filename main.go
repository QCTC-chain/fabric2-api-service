package main

import (
	"fmt"
	"github.com/apache/rocketmq-clients/golang/v5"
	"github.com/apache/rocketmq-clients/golang/v5/credentials"
	"github.com/qctc/fabric2-api-server/define"
	"github.com/qctc/fabric2-api-server/router"
	"github.com/qctc/fabric2-api-server/service"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
)

//	func main() {
//		router := router.SetUpRouter()
//		log.Println("Starting server on port 8080")
//		log.Fatal(http.ListenAndServe(":8080", router))
//	}

func init() {
	// 读取配置文件
	data, err := ioutil.ReadFile("./config.yaml")
	if err != nil {
		log.Fatalf("无法读取配置文件: %v", err)
	}

	// 解析配置文件
	var config *define.Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		log.Fatalf("无法解析配置文件: %v", err)
	}

	// 设置全局配置
	define.GlobalConfig = config
	mqConfig := define.GlobalConfig.MQ
	//配置mq
	golang.ResetLogger()
	define.GlobalProducer, err = golang.NewProducer(&golang.Config{
		Endpoint: fmt.Sprintf("%s:%d", mqConfig.Host, mqConfig.Port),
		Credentials: &credentials.SessionCredentials{
			AccessKey:    "",
			AccessSecret: "",
		},
	},
		golang.WithTopics(mqConfig.Topic),
	)
	if err != nil {
		log.Fatalf("Failed to initialize RocketMQ producer: %v", err)
	}
	err = define.GlobalProducer.Start()
	if err != nil {
		log.Fatalf("Failed to start RocketMQ producer: %v", err)
	}
}

func main() {
	port := define.GlobalConfig.Server.Port
	useRouter := router.SetUpRouter()
	defer func() {
		log.Println("开始执行清理任务...")

		// 关闭 RocketMQ Producer
		if define.GlobalProducer != nil {
			if err := define.GlobalProducer.GracefulStop(); err != nil {
				log.Printf("RocketMQ Producer 关闭失败: %v", err)
			} else {
				log.Println("RocketMQ Producer 已关闭")
			}
		}

		// 取消所有事件订阅
		unsubscribeAll()

		log.Println("清理任务完成，服务即将退出。")
	}()

	log.Printf("服务器正在端口 %d 上运行...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), useRouter))
}

func unsubscribeAll() {
	define.SubscriptionMutex.RLock()
	defer define.SubscriptionMutex.RUnlock()

	for key, regID := range define.EventSubscriptions {
		parts := strings.Split(key, "-")
		if len(parts) < 3 {
			continue
		}
		sdkID := parts[0]
		sdk, _ := service.Fabric2ServicePool[sdkID]
		_ = sdk.UnsubscribeEvent(regID)
		log.Printf("已取消订阅: %s", key)
	}
}
