package gotg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"getGate/config"
	"net/http"
	"os"
	"strings"
)

const (
	botToken = ""
	chatID   = ""
)

type TelegramMessage struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func sendMessage(token, chatID, message string) error {
	// 确保token不包含"bot"前缀，因为URL中已经添加了
	token = strings.TrimPrefix(token, "bot")
	
	// 打印完整URL（仅用于调试，生产环境应移除）
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	fmt.Printf("请求URL: %s\n", url)

	msg := TelegramMessage{
		ChatID: chatID,
		Text:   message,
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// 读取响应体以获取更多错误信息
		var respBody bytes.Buffer
		_, err := respBody.ReadFrom(resp.Body)
		if err != nil {
			return fmt.Errorf("unexpected status code: %d, failed to read response body: %v", resp.StatusCode, err)
		}
		return fmt.Errorf("unexpected status code: %d, response: %s", resp.StatusCode, respBody.String())
	}

	return nil
}

// 验证token是否有效
func validateToken(token string) error {
	// 确保token不包含"bot"前缀
	token = strings.TrimPrefix(token, "bot")
	
	// 构建getMe API URL
	url := fmt.Sprintf("https://api.telegram.org/bot%s/getMe", token)
	fmt.Printf("验证Token URL: %s\n", url)
	
	// 发送请求
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("验证token失败: %w", err)
	}
	defer resp.Body.Close()
	
	// 检查响应状态
	if resp.StatusCode != http.StatusOK {
		// 读取响应体以获取更多错误信息
		var respBody bytes.Buffer
		_, err := respBody.ReadFrom(resp.Body)
		if err != nil {
			return fmt.Errorf("验证token失败，状态码: %d, 无法读取响应: %v", resp.StatusCode, err)
		}
		return fmt.Errorf("验证token失败，状态码: %d, 响应: %s", resp.StatusCode, respBody.String())
	}
	
	fmt.Println("Token验证成功")
	return nil
}

func MyGOTG(msg string) {
	message := msg

	// 打印配置信息（隐藏部分token以保护安全）
	tokenLen := len(config.Myconfig.Telegram.BotToken)
	maskedToken := "****"
	if tokenLen > 4 {
		maskedToken = config.Myconfig.Telegram.BotToken[0:4] + "****" + config.Myconfig.Telegram.BotToken[tokenLen-4:]
	}
	fmt.Printf("准备发送Telegram消息，Token: %s, ChatID: %s\n", maskedToken, config.Myconfig.Telegram.ChatID)

	// 先验证token
	err := validateToken(config.Myconfig.Telegram.BotToken)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Token验证错误: %v\n", err)
		os.Exit(1)
	}

	// 发送消息
	err = sendMessage(config.Myconfig.Telegram.BotToken, config.Myconfig.Telegram.ChatID, message)
	if err != nil {
		fmt.Fprintf(os.Stderr, "发送消息错误: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("发送消息完成")
}
