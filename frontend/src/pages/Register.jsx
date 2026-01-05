import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { UserPlus } from 'lucide-react';
import { api } from '../api/client';

const Register = () => {
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [error, setError] = useState('');
    const [success, setSuccess] = useState(false);
    const navigate = useNavigate();

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');

        try {
            await api.post('/auth/register', { username, password });
            setSuccess(true);
            setTimeout(() => navigate('/login'), 2000);
        } catch (err) {
            if (err.response && err.response.status === 501) {
                setError("Automatic registration is currently disabled. Please contact the administrator.");
            } else {
                setError('Registration failed. Please try again.');
            }
        }
    };

    return (
        <div className="flex items-center justify-center min-h-screen bg-parchment-white text-ink-charcoal font-sans">
            <div className="w-full max-w-md p-8 bg-antique-beige rounded-lg shadow-lg border border-bronze/30">
                <div className="flex justify-center mb-6">
                    <div className="p-3 bg-white rounded-full border border-bronze/50">
                        <UserPlus className="w-8 h-8 text-ink-sepia" />
                    </div>
                </div>
                <h2 className="text-3xl font-serif text-center mb-6 text-ink-sepia">Enlist Identity</h2>

                {error && (
                    <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded text-sm text-center">
                        {error}
                    </div>
                )}
                {success && (
                    <div className="mb-4 p-3 bg-green-100 border border-green-400 text-green-700 rounded text-sm text-center">
                        Identity requested. Redirecting...
                    </div>
                )}

                <form onSubmit={handleSubmit} className="space-y-4">
                    <div>
                        <label className="block text-sm font-semibold mb-1">New Identity ID</label>
                        <input
                            type="text"
                            className="w-full p-2 border border-bronze/30 rounded focus:ring-2 focus:ring-wax-red focus:border-transparent bg-white/50"
                            placeholder="username"
                            value={username}
                            onChange={(e) => setUsername(e.target.value)}
                        />
                    </div>
                    <div>
                        <label className="block text-sm font-semibold mb-1">Secret</label>
                        <input
                            type="password"
                            className="w-full p-2 border border-bronze/30 rounded focus:ring-2 focus:ring-wax-red focus:border-transparent bg-white/50"
                            value={password}
                            onChange={(e) => setPassword(e.target.value)}
                        />
                    </div>

                    <button
                        type="submit"
                        className="w-full py-2 bg-ink-sepia text-white font-serif font-bold rounded hover:bg-ink-charcoal transition-colors shadow-md"
                    >
                        Request Enrollment
                    </button>

                    <div className="text-center text-sm text-ink-sepia/70 mt-4">
                        <span className="cursor-pointer hover:underline" onClick={() => navigate('/login')}>Back to Login</span>
                    </div>
                </form>
            </div>
        </div>
    );
};

export default Register;
