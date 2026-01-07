import React, { useState, useEffect, useRef } from 'react';
import { Scroll, Feather, Globe, User, Bell, CheckCircle, Info, AlertTriangle } from 'lucide-react';
import { Link, useNavigate } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';
import { fetchNotifications, markNotificationRead } from '../api/client';

const Layout = ({ children }) => {
    const { user, role, logout, isAuthenticated } = useAuth();
    const navigate = useNavigate();
    const [notifications, setNotifications] = useState([]);
    const [showNotifs, setShowNotifs] = useState(false);
    const dropdownRef = useRef(null);

    useEffect(() => {
        if (isAuthenticated) {
            loadNotifications();
            const interval = setInterval(loadNotifications, 30000); // Poll every 30s
            return () => clearInterval(interval);
        }
    }, [isAuthenticated]);

    // Close dropdown on outside click
    useEffect(() => {
        const handleClickOutside = (event) => {
            if (dropdownRef.current && !dropdownRef.current.contains(event.target)) {
                setShowNotifs(false);
            }
        };
        document.addEventListener('mousedown', handleClickOutside);
        return () => document.removeEventListener('mousedown', handleClickOutside);
    }, []);

    const loadNotifications = async () => {
        try {
            const data = await fetchNotifications();
            setNotifications(data || []);
        } catch (err) {
            console.error("Failed to load notifications", err);
        }
    };

    const handleMarkRead = async (e, id, link) => {
        e.preventDefault();
        e.stopPropagation();
        try {
            await markNotificationRead(id);
            setNotifications(notifications.map(n => n.id === id ? { ...n, is_read: true } : n));
            if (link) {
                setShowNotifs(false);
                navigate(link);
            }
        } catch (err) {
            console.error("Failed to mark notification as read", err);
        }
    };

    const unreadCount = notifications.filter(n => !n.is_read).length;

    const getIcon = (type) => {
        switch (type) {
            case 'success': return <CheckCircle className="w-4 h-4 text-green-500" />;
            case 'warning': return <AlertTriangle className="w-4 h-4 text-amber-500" />;
            default: return <Info className="w-4 h-4 text-blue-500" />;
        }
    };

    return (
        <div className="min-h-screen bg-parchment-100 font-sans text-ink-900 selection:bg-bronze selection:text-white">
            {/* Header / Navbar */}
            <header className="sticky top-0 z-50 bg-parchment-200/90 backdrop-blur-sm border-b border-ink-800/20 shadow-sm">
                <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                    <div className="flex justify-between items-center h-20">
                        {/* Logo */}
                        <Link to="/" className="flex items-center gap-3">
                            <div className="bg-ink-800 p-2 rounded-full">
                                <Scroll className="w-6 h-6 text-parchment-100" />
                            </div>
                            <h1 className="text-2xl font-serif font-bold tracking-tight text-ink-800">
                                Ownership Registry
                            </h1>
                        </Link>

                        {/* Nav */}
                        <nav className="flex gap-8 items-center">
                            <div className="flex gap-6 text-sm font-bold uppercase tracking-widest text-ink-800/70">
                                <Link to="/" className="flex items-center gap-2 hover:text-bronze transition-colors">
                                    <User className="w-4 h-4" /> Collection
                                </Link>
                                <Link to="/gallery" className="flex items-center gap-2 hover:text-bronze transition-colors">
                                    <Globe className="w-4 h-4" /> Gallery
                                </Link>
                                {role === 'admin' && (
                                    <Link to="/admin" className="text-wax-red hover:text-red-700 transition-colors">Admin Console</Link>
                                )}
                            </div>

                            <div className="flex items-center gap-4">
                                {/* Notifications Bell */}
                                {isAuthenticated && (
                                    <div className="relative" ref={dropdownRef}>
                                        <button
                                            onClick={() => setShowNotifs(!showNotifs)}
                                            className="p-2 text-ink-800/60 hover:text-bronze transition-colors relative"
                                        >
                                            <Bell className="w-5 h-5" />
                                            {unreadCount > 0 && (
                                                <span className="absolute top-1 right-1 bg-wax-red text-white text-[10px] font-bold w-4 h-4 flex items-center justify-center rounded-full border-2 border-parchment-200">
                                                    {unreadCount}
                                                </span>
                                            )}
                                        </button>

                                        {/* Dropdown */}
                                        {showNotifs && (
                                            <div className="absolute right-0 mt-2 w-80 bg-white border border-ink-900/10 rounded-xl shadow-xl z-50 overflow-hidden animate-in fade-in slide-in-from-top-2 duration-200">
                                                <div className="px-4 py-3 border-b border-ink-900/5 bg-parchment-50 flex justify-between items-center">
                                                    <h3 className="text-xs font-bold uppercase tracking-wider text-ink-900/40">Notifications</h3>
                                                    {unreadCount > 0 && <span className="text-[10px] font-bold bg-wax-red/10 text-wax-red px-2 py-0.5 rounded-full">{unreadCount} New</span>}
                                                </div>
                                                <div className="max-h-96 overflow-y-auto">
                                                    {notifications.length > 0 ? (
                                                        notifications.map(n => (
                                                            <div
                                                                key={n.id}
                                                                onClick={(e) => handleMarkRead(e, n.id, n.link)}
                                                                className={`px-4 py-3 border-b border-ink-900/5 cursor-pointer hover:bg-parchment-50 transition-colors ${!n.is_read ? 'bg-bronze/5 border-l-2 border-l-bronze' : ''}`}
                                                            >
                                                                <div className="flex gap-3">
                                                                    <div className="mt-1">{getIcon(n.type)}</div>
                                                                    <div>
                                                                        <div className="text-sm font-bold text-ink-900">{n.title}</div>
                                                                        <div className="text-xs text-ink-900/60 line-clamp-2 mt-0.5">{n.message}</div>
                                                                        <div className="text-[10px] text-ink-900/30 mt-1">{new Date(n.created_at).toLocaleString()}</div>
                                                                    </div>
                                                                </div>
                                                            </div>
                                                        ))
                                                    ) : (
                                                        <div className="px-4 py-10 text-center text-ink-900/30 italic text-sm font-serif">
                                                            No notifications yet
                                                        </div>
                                                    )}
                                                </div>
                                            </div>
                                        )}
                                    </div>
                                )}

                                <span className="flex items-center gap-2 px-4 py-2 bg-parchment-50 rounded-lg border border-ink-900/10 shadow-sm">
                                    <Feather className="w-4 h-4 text-bronze" />
                                    <span className="text-sm font-medium">{user?.username || 'Guest'}</span>
                                </span>
                                <button
                                    onClick={logout}
                                    className="text-xs uppercase font-bold text-ink-800/50 hover:text-wax-red transition-colors"
                                >
                                    Sign Out
                                </button>
                            </div>
                        </nav>
                    </div>
                </div>
            </header>

            {/* Main Content */}
            <main className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
                <div className="bg-parchment-50 min-h-[600px] rounded-xl shadow-[inset_0_2px_15px_rgba(0,0,0,0.05)] border border-ink-900/5 p-8 relative overflow-hidden">
                    <div className="relative z-10">
                        {children}
                    </div>
                </div>
            </main>

            {/* Footer */}
            <footer className="text-center py-6 text-ink-900/40 text-sm font-serif italic">
                Secured by Hyperledger Fabric &bull; Running on Chaincode-as-a-Service
            </footer>
        </div>
    );
};

export default Layout;
