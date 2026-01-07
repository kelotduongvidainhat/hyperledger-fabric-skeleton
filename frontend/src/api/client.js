import axios from 'axios';

export const api = axios.create({
    baseURL: 'http://localhost:3000',
    headers: {
        'Content-Type': 'application/json',
    },
});

api.interceptors.request.use((config) => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

export const fetchAssets = async () => {
    const response = await api.get('/assets');
    return response.data;
};

export const fetchAssetById = async (id) => {
    const response = await api.get(`/assets/${id}`);
    return response.data;
};

export const createAsset = async (asset) => {
    const response = await api.post('/assets', asset);
    return response.data;
};

export const proposeTransfer = async (id, targetUser) => {
    const response = await api.post(`/assets/${id}/transfer`, { target_user: targetUser });
    return response.data;
};

export const acceptTransfer = async (id) => {
    const response = await api.post(`/assets/${id}/accept`);
    return response.data;
};

export const updateAssetView = async (id, view) => {
    const response = await api.post(`/assets/${id}/view`, { view });
    return response.data;
};

export const deleteAsset = async (id) => {
    const response = await api.delete(`/assets/${id}`);
    return response.data;
};

export const fetchHistory = async (id) => {
    const response = await api.get(`/assets/${id}/history`);
    return response.data;
};

export const fetchNotifications = async () => {
    const response = await api.get('/notifications');
    return response.data;
};

export const markNotificationRead = async (id) => {
    const response = await api.post(`/notifications/${id}/read`);
    return response.data;
};

export default api;
