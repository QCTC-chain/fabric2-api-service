package main

import (
	"fmt"
	"github.com/qctc/fabric2-api-server/define"
	"github.com/qctc/fabric2-api-server/router"
	"gopkg.in/yaml.v2"
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
}

func main() {
	// 使用全局配置中的端口
	port := define.GlobalConfig.Server.Port
	useRouter := router.SetUpRouter()

	log.Printf("服务器正在端口 %d 上运行...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), useRouter))
}
