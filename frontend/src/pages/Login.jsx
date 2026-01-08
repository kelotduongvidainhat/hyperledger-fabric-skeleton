import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { useNavigate } from 'react-router-dom';
import { Key } from 'lucide-react';

const Login = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [org, setOrg] = useState('');
    const [error, setError] = useState('');
    const { login } = useAuth();
    const navigate = useNavigate();

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');
        const success = await login(username, password, org);
        if (success) {
            navigate('/');
        } else {
            setError('Invalid credentials');
        }
    };

    return (
        <div className="flex items-center justify-center min-h-screen bg-parchment-white text-ink-charcoal font-sans">
            <div className="w-full max-w-md p-8 bg-antique-beige rounded-lg shadow-lg border border-bronze/30">
                <div className="flex justify-center mb-6">
                    <div className="p-3 bg-white rounded-full border border-bronze/50">
                        <Key className="w-8 h-8 text-wax-red" />
                    </div>
                </div>
                <h2 className="text-3xl font-serif text-center mb-6 text-ink-sepia">Access Registry</h2>

                {error && (
                    <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded text-sm text-center">
                        {error}
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-sm font-semibold mb-1">Identity ID</label>
                        <input
                            type="text"
                            className="w-full p-2 border border-bronze/30 rounded focus:ring-2 focus:ring-wax-red focus:border-transparent bg-white/50"
                            placeholder="e.g. admin"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-semibold mb-1">Organization (Optional)</label>
                        <select
                            className="w-full p-2 border border-bronze/30 rounded focus:ring-2 focus:ring-wax-red focus:border-transparent bg-white/50"
                            value={org}
                            onChange={(e) => setOrg(e.target.value)}
                        >
                            <option value="">Auto-Detect Organization</option>
                            <option value="Org1MSP">Organization 1 (Org1MSP)</option>
                            <option value="Org2MSP">Organization 2 (Org2MSP)</option>
                        </select>
                    </div>
                    <div>
                        <label className="block text-sm font-semibold mb-1">Secret</label>
                        <input
                            type="password"
                            className="w-full p-2 border border-bronze/30 rounded focus:ring-2 focus:ring-wax-red focus:border-transparent bg-white/50"
                            placeholder="••••••"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                        />
                    </div>

                    <button
                        type="submit"
                        className="w-full py-2 bg-wax-red text-white font-serif font-bold rounded hover:bg-red-900 transition-colors shadow-md"
                    >
                        Authenticate
                    </button>

                    <div className="text-center text-sm text-ink-sepia/70 mt-4">
                        <span>New to the registry? </span>
                        <span
                            className="font-semibold cursor-pointer hover:underline text-wax-red"
                            onClick={() => navigate('/register')}
                        >
                            Request Enrollment
                        </span>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default Login;
