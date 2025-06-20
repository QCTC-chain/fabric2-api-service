package main

import (
	"fmt"
	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/qctc/fabric2-api-server/define"
	"github.com/qctc/fabric2-api-server/router"
	"github.com/qctc/fabric2-api-server/utils"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"log"
	"net/http"
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
	define.GlobalProducer, err = rocketmq.NewProducer(
		producer.WithGroupName(mqConfig.Group),
		producer.WithNameServer([]string{fmt.Sprintf("%s:%d", mqConfig.Host, mqConfig.Port)}),
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
	// 使用全局配置中的端口
	for chainName, _ := range define.GlobalConfig.Fabric {
		err := utils.InitializeSDKByChainName(chainName)
		if err != nil {
			log.Fatalf("初始化SDK失败: %v", err)
			return
		}
	}
	port := define.GlobalConfig.Server.Port
	useRouter := router.SetUpRouter()

	log.Printf("服务器正在端口 %d 上运行...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), useRouter))
}
