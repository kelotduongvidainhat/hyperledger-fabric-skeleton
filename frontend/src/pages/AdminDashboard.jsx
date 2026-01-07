import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { Layout, Users, Activity, BarChart3, Database, ShieldAlert, RefreshCcw } from 'lucide-react';

const AdminDashboard = () => {
    const { token } = useAuth();
    const [stats, setStats] = useState(null);
    const [users, setUsers] = useState([]);
    const [recentAssets, setRecentAssets] = useState([]);
    const [loading, setLoading] = useState(true);
    const [syncing, setSyncing] = useState(false);

    const fetchAdminData = async () => {
        try {
            const [statsRes, usersRes, assetsRes] = await Promise.all([
                api.get('/admin/stats', { headers: { Authorization: `Bearer ${token}` } }),
                api.get('/admin/users', { headers: { Authorization: `Bearer ${token}` } }),
                api.get('/admin/assets?source=blockchain', { headers: { Authorization: `Bearer ${token}` } })
            ]);
            setStats(statsRes.data);
            setUsers(usersRes.data.identities);
            setRecentAssets(assetsRes.data.assets || []);
        } catch (error) {
            console.error("Failed to fetch admin data", error);
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        if (token) fetchAdminData();
    }, [token]);

    const handleSyncLedger = async () => {
        if (!window.confirm("Synchronize database with current blockchain state?")) return;
        setSyncing(true);
        try {
            await api.post('/admin/sync', {}, { headers: { Authorization: `Bearer ${token}` } });
            await fetchAdminData();
            alert("Ledger synchronization successful.");
        } catch (error) {
            console.error("Sync failed", error);
            alert("Sync failed: " + (error.response?.data?.error || error.message));
        } finally {
            setSyncing(false);
        }
    };

    if (loading) return (
        <div className="min-h-screen bg-parchment-100 p-8 flex items-center justify-center font-serif flex-col gap-4">
            <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-bronze"></div>
            Deeply pondering the ledger...
        </div>
    );

    return (
        <div className="min-h-screen bg-parchment-100 p-6 md:p-12 font-sans text-ink-900">
            {/* Header */}
            <header className="mb-12 border-b-2 border-ink-800 pb-6 flex justify-between items-end">
                <div>
                    <h1 className="text-4xl font-serif text-ink-800 mb-2 italic">Admin Console</h1>
                    <p className="text-sm uppercase tracking-widest text-ink-800 opacity-70">Hyperledger Fabric Oversight & Governance</p>
                </div>
                <div className="flex flex-col items-end gap-3">
                    <span className="bg-wax-red text-white px-3 py-1 rounded text-xs font-bold uppercase tracking-tighter shadow-sm flex items-center gap-2">
                        <ShieldAlert size={14} /> System Secure
                    </span>
                    <button
                        onClick={handleSyncLedger}
                        disabled={syncing}
                        className={`flex items-center gap-2 px-4 py-2 bg-bronze text-white rounded text-[10px] font-bold uppercase tracking-widest transition-all hover:bg-ink-800 shadow-md disabled:opacity-50`}
                    >
                        <RefreshCcw size={14} className={syncing ? 'animate-spin' : ''} />
                        {syncing ? 'Syncing...' : 'Sync Ledger'}
                    </button>
                </div>
            </header>

            {/* Navigation / Quick Actions */}
            <div className="flex gap-4 mb-12">
                <AdminNavLink to="/admin" label="Overview" icon={<Layout size={14} />} active />
                <AdminNavLink to="/admin/users" label="Identity Audit" icon={<Users size={14} />} />
                <AdminNavLink to="/admin/assets" label="Global Inventory" icon={<Database size={14} />} />
            </div>

            {/* Stats Grid */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-6 mb-12">
                <StatCard
                    title="Total Assets"
                    value={stats?.total_assets || 0}
                    icon={<BarChart3 className="text-bronze" />}
                    detail="Verified on-chain"
                />
                <StatCard
                    title="Unique Owners"
                    value={stats?.total_owners || 0}
                    icon={<Users className="text-bronze" />}
                    detail="Distributed Identities"
                />
                <StatCard
                    title="Pending Flows"
                    value={stats?.pending_transfers || 0}
                    icon={<Activity className="text-bronze" />}
                    detail="Active multi-sig requests"
                />
            </div>

            <div className="grid grid-cols-1 lg:grid-cols-2 gap-12">
                {/* User List */}
                <section className="bg-parchment-50 border-2 border-ink-800/20 rounded-lg p-6 shadow-md shadow-ink-800/5">
                    <div className="flex justify-between items-center mb-6">
                        <h2 className="text-2xl font-serif text-ink-800 flex items-center gap-2">
                            <Users size={20} className="text-bronze" /> Identity Oversight
                        </h2>
                        <Link to="/admin/users" className="text-[10px] uppercase font-bold text-bronze hover:underline">View All</Link>
                    </div>
                    <div className="space-y-3">
                        {users.slice(0, 5).map((id) => (
                            <div key={`${id.org}_${id.name}`} className="flex justify-between items-center p-4 bg-white/50 border border-ink-800/10 rounded hover:border-bronze transition-colors">
                                <div className="flex flex-col">
                                    <div className="flex items-center gap-2">
                                        <span className="font-medium tracking-tight text-ink-800">{id.name}</span>
                                        <span className={`text-[8px] px-1.5 py-0.5 rounded font-bold uppercase border ${id.role === 'admin' ? 'bg-bronze text-white border-bronze' : id.role === 'auditor' ? 'bg-blue-50 text-blue-700 border-blue-200' : 'bg-parchment-100 text-ink-800/60 border-ink-800/10'}`}>
                                            {id.role}
                                        </span>
                                    </div>
                                    <span className="text-[10px] uppercase text-ink-800/50">{id.org}</span>
                                </div>
                                <span className={`text-[10px] uppercase px-3 py-1 rounded-full font-bold border ${id.status === 'ACTIVE' ? 'bg-green-50 text-green-700 border-green-200' : 'bg-amber-50 text-amber-700 border-amber-200'}`}>
                                    {id.status}
                                </span>
                            </div>
                        ))}
                    </div>
                </section>

                {/* Recent Assets (Blockchain Source) */}
                <section className="bg-parchment-50 border-2 border-ink-800/20 rounded-lg p-6 shadow-md shadow-ink-800/5">
                    <div className="flex justify-between items-center mb-6">
                        <h2 className="text-2xl font-serif text-ink-800 flex items-center gap-2">
                            <Database size={20} className="text-bronze" /> Recent Ledger Activity
                        </h2>
                        <Link to="/admin/assets" className="text-[10px] uppercase font-bold text-bronze hover:underline">Full Inventory</Link>
                    </div>
                    <div className="space-y-3">
                        {recentAssets.slice(0, 5).map((asset) => (
                            <div key={asset.ID} className="flex justify-between items-center p-4 bg-white/50 border border-ink-800/10 rounded hover:border-bronze transition-colors">
                                <div className="flex flex-col">
                                    <div className="flex items-center gap-2">
                                        <span className="font-medium tracking-tight text-ink-800">{asset.Name}</span>
                                        <span className="text-[8px] px-1.5 py-0.5 rounded font-bold uppercase border bg-ink-800 text-parchment-100 border-ink-800">
                                            {asset.Action}
                                        </span>
                                    </div>
                                    <span className="text-[9px] font-mono text-ink-800/40 uppercase tracking-tighter">ID: {asset.ID}</span>
                                </div>
                                <div className="text-right">
                                    <div className="text-[10px] font-bold text-ink-800/60">{asset.LastUpdatedBy.split('::')[1] || asset.LastUpdatedBy}</div>
                                    <div className="text-[8px] uppercase text-ink-800/30">Custodian</div>
                                </div>
                            </div>
                        ))}
                        {recentAssets.length === 0 && (
                            <div className="py-12 text-center text-ink-800/40 italic text-sm">
                                No active ledger records found.
                            </div>
                        )}
                    </div>
                </section>
            </div>
        </div>
    );
};

const StatCard = ({ title, value, icon, detail }) => (
    <div className="bg-white p-6 rounded-lg border-2 border-ink-800/10 shadow-sm border-b-4 border-b-bronze flex flex-col h-full">
        <div className="flex justify-between items-start mb-4">
            <h3 className="text-sm font-bold uppercase tracking-widest text-ink-800/50">{title}</h3>
            {icon}
        </div>
        <div className="text-5xl font-serif text-ink-800 mb-2">{value}</div>
        <div className="mt-auto text-xs text-ink-800/40 italic">{detail}</div>
    </div>
);

const AdminNavLink = ({ to, label, icon, active }) => (
    <Link
        to={to}
        className={`flex items-center gap-2 px-4 py-2 rounded-lg text-[10px] font-bold uppercase tracking-widest transition-all border ${active ? 'bg-ink-800 text-parchment-100 border-ink-800 shadow-md' : 'bg-white text-ink-800/60 border-ink-800/10 hover:border-bronze hover:text-bronze shadow-sm'}`}
    >
        {icon} {label}
    </Link>
);

export default AdminDashboard;
