package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"getGate/config"
	"getGate/gotg"
	"getGate/handlers"
	"getGate/router"
	"io"
	"log"
	"net/http"
	"os"
)

func GetFearAndGreedIndex(apiKey string) (int, error) {
	url := "https://pro-api.coinmarketcap.com/v3/fear-and-greed/latest"

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("X-CMC_PRO_API_KEY", apiKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, err
	}

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return 0, fmt.Errorf("无法解析返回数据")
	}

	value, ok := data["value"].(float64)
	if !ok {
		return 0, fmt.Errorf("无法获取恐惧与贪婪指数值")
	}

	return int(value), nil
}

func CheckFearGreedIndex(cfg *config.Config) {
	if cfg.CoinMarketCap.APIKey != "" && cfg.FearGreed > 0 {
		currentFG, err := GetFearAndGreedIndex(cfg.CoinMarketCap.APIKey)
		if err != nil {
			fmt.Printf("获取恐惧与贪婪指数失败: %s\n", err.Error())
		} else {
			fmt.Printf("当前恐惧与贪婪指数: %d, 阈值: %d\n", currentFG, cfg.FearGreed)
			if currentFG < cfg.FearGreed {
				fgMsg := fmt.Sprintf("⚠️ 恐惧与贪婪指数预警\n当前指数: %d\n阈值: %d\n状态: 低于设定阈值!", currentFG, cfg.FearGreed)
				gotg.MyGOTG(fgMsg)
			}
		}
	}
}

func main() {
	webMode := flag.Bool("web", false, "启动Web服务器模式")
	flag.Parse()

	if *webMode {
		startWebServer()
		return
	}

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("加载配置文件失败: %s\n", err.Error())
		os.Exit(1)
	}

	pairInfos := handlers.QueryGatePairs(cfg.Pairs)

	result, err := config.PrintPairInfos(pairInfos)
	if err != nil {
	} else {
		gotg.MyGOTG(string(result))
	}

	CheckFearGreedIndex(cfg)
	fmt.Println("程序运行完成")
}

func startWebServer() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("加载配置文件失败: %s\n", err.Error())
		os.Exit(1)
	}

	port := cfg.Server.Port
	if port == 0 {
		port = 8080
	}

	portStr := fmt.Sprintf(":%d", port)

	router.SetupRoutes()

	http.Handle("/", http.FileServer(http.Dir("./frontend")))

	fmt.Printf("API服务器启动在端口 %d\n", port)
	log.Fatal(http.ListenAndServe(portStr, nil))
}
