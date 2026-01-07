import { Scroll, Feather, Globe, User } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const Layout = ({ children }) => {
    const { user, role, logout } = useAuth();

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
                    {/* Subtle Texture/Watermark effect could go here */}
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
