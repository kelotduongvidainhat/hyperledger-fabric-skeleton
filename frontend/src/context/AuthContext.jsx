import React, { createContext, useState, useContext, useEffect } from 'react';
import { api } from '../api/client';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [token, setToken] = useState(localStorage.getItem('token'));
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        // Check if token exists on mount
        if (token) {
            // Ideally we validate the token with backend, but for MVP we assume validity if present
            // decoding logic could be added here
            const savedUser = localStorage.getItem('username');
            if (savedUser) {
                setUser({ username: savedUser });
                setIsAuthenticated(true);
            }
        }
        setLoading(false);
    }, [token]);

    const login = async (username, password) => {
        try {
            const response = await api.post('/auth/login', { username, password });
            const { token: newToken, username: newUsername } = response.data;

            localStorage.setItem('token', newToken);
            localStorage.setItem('username', newUsername);

            setToken(newToken);
            setUser({ username: newUsername });
            setIsAuthenticated(true);
            return true;
        } catch (error) {
            console.error("Login failed", error);
            return false;
        }
    };

    const logout = () => {
        localStorage.removeItem('token');
        localStorage.removeItem('username');
        setToken(null);
        setUser(null);
        setIsAuthenticated(false);
    };

    return (
        <AuthContext.Provider value={{ user, token, isAuthenticated, login, logout, loading }}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => useContext(AuthContext);
