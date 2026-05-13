(function() {
    const elements = {
        pairsContainer: document.getElementById('pairs-container'),
        addPairBtn: document.getElementById('add-pair-btn'),
        saveBtn: document.getElementById('save-btn'),
        reloadBtn: document.getElementById('reload-btn'),
        serverHostInput: document.getElementById('server-host'),
        serverPortInput: document.getElementById('server-port'),
        apiPrefixInput: document.getElementById('api-prefix'),
        botTokenInput: document.getElementById('botToken'),
        chatIDInput: document.getElementById('chatID'),
        fearGreedInput: document.getElementById('fear-greed'),
        fearGreedValue: document.getElementById('fear-greed-value'),
        fearGreedLabel: document.getElementById('fear-greed-label'),
        cmcApiKeyInput: document.getElementById('cmc-api-key'),
        notification: document.getElementById('notification')
    };

    let currentConfig = null;

    function init() {
        elements.serverHostInput.value = API_CONFIG.HOST;
        elements.serverPortInput.value = API_CONFIG.PORT;
        elements.apiPrefixInput.value = API_CONFIG.PATH_PREFIX;

        elements.addPairBtn.addEventListener('click', addNewPair);
        elements.saveBtn.addEventListener('click', saveConfig);
        elements.reloadBtn.addEventListener('click', loadConfig);
        elements.fearGreedInput.addEventListener('input', updateFearGreedDisplay);

        loadConfig();
    }

    async function loadConfig() {
        try {
            const data = await API.getConfig();
            currentConfig = data;
            displayConfig(data);
            showNotification('配置已加载', 'success');
        } catch (error) {
            console.error('加载配置失败:', error);
            showNotification('加载配置失败: ' + error.message, 'error');
        }
    }

    function displayConfig(config) {
        if (config.server) {
            elements.serverPortInput.value = config.server.port || '';
        }

        elements.pairsContainer.innerHTML = '';
        if (config.pairs && Array.isArray(config.pairs)) {
            config.pairs.forEach(pair => {
                addPairToDOM(pair);
            });
        }

        if (config.telegram) {
            elements.botTokenInput.value = config.telegram.botToken || '';
            elements.chatIDInput.value = config.telegram.chatID || '';
        }

        if (config.coinMarketCap) {
            elements.cmcApiKeyInput.value = config.coinMarketCap.apiKey || '';
        }

        if (config.fearGreed !== undefined) {
            elements.fearGreedInput.value = config.fearGreed;
            updateFearGreedDisplay();
        }
    }

    function updateFearGreedDisplay() {
        const value = parseInt(elements.fearGreedInput.value) || 50;
        elements.fearGreedValue.textContent = value;

        let label = '中性';
        let colorClass = 'neutral';

        if (value <= 20) {
            label = '极度恐惧';
            colorClass = 'extreme-fear';
        } else if (value <= 40) {
            label = '恐惧';
            colorClass = 'fear';
        } else if (value <= 60) {
            label = '中性';
            colorClass = 'neutral';
        } else if (value <= 80) {
            label = '贪婪';
            colorClass = 'greed';
        } else {
            label = '极度贪婪';
            colorClass = 'extreme-greed';
        }

        elements.fearGreedLabel.textContent = label;
        elements.fearGreedValue.className = 'fear-greed-value ' + colorClass;
    }

    function addPairToDOM(pair = { symbol: '', max_price: 0, min_price: 0 }) {
        const pairItem = document.createElement('div');
        pairItem.className = 'pair-item';

        const maxPriceVal = pair.max_price !== undefined && pair.max_price !== null ? pair.max_price : '';
        const minPriceVal = pair.min_price !== undefined && pair.min_price !== null ? pair.min_price : '';

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
                <input type="number" step="any" class="pair-max-price" value="${maxPriceVal}">
            </div>
            <div class="form-group">
                <label>最低价格:</label>
                <input type="number" step="any" class="pair-min-price" value="${minPriceVal}">
            </div>
        `;

        pairItem.querySelector('.remove-pair').addEventListener('click', function() {
            pairItem.remove();
        });

        elements.pairsContainer.appendChild(pairItem);
    }

    function addNewPair() {
        addPairToDOM({ symbol: '', max_price: 0, min_price: 0 });
    }

    async function saveConfig() {
        try {
            const pairs = [];
            const pairElements = elements.pairsContainer.querySelectorAll('.pair-item');

            pairElements.forEach(pairElement => {
                const symbol = pairElement.querySelector('.pair-symbol').value.trim();
                const maxPriceInput = pairElement.querySelector('.pair-max-price').value.trim();
                const minPriceInput = pairElement.querySelector('.pair-min-price').value.trim();

                if (symbol) {
                    const pairData = { symbol };
                    if (maxPriceInput) {
                        pairData.max_price = parseFloat(maxPriceInput);
                    }
                    if (minPriceInput) {
                        pairData.min_price = parseFloat(minPriceInput);
                    }
                    pairs.push(pairData);
                }
            });

            const configData = {
                server: {
                    port: parseInt(elements.serverPortInput.value) || 8080
                },
                pairs: pairs,
                telegram: {
                    botToken: elements.botTokenInput.value.trim(),
                    chatID: elements.chatIDInput.value.trim()
                },
                fearGreed: parseInt(elements.fearGreedInput.value) || 50,
                coinMarketCap: {
                    apiKey: elements.cmcApiKeyInput.value.trim()
                }
            };

            await API.saveConfig(configData);
            showNotification('配置保存成功', 'success');
        } catch (error) {
            console.error('保存配置失败:', error);
            showNotification('保存配置失败: ' + error.message, 'error');
        }
    }

    function showNotification(message, type = 'info') {
        elements.notification.textContent = message;
        elements.notification.className = 'notification ' + type;

        setTimeout(() => {
            elements.notification.classList.add('hidden');
        }, 3000);
    }

    if (document.readyState === 'loading') {
        document.addEventListener('DOMContentLoaded', init);
    } else {
        init();
    }
})();
