document.addEventListener('DOMContentLoaded', function() {
    // 初始化页面元素
    const pairsContainer = document.getElementById('pairs-container');
    const addPairBtn = document.getElementById('add-pair-btn');
    const saveBtn = document.getElementById('save-btn');
    const reloadBtn = document.getElementById('reload-btn');
    const serverHostInput = document.getElementById('server-host');
    const serverPortInput = document.getElementById('server-port');
    const apiPrefixInput = document.getElementById('api-prefix');
    const botTokenInput = document.getElementById('botToken');
    const chatIDInput = document.getElementById('chatID');
    const notification = document.getElementById('notification');
    
    // 加载配置
    loadConfig();
    
    // 添加事件监听器
    addPairBtn.addEventListener('click', addNewPair);
    saveBtn.addEventListener('click', saveConfig);
    reloadBtn.addEventListener('click', loadConfig);
    
    // 显示服务器配置（来自api-config.js）
    serverHostInput.value = API_CONFIG.HOST;
    serverPortInput.value = API_CONFIG.PORT;
    apiPrefixInput.value = API_CONFIG.PATH_PREFIX;
    
    // 加载配置函数
    function loadConfig() {
        fetch(API_ENDPOINTS.CONFIG)
            .then(response => {
                if (!response.ok) {
                    throw new Error('无法加载配置文件');
                }
                return response.text();
            })
            .then(data => {
                parseAndDisplayConfig(data);
                showNotification('配置已加载', 'success');
            })
            .catch(error => {
                console.error('加载配置失败:', error);
                showNotification('加载配置失败: ' + error.message, 'error');
            });
    }
    
    // 解析并显示配置
    function parseAndDisplayConfig(configText) {
        const pairsConfig = [];
        let telegramConfig = { botToken: '', chatID: '' };
        
        // 解析配置文件
        const lines = configText.split('\n');
        let currentSection = '';
        
        for (const line of lines) {
            const trimmedLine = line.trim();
            
            // 跳过空行和注释
            if (!trimmedLine || trimmedLine.startsWith('#')) {
                continue;
            }
            
            // 检查是否是新的配置段
            if (trimmedLine.endsWith(':')) {
                currentSection = trimmedLine.substring(0, trimmedLine.length - 1);
                continue;
            }
            
            // 根据当前配置段解析配置项
            if (currentSection === 'pairs') {
                // 解析交易对配置
                if (trimmedLine.startsWith('- symbol:')) {
                    const symbol = trimmedLine.substring('- symbol:'.length).trim();
                    pairsConfig.push({ symbol, max_price: '', min_price: '' });
                } else if (trimmedLine.includes('max_price:') && pairsConfig.length > 0) {
                    const maxPrice = trimmedLine.substring(trimmedLine.indexOf('max_price:') + 'max_price:'.length).trim();
                    pairsConfig[pairsConfig.length - 1].max_price = maxPrice;
                } else if (trimmedLine.includes('min_price:') && pairsConfig.length > 0) {
                    const minPrice = trimmedLine.substring(trimmedLine.indexOf('min_price:') + 'min_price:'.length).trim();
                    pairsConfig[pairsConfig.length - 1].min_price = minPrice;
                }
            } else if (currentSection === 'Telegram') {
                // 解析Telegram配置
                if (trimmedLine.startsWith('botToken:')) {
                    telegramConfig.botToken = trimmedLine.substring('botToken:'.length).trim();
                } else if (trimmedLine.startsWith('chatID:')) {
                    telegramConfig.chatID = trimmedLine.substring('chatID:'.length).trim();
                }
            }
        }
        
        // 显示服务器配置（从api-config.js获取）
        serverHostInput.value = API_CONFIG.HOST;
        serverPortInput.value = API_CONFIG.PORT;
        apiPrefixInput.value = API_CONFIG.PATH_PREFIX;
        
        // 显示交易对配置
        pairsContainer.innerHTML = '';
        pairsConfig.forEach(pair => {
            addPairToDOM(pair);
        });
        
        // 显示Telegram配置
        botTokenInput.value = telegramConfig.botToken;
        chatIDInput.value = telegramConfig.chatID;
    }
    
    // 添加交易对到DOM
    function addPairToDOM(pair) {
        const pairItem = document.createElement('div');
        pairItem.className = 'pair-item';
        
        pairItem.innerHTML = `
            <div class="pair-header">
                <div class="pair-title">交易对</div>
                <button class="remove-pair">×</button>
            </div>
            <div class="form-group">
                <label>交易对符号:</label>
                <input type="text" class="pair-symbol" value="${pair.symbol || ''}">
            </div>
            <div class="form-group">
                <label>最高价格:</label>
                <input type="text" class="pair-max-price" value="${pair.max_price || ''}">
            </div>
            <div class="form-group">
                <label>最低价格:</label>
                <input type="text" class="pair-min-price" value="${pair.min_price || ''}">
            </div>
        `;
        
        // 添加删除按钮事件
        pairItem.querySelector('.remove-pair').addEventListener('click', function() {
            pairItem.remove();
        });
        
        pairsContainer.appendChild(pairItem);
    }
    
    // 添加新的交易对
    function addNewPair() {
        addPairToDOM({
            symbol: '',
            max_price: '',
            min_price: ''
        });
    }
    
    // 保存配置
    function saveConfig() {
        // 收集交易对配置
        const pairs = [];
        const pairElements = pairsContainer.querySelectorAll('.pair-item');
        
        pairElements.forEach(pairElement => {
            const symbol = pairElement.querySelector('.pair-symbol').value.trim();
            const maxPrice = pairElement.querySelector('.pair-max-price').value.trim();
            const minPrice = pairElement.querySelector('.pair-min-price').value.trim();
            
            if (symbol) {
                pairs.push({
                    symbol,
                    max_price: maxPrice,
                    min_price: minPrice
                });
            }
        });
        
        // 收集Telegram配置
        const botToken = botTokenInput.value.trim();
        const chatID = chatIDInput.value.trim();
        
        // 构建YAML内容
        let yamlContent = '# Gate.io API 配置文件\n\n';
        
        // 添加服务器配置
        yamlContent += '# 服务器配置\n';
        yamlContent += 'Server:\n';
        yamlContent += `  port: ${serverPortInput.value}\n\n`;
        
        // 添加交易对配置
        yamlContent += '# 交易对列表，包含最大预期价格和最低预期价格\n';
        yamlContent += 'pairs:\n';
        pairs.forEach(pair => {
            yamlContent += `  - symbol: ${pair.symbol}\n`;
            yamlContent += `    max_price: ${pair.max_price}\n`;
            yamlContent += `    min_price: ${pair.min_price}\n`;
        });
        yamlContent += '\n';
        
        // 添加Telegram配置
        yamlContent += 'Telegram:\n';
        yamlContent += `  botToken: ${botToken}\n`;
        yamlContent += `  chatID: ${chatID}\n`;
        
        // 发送到后端保存
        fetch(API_ENDPOINTS.SAVE_CONFIG, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded',
            },
            body: 'config=' + encodeURIComponent(yamlContent)
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('保存配置失败');
            }
            return response.json();
        })
        .then(data => {
            showNotification('配置保存成功', 'success');
        })
        .catch(error => {
            showNotification('保存配置失败: ' + error.message, 'error');
            console.error('保存配置失败:', error);
        });
    }
    
    // 显示通知
    function showNotification(message, type) {
        notification.textContent = message;
        notification.className = 'notification ' + type;
        
        // 3秒后隐藏通知
        setTimeout(() => {
            notification.className = 'notification hidden';
        }, 3000);
    }
});