package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"getGate/config"
	"getGate/gotg"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/antihax/optional"
	gateapi "github.com/gateio/gateapi-go/v6"
)

// QueryGatePairs 查询Gate交易对信息
func QueryGatePairs(pairs []config.PairConfig) []config.PairInfo {
	// 创建Gate API客户端实例
	client := gateapi.NewAPIClient(gateapi.NewConfiguration())
	// uncomment the next line if your are testing against testnet
	// client.ChangeBasePath("https://fx-api-testnet.gateio.ws/api/v4")

	// 创建上下文
	ctx := context.Background()

	// 存储所有交易对信息
	pairInfos := []config.PairInfo{}

	// 遍历配置中的交易对
	for _, pairConfig := range pairs {
		// 添加自定义的ListTickersOpts结构参数
		opts := &gateapi.ListTickersOpts{
			CurrencyPair: optional.NewString(pairConfig.Symbol),
		}

		// 调用API获取行情数据
		result, _, err := client.SpotApi.ListTickers(ctx, opts)
		if err != nil {
			// 处理API错误
			var e gateapi.GateAPIError
			if errors.As(err, &e) {
				fmt.Printf("获取 %s 数据失败: %s\n", pairConfig.Symbol, e.Error())
			}
			continue
		}

		if len(result) > 0 {
			// 将价格字符串转换为浮点数
			priceFloat, err := strconv.ParseFloat(result[0].Last, 64)
			if err != nil {
				fmt.Printf("解析 %s 价格失败: %s\n", pairConfig.Symbol, err.Error())
				continue
			}

			// 判断价格是否在预期范围内
			isOutOfRange := priceFloat < pairConfig.MinPrice || priceFloat > pairConfig.MaxPrice

			// 保存交易对信息
			pairInfos = append(pairInfos, config.PairInfo{
				Pair:             pairConfig.Symbol,
				Price:            result[0].Last,
				PriceFloat:       priceFloat,
				ChangePercentage: result[0].ChangePercentage,
				MaxPrice:         pairConfig.MaxPrice,
				MinPrice:         pairConfig.MinPrice,
				IsOutOfRange:     isOutOfRange,
			})
		}
	}

	return pairInfos
}

func main() {
	// 解析命令行参数
	webMode := flag.Bool("web", false, "启动Web服务器模式")
	flag.Parse()

	// 如果是Web模式，启动Web服务器
	if *webMode {
		// 直接在main.go中实现Web服务器功能
		startWebServer()
		return
	}

	// 读取配置文件
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("加载配置文件失败: %s\n", err.Error())
		os.Exit(1)
	}

	// 查询交易对信息
	pairInfos := QueryGatePairs(cfg.Pairs)

	// 一次性打印所有交易对信息
	result, err := config.PrintPairInfos(pairInfos)
	// 如果没有超出范围的交易对，打印提示信息
	if err != nil {
		// fmt.Printf("打印交易对信息失败: %s\n", err.Error())
		return
	}
	// 如果超出了预期则打印
	// fmt.Print(result)
	//并且发送tg消息个给主人
	gotg.MyGOTG(string(result))
}

// startWebServer 启动Web服务器（仅API服务）
func startWebServer() {
	// 读取配置文件获取端口设置
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("加载配置文件失败: %s\n", err.Error())
		os.Exit(1)
	}
	
	// 获取端口配置，如果配置文件中没有设置，则使用默认端口8080
	port := cfg.Server.Port
	if port == 0 {
		port = 8080 // 默认端口
	}
	
	// 将端口转换为字符串
	portStr := fmt.Sprintf(":%d", port)
	
	// 设置CORS中间件
	corsMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 设置CORS头
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			
			// 如果是预检请求，直接返回
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			
			// 调用下一个处理函数
			next(w, r)
		}
	}
	
	// 处理配置文件读取请求
	http.HandleFunc("/api/config", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 读取配置文件
		configPath := "config.yaml"
		content, err := ioutil.ReadFile(configPath)
		if err != nil {
			http.Error(w, fmt.Sprintf("读取配置文件失败: %v", err), http.StatusInternalServerError)
			return
		}

		// 设置正确的Content-Type
		w.Header().Set("Content-Type", "text/yaml; charset=utf-8")
		w.Write(content)
	}))

	// 处理配置文件保存请求
	http.HandleFunc("/api/save-config", corsMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// 解析请求体
		if err := r.ParseForm(); err != nil {
			http.Error(w, fmt.Sprintf("解析请求失败: %v", err), http.StatusBadRequest)
			return
		}

		// 获取配置内容
		config := r.FormValue("config")
		if config == "" {
			http.Error(w, "配置内容为空", http.StatusBadRequest)
			return
		}

		// 备份原配置文件
		configPath := "config.yaml"
		backupPath := "config.yaml.bak"
		if _, err := os.Stat(configPath); err == nil {
			if err := copyFile(configPath, backupPath); err != nil {
				http.Error(w, fmt.Sprintf("备份配置文件失败: %v", err), http.StatusInternalServerError)
				return
			}
		}

		// 写入新配置
		if err := ioutil.WriteFile(configPath, []byte(config), 0644); err != nil {
			http.Error(w, fmt.Sprintf("写入配置文件失败: %v", err), http.StatusInternalServerError)
			return
		}

		// 返回成功响应
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status": "success", "message": "配置保存成功"}`)
	}))

	// 获取当前目录的绝对路径
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("API服务器启动在端口 %d\n", port)
	fmt.Printf("当前目录: %s\n", dir)
	log.Fatal(http.ListenAndServe(portStr, nil))
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}

	return ioutil.WriteFile(dst, input, 0644)
}
