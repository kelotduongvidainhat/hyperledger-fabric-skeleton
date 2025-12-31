import axios from 'axios';

const API_URL = 'http://localhost:8080';

const api = axios.create({
    baseURL: API_URL,
});

export const setAuthToken = (userId) => {
    if (userId) {
        api.defaults.headers.common['X-User-ID'] = userId;
    } else {
        delete api.defaults.headers.common['X-User-ID'];
    }
};

export const fetchAssets = async () => {
    const response = await api.get('/assets');
    return response.data;
};

export const fetchAsset = async (id) => {
    const response = await api.get(`/assets/${id}`);
    return response.data;
};

export const createAsset = async (assetData) => {
    const response = await api.post('/assets', assetData);
    return response.data;
};

export const transferAsset = async (id, newOwner) => {
    const response = await api.put(`/assets/${id}/transfer`, { newOwner });
    return response.data;
};

export const lockAsset = async (id) => {
    const response = await api.put(`/assets/${id}/lock`);
    return response.data;
};

export const unlockAsset = async (id) => {
    const response = await api.put(`/assets/${id}/unlock`);
    return response.data;
};

export const fetchIdentities = async () => {
    const response = await api.get('/admin/identities');
    return response.data;
};

export const fetchAssetHistory = async (id) => {
    const response = await api.get(`/assets/${id}/history`);
    console.log("fetchAssetHistory response:", response.data);
    return response.data;
};

export const fetchAssetsFromDB = async () => {
    const response = await api.get('/api/query/assets'); // Modified specific to project structure
    // Normalize data structure if needed to match blockchain response
    return response.data;
};
