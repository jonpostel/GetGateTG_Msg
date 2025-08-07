package main

import (
	"context"
	"errors"
	"fmt"
	"getGate/config"
	"getGate/gotg"
	"os"
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
