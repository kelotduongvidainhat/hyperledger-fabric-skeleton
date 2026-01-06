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
        if (token) {
            const savedUser = localStorage.getItem('username');
            const savedRole = localStorage.getItem('role');
            const savedOrg = localStorage.getItem('org');
            if (savedUser) {
                setUser({ username: savedUser, org: savedOrg });
                setRole(savedRole);
                setIsAuthenticated(true);
            }
        }
        setLoading(false);
    }, [token]);

    const login = async (username, password) => {
        try {
            const response = await api.post('/auth/login', { username, password });
            const { token: newToken, username: newUsername, role: newRole, org: newOrg } = response.data;

            localStorage.setItem('token', newToken);
            localStorage.setItem('username', newUsername);
            localStorage.setItem('role', newRole);
            localStorage.setItem('org', newOrg);

            setToken(newToken);
            setRole(newRole);
            setUser({ username: newUsername, org: newOrg });
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
        localStorage.removeItem('role');
        localStorage.removeItem('org');
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
