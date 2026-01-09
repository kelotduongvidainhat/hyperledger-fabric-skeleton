import React, { createContext, useState, useContext, useEffect } from 'react';
import { api } from '../api/client';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [token, setToken] = useState(localStorage.getItem('token'));
    const [role, setRole] = useState(localStorage.getItem('role'));
    const [isAuthenticated, setIsAuthenticated] = useState(false);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const checkSession = async () => {
            const savedUser = localStorage.getItem('username');
            if (savedUser) {
                try {
                    // Try to refresh or verify session implicitly via cookies
                    // Our interceptor in client.js will handle the actual refresh if 401
                    // For now, we trust localStorage for UI state if cookies exist
                    const savedRole = localStorage.getItem('role');
                    const savedOrg = localStorage.getItem('org');
                    setUser({ username: savedUser, org: savedOrg });
                    setRole(savedRole);
                    setIsAuthenticated(true);
                } catch (err) {
                    logout();
                }
            }
            setLoading(false);
        };
        checkSession();
    }, []);

    const login = async (username, password, org = "") => {
        try {
            const response = await api.post('/auth/login', { username, password, org });
            const { username: newUsername, role: newRole, org: newOrg } = response.data;

            // Store non-sensitive profile info for UI persist
            localStorage.setItem('username', newUsername);
            localStorage.setItem('role', newRole);
            localStorage.setItem('org', newOrg);

            setRole(newRole);
            setUser({ username: newUsername, org: newOrg });
            setIsAuthenticated(true);
            return true;
        } catch (error) {
            console.error("Login failed", error);
            return false;
        }
    };

    const logout = async () => {
        try {
            await api.post('/auth/logout');
        } catch (err) {
            console.error("Logout request failed", err);
        }
        localStorage.clear();
        setToken(null);
        setRole(null);
        setUser(null);
        setIsAuthenticated(false);
    };

    return (
        <AuthContext.Provider value={{ user, token, role, isAuthenticated, login, logout, loading }}>
            {children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => useContext(AuthContext);
