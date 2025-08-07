package config

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Pairs     []PairConfig `yaml:"pairs"`
	Cellframe Cellframe    `yaml:"Cellframe"`
	Telegram  Telegram     `yaml:"Telegram"`
}

// PairConfig 交易对配置结构体
type PairConfig struct {
	Symbol   string  `yaml:"symbol"`
	MaxPrice float64 `yaml:"max_price"`
	MinPrice float64 `yaml:"min_price"`
}

type Cellframe struct {
	WalletName string `yaml:"walletName"`
}

type Telegram struct {
	BotToken string `yaml:"botToken"`
	ChatID   string `yaml:"chatID"`
}

// PairInfo 交易对信息结构体
type PairInfo struct {
	Pair             string
	Price            string
	PriceFloat       float64
	ChangePercentage string
	MaxPrice         float64
	MinPrice         float64
	IsOutOfRange     bool
}

// 配置对象
var Myconfig = &Config{
	Pairs: []PairConfig{},
	Cellframe: Cellframe{
		WalletName: "",
	},
	Telegram: Telegram{
		BotToken: "",
		ChatID:   "",
	},
}

// LoadConfig 加载配置文件
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	// 更新全局配置对象
	Myconfig = &config

	return &config, nil
}

// LodConfig 兼容旧版本的配置加载方法
func LodConfig() *Config {
	// 打开 YAML 文件
	config, err := LoadConfig("./config.yaml")
	if err != nil {
		fmt.Println("Error loading config:", err)
		return nil
	}

	fmt.Println("读取配置文件：")
	fmt.Printf("%+v \n", config)
	return config
}

// PrintPairInfos 打印交易对信息
func PrintPairInfos(infos []PairInfo) (string, error) {
	// 筛选出价格超出预期范围的交易对
	outOfRangeInfos := []PairInfo{}
	for _, info := range infos {
		if info.IsOutOfRange {
			outOfRangeInfos = append(outOfRangeInfos, info)
		}
	}

	if len(outOfRangeInfos) == 0 {
		return "没有交易对的价格超出预期范围", errors.New("没有交易对的价格超出预期范围")
	}

	// 构建输出字符串
	var sb strings.Builder
	sb.WriteString("价格超出预期范围的交易对信息汇总:\n")
	sb.WriteString("----------------------------------------\n")

	// 添加每个交易对的信息
	for _, info := range outOfRangeInfos {
		sb.WriteString(fmt.Sprintf("交易对: %s\n", info.Pair))
		sb.WriteString(fmt.Sprintf("  当前价格: %s\n", info.Price))
		sb.WriteString(fmt.Sprintf("  预期价格范围: %.4f - %.4f\n", info.MinPrice, info.MaxPrice))

		// 判断价格是高于最大预期还是低于最小预期
		if info.PriceFloat > info.MaxPrice {
			sb.WriteString(fmt.Sprintf("  价格状态: 高于最大预期 %.4f\n", info.MaxPrice))
		} else if info.PriceFloat < info.MinPrice {
			sb.WriteString(fmt.Sprintf("  价格状态: 低于最小预期 %.4f\n", info.MinPrice))
		}

		sb.WriteString(fmt.Sprintf("  24h涨跌百分比: %s%%\n", info.ChangePercentage))
		sb.WriteString("----------------------------------------\n")
	}

	return sb.String(), nil
}
