// API配置文件
// 修改此文件中的配置可以更改后端API的地址和端口

// 后端服务器配置
const API_CONFIG = {
    // 后端服务器地址（不包含端口）
    HOST: 'localhost',
    // 后端服务器端口
    PORT: '8080',
    // API路径前缀
    PATH_PREFIX: '/api'
};

// 构建完整的API基础URL
const API_BASE_URL = `http://${API_CONFIG.HOST}:${API_CONFIG.PORT}${API_CONFIG.PATH_PREFIX}`;

// API端点
const API_ENDPOINTS = {
    CONFIG: `${API_BASE_URL}/config`,
    SAVE_CONFIG: `${API_BASE_URL}/save-config`
};