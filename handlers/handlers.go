package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"getGate/config"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/antihax/optional"
	gateapi "github.com/gateio/gateapi-go/v6"
)

type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Success(w http.ResponseWriter, message string, data interface{}) {
	JSON(w, http.StatusOK, Response{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, Response{
		Status: "error",
		Error:  message,
	})
}

func GetConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("加载配置失败: %v", err))
		return
	}

	pairs := []map[string]interface{}{}
	for _, p := range cfg.Pairs {
		pairs = append(pairs, map[string]interface{}{
			"symbol":    p.Symbol,
			"max_price": p.MaxPrice,
			"min_price": p.MinPrice,
		})
	}

	configData := map[string]interface{}{
		"server": map[string]int{
			"port": cfg.Server.Port,
		},
		"pairs": pairs,
		"telegram": map[string]string{
			"botToken": cfg.Telegram.BotToken,
			"chatID":   cfg.Telegram.ChatID,
		},
		"fearGreed": cfg.FearGreed,
		"coinMarketCap": map[string]string{
			"apiKey": cfg.CoinMarketCap.APIKey,
		},
	}

	Success(w, "配置获取成功", configData)
}

type SaveConfigRequest struct {
	Server        *ServerConfig        `json:"server,omitempty"`
	Pairs         []PairRequest        `json:"pairs,omitempty"`
	Telegram      *TelegramConfig      `json:"telegram,omitempty"`
	FearGreed     *int                 `json:"fearGreed,omitempty"`
	CoinMarketCap *CoinMarketCapConfig `json:"coinMarketCap,omitempty"`
}

type PairRequest struct {
	Symbol   string   `json:"symbol"`
	MaxPrice *float64 `json:"max_price"`
	MinPrice *float64 `json:"min_price"`
}

type ServerConfig struct {
	Port int `json:"port"`
}

type TelegramConfig struct {
	BotToken string `json:"botToken"`
	ChatID   string `json:"chatID"`
}

type CoinMarketCapConfig struct {
	APIKey string `json:"apiKey"`
}

func SaveConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Error(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	var req SaveConfigRequest
	if err := json.Unmarshal(body, &req); err != nil {
		Error(w, http.StatusBadRequest, fmt.Sprintf("解析JSON失败: %v", err))
		return
	}

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("加载配置失败: %v", err))
		return
	}

	if req.Server != nil {
		cfg.Server.Port = req.Server.Port
	}
	if req.Pairs != nil {
		pairs := make([]config.PairConfig, 0, len(req.Pairs))
		for _, p := range req.Pairs {
			pair := config.PairConfig{Symbol: p.Symbol}
			if p.MaxPrice != nil {
				pair.MaxPrice = *p.MaxPrice
			}
			if p.MinPrice != nil {
				pair.MinPrice = *p.MinPrice
			}
			pairs = append(pairs, pair)
		}
		cfg.Pairs = pairs
	}
	if req.Telegram != nil {
		cfg.Telegram.BotToken = req.Telegram.BotToken
		cfg.Telegram.ChatID = req.Telegram.ChatID
	}
	if req.FearGreed != nil {
		cfg.FearGreed = *req.FearGreed
	}
	if req.CoinMarketCap != nil {
		cfg.CoinMarketCap.APIKey = req.CoinMarketCap.APIKey
	}

	backupPath := "config.yaml.bak"
	if _, err := os.Stat("config.yaml"); err == nil {
		if err := copyFile("config.yaml", backupPath); err != nil {
			Error(w, http.StatusInternalServerError, fmt.Sprintf("备份配置文件失败: %v", err))
			return
		}
	}

	data, err := yamlFromConfig(cfg)
	if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("生成配置失败: %v", err))
		return
	}

	if err := ioutil.WriteFile("config.yaml", []byte(data), 0644); err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("写入配置文件失败: %v", err))
		return
	}

	config.Myconfig = cfg
	Success(w, "配置保存成功", nil)
}

func yamlFromConfig(cfg *config.Config) (string, error) {
	var sb strings.Builder

	sb.WriteString("# Gate.io API 配置文件\n\n")
	sb.WriteString("# 服务器配置\n")
	sb.WriteString("Server:\n")
	sb.WriteString(fmt.Sprintf("  port: %d\n\n", cfg.Server.Port))

	sb.WriteString("# 交易对列表，包含最大预期价格和最低预期价格\n")
	sb.WriteString("pairs:\n")
	for _, pair := range cfg.Pairs {
		sb.WriteString(fmt.Sprintf("  - symbol: %s\n", pair.Symbol))
		if pair.MaxPrice != 0 {
			sb.WriteString(fmt.Sprintf("    max_price: %v\n", pair.MaxPrice))
		}
		if pair.MinPrice != 0 {
			sb.WriteString(fmt.Sprintf("    min_price: %v\n", pair.MinPrice))
		}
	}
	sb.WriteString("\n")

	sb.WriteString(fmt.Sprintf("FearGreed: %d\n\n", cfg.FearGreed))

	sb.WriteString("# CoinMarketCap API 配置\n")
	sb.WriteString("CoinMarketCap:\n")
	sb.WriteString(fmt.Sprintf("  apiKey: %s\n\n", cfg.CoinMarketCap.APIKey))

	sb.WriteString("# Telegram 配置\n")
	sb.WriteString("Telegram:\n")
	sb.WriteString(fmt.Sprintf("  botToken: %s\n", cfg.Telegram.BotToken))
	sb.WriteString(fmt.Sprintf("  chatID: %s\n", cfg.Telegram.ChatID))

	return sb.String(), nil
}

func GetFearGreed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("加载配置失败: %v", err))
		return
	}

	data := map[string]interface{}{
		"fearGreed": cfg.FearGreed,
	}

	Success(w, "获取成功", data)
}

func FearGreedHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetFearGreed(w, r)
	case http.MethodPost:
		UpdateFearGreed(w, r)
	default:
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetConfig(w, r)
	case http.MethodPost:
		SaveConfig(w, r)
	default:
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

type UpdateFearGreedRequest struct {
	FearGreed int `json:"fearGreed"`
}

func UpdateFearGreed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		Error(w, http.StatusBadRequest, "读取请求体失败")
		return
	}
	defer r.Body.Close()

	var req UpdateFearGreedRequest
	if err := json.Unmarshal(body, &req); err != nil {
		Error(w, http.StatusBadRequest, fmt.Sprintf("解析JSON失败: %v", err))
		return
	}

	if req.FearGreed < 0 || req.FearGreed > 100 {
		Error(w, http.StatusBadRequest, "fearGreed 值必须在 0-100 之间")
		return
	}

	content, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("读取配置文件失败: %v", err))
		return
	}

	lines := strings.Split(string(content), "\n")
	var newLines []string
	found := false
	for _, line := range lines {
		if strings.HasPrefix(strings.TrimSpace(line), "FearGreed:") {
			newLines = append(newLines, fmt.Sprintf("FearGreed: %d", req.FearGreed))
			found = true
		} else {
			newLines = append(newLines, line)
		}
	}
	if !found {
		newLines = append(newLines, fmt.Sprintf("\nFearGreed: %d", req.FearGreed))
	}

	backupPath := "config.yaml.bak"
	if _, err := os.Stat("config.yaml"); err == nil {
		if err := copyFile("config.yaml", backupPath); err != nil {
			Error(w, http.StatusInternalServerError, fmt.Sprintf("备份配置文件失败: %v", err))
			return
		}
	}

	newContent := strings.Join(newLines, "\n")
	if err := ioutil.WriteFile("config.yaml", []byte(newContent), 0644); err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("写入配置文件失败: %v", err))
		return
	}

	config.Myconfig.FearGreed = req.FearGreed

	data := map[string]interface{}{
		"fearGreed": req.FearGreed,
	}
	Success(w, "更新成功", data)
}

func GetMarketData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		Error(w, http.StatusInternalServerError, fmt.Sprintf("加载配置失败: %v", err))
		return
	}

	pairInfos := QueryGatePairs(cfg.Pairs)

	data := map[string]interface{}{
		"pairs": pairInfos,
	}

	Success(w, "获取成功", data)
}

func QueryGatePairs(pairs []config.PairConfig) []config.PairInfo {
	client := gateapi.NewAPIClient(gateapi.NewConfiguration())
	ctx := context.Background()

	pairInfos := []config.PairInfo{}

	for _, pairConfig := range pairs {
		opts := &gateapi.ListTickersOpts{
			CurrencyPair: optional.NewString(pairConfig.Symbol),
		}

		result, _, err := client.SpotApi.ListTickers(ctx, opts)
		if err != nil {
			var e gateapi.GateAPIError
			if errors.As(err, &e) {
				fmt.Printf("获取 %s 数据失败: %s\n", pairConfig.Symbol, e.Error())
			}
			continue
		}

		if len(result) > 0 {
			priceFloat, err := strconv.ParseFloat(result[0].Last, 64)
			if err != nil {
				fmt.Printf("解析 %s 价格失败: %s\n", pairConfig.Symbol, err.Error())
				continue
			}

			isOutOfRange := priceFloat < pairConfig.MinPrice || priceFloat > pairConfig.MaxPrice

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

func copyFile(src, dst string) error {
	input, err := ioutil.ReadFile(src)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(dst, input, 0644)
}
