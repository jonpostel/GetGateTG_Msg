package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// StartWebServer 启动Web服务器
func StartWebServer() {
	// 设置静态文件服务
	fs := http.FileServer(http.Dir("."))
	http.Handle("/", fs)

	// 处理配置文件读取请求
	http.HandleFunc("/config.yaml", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// 处理配置文件保存请求
	http.HandleFunc("/save-config", func(w http.ResponseWriter, r *http.Request) {
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
	})

	// 获取当前目录的绝对路径
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("服务器启动在 http://localhost:8080\n")
	fmt.Printf("当前目录: %s\n", dir)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
