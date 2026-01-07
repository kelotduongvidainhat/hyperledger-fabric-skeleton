import React, { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { deleteAccount } from '../api/client';
import { ShieldAlert, Trash2, AlertCircle, Info } from 'lucide-react';

const Settings = () => {
    const { user, logout } = useAuth();
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState('');

    const handleDeleteAccount = async () => {
        const confirmed = window.confirm(
            "CRITICAL WARNING: This will deactivate your account on this registry. " +
            "Fabric identities and existing batch history cannot be fully erased from the blockchain, " +
            "but you will lose access to your dashboard and private records. " +
            "Are you absolutely sure you want to proceed?"
        );

        if (!confirmed) return;

        setLoading(true);
        setError('');
        try {
            await deleteAccount();
            alert("Account deactivated. You will now be signed out.");
            logout();
        } catch (err) {
            console.error("Account deletion failed", err);
            setError(err.response?.data?.error || "Failed to deactivate account. Ensure you have no active assets.");
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="max-w-2xl mx-auto py-8">
            <header className="mb-12 border-b border-ink-900/10 pb-6">
                <h1 className="text-3xl font-serif text-ink-900 mb-2 italic">Account Settings</h1>
                <p className="text-xs uppercase tracking-widest text-ink-900/50">Personal Identity & Privacy Control</p>
            </header>

            <div className="space-y-8">
                {/* Profile Info */}
                <section className="bg-white p-6 rounded-xl border border-ink-900/10 shadow-sm">
                    <h2 className="text-sm font-bold uppercase tracking-widest text-ink-900/40 mb-6 flex items-center gap-2">
                        <Info size={14} className="text-bronze" /> Profile Information
                    </h2>
                    <div className="grid grid-cols-2 gap-6">
                        <div>
                            <label className="block text-[10px] uppercase font-bold text-ink-900/30 mb-1">Username</label>
                            <div className="text-lg font-serif">{user?.username}</div>
                        </div>
                        <div>
                            <label className="block text-[10px] uppercase font-bold text-ink-900/30 mb-1">Organization</label>
                            <div className="text-sm font-mono text-bronze">{user?.org}</div>
                        </div>
                    </div>
                </section>

                {/* Danger Zone */}
                <section className="bg-wax-red/[0.02] p-8 rounded-xl border-2 border-wax-red/10">
                    <div className="flex items-center gap-3 mb-6 text-wax-red">
                        <ShieldAlert size={20} />
                        <h2 className="text-sm font-bold uppercase tracking-widest">Danger Zone</h2>
                    </div>

                    <div className="space-y-4">
                        <div className="p-4 bg-white border border-wax-red/20 rounded-lg flex gap-4 items-start shadow-sm">
                            <AlertCircle className="text-wax-red shrink-0" size={20} />
                            <div className="space-y-2">
                                <h4 className="text-sm font-bold text-ink-800">Deactivate Account</h4>
                                <p className="text-xs text-ink-900/60 leading-relaxed font-serif">
                                    By deactivating your account, you will lose access to the Ownership Registry.
                                    Your historical ledger records will persist on the blockchain (immutability),
                                    but your profile will be retracted from the app.
                                </p>
                            </div>
                        </div>

                        {error && (
                            <div className="p-3 bg-red-50 text-red-700 text-xs font-bold rounded border border-red-100 italic">
                                {error}
                            </div>
                        )}

                        <button
                            onClick={handleDeleteAccount}
                            disabled={loading}
                            className="w-full flex items-center justify-center gap-2 py-3 bg-wax-red text-white rounded-lg font-bold uppercase text-xs tracking-widest hover:bg-red-900 transition-all shadow-md disabled:opacity-50"
                        >
                            <Trash2 size={16} />
                            {loading ? "Deactivating..." : "Permanently Deactivate My Account"}
                        </button>
                    </div>
                </section>
            </div>
        </div>
    );
};

export default Settings;
