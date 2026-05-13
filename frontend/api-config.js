const API_CONFIG = {
    HOST: 'localhost',
    PORT: '8080',
    PATH_PREFIX: '/api'
};

const API_BASE_URL = `http://${API_CONFIG.HOST}:${API_CONFIG.PORT}${API_CONFIG.PATH_PREFIX}`;

const API_ENDPOINTS = {
    CONFIG: `${API_BASE_URL}/config`,
    SAVE_CONFIG: `${API_BASE_URL}/config`,
    FEAR_GREED: `${API_BASE_URL}/fear-greed`,
    MARKET: `${API_BASE_URL}/market`
};

const API = {
    async getConfig() {
        const response = await fetch(API_ENDPOINTS.CONFIG);
        const result = await response.json();
        if (result.status !== 'success') {
            throw new Error(result.error || '获取配置失败');
        }
        return result.data;
    },

    async saveConfig(configData) {
        const response = await fetch(API_ENDPOINTS.SAVE_CONFIG, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(configData)
        });
        const result = await response.json();
        if (result.status !== 'success') {
            throw new Error(result.error || '保存配置失败');
        }
        return result;
    },

    async getFearGreed() {
        const response = await fetch(API_ENDPOINTS.FEAR_GREED);
        const result = await response.json();
        if (result.status !== 'success') {
            throw new Error(result.error || '获取恐惧贪婪指数失败');
        }
        return result.data;
    },

    async updateFearGreed(fearGreed) {
        const response = await fetch(API_ENDPOINTS.FEAR_GREED, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ fearGreed })
        });
        const result = await response.json();
        if (result.status !== 'success') {
            throw new Error(result.error || '更新失败');
        }
        return result.data;
    },

    async getMarketData() {
        const response = await fetch(API_ENDPOINTS.MARKET);
        const result = await response.json();
        if (result.status !== 'success') {
            throw new Error(result.error || '获取市场数据失败');
        }
        return result.data;
    }
};
