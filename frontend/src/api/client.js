import axios from 'axios';

export const api = axios.create({
    baseURL: import.meta.env.VITE_API_URL || 'http://localhost:3000',
    headers: {
        'Content-Type': 'application/json',
    },
    withCredentials: true, // Required for cookies
});

api.interceptors.request.use((config) => {
    const token = localStorage.getItem('token');
    if (token) {
        config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
});

// Response interceptor for silent token refresh
api.interceptors.response.use(
    (response) => response,
    async (error) => {
        const originalRequest = error.config;

        // If 401 and we haven't retried yet
        if (error.response?.status === 401 && !originalRequest._retry) {
            originalRequest._retry = true;

            try {
                // Try to refresh the token using the refresh_token cookie
                const { data } = await axios.post(
                    `${api.defaults.baseURL}/auth/refresh`,
                    {},
                    { withCredentials: true }
                );

                // Update localStorage if we still use it (backup)
                if (data.token) {
                    localStorage.setItem('token', data.token);
                }

                // Retry original request
                return api(originalRequest);
            } catch (refreshError) {
                // Clear state if refresh fails
                localStorage.removeItem('token');
                localStorage.removeItem('user');
                // Redirecting to login is handled by the component using the error
                return Promise.reject(refreshError);
            }
        }
        return Promise.reject(error);
    }
);

export const fetchAssets = async () => {
    const response = await api.get('/assets');
    return response.data;
};

export const fetchAssetById = async (id) => {
    const response = await api.get(`/assets/${id}`);
    return response.data;
};

export const fetchBlockchainAsset = async (id) => {
    const response = await api.get(`/assets/${id}/blockchain`);
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

export const deleteAccount = async () => {
    const response = await api.delete('/auth/me');
    return response.data;
};

export const uploadToIPFS = async (file) => {
    const formData = new FormData();
    formData.append('file', file);
    const response = await api.post('/api/ipfs/upload', formData, {
        headers: {
            'Content-Type': 'multipart/form-data',
        },
    });
    return response.data;
};

export default api;
