import React, { createContext, useContext, useState, useEffect } from 'react';
import axios from 'axios';

const AuthContext = createContext(null);

export const AuthProvider = ({ children }) => {
    const [user, setUser] = useState(null);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        // Check if user is logged in from local storage
        const storedUser = localStorage.getItem('fabric_user');
        if (storedUser) {
            const parsedUser = JSON.parse(storedUser);
            setUser(parsedUser);
            axios.defaults.headers.common['X-User-ID'] = parsedUser.username;
        }
        setLoading(false);
    }, []);

    const login = (userData) => {
        setUser(userData);
        localStorage.setItem('fabric_user', JSON.stringify(userData));
        axios.defaults.headers.common['X-User-ID'] = userData.username;
    };

    const logout = () => {
        setUser(null);
        localStorage.removeItem('fabric_user');
        delete axios.defaults.headers.common['X-User-ID'];
    };

    return (
        <AuthContext.Provider value={{ user, login, logout, loading }}>
            {!loading && children}
        </AuthContext.Provider>
    );
};

export const useAuth = () => useContext(AuthContext);
