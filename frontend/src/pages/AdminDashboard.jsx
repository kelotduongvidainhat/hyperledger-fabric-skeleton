import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { Layout, Users, Activity, BarChart3, Database, ShieldAlert } from 'lucide-react';

const AdminDashboard = () => {
    const { token } = useAuth();
    const [stats, setStats] = useState(null);
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchAdminData = async () => {
            try {
                const [statsRes, usersRes] = await Promise.all([
                    api.get('/admin/stats', { headers: { Authorization: `Bearer ${token}` } }),
                    api.get('/admin/users', { headers: { Authorization: `Bearer ${token}` } })
                ]);
                setStats(statsRes.data);
                setUsers(usersRes.data.identities);
            } catch (error) {
                console.error("Failed to fetch admin data", error);
            } finally {
                setLoading(false);
            }
        };

        if (token) fetchAdminData();
    }, [token]);

    if (loading) return <div className="min-h-screen bg-parchment-100 p-8 flex items-center justify-center font-serif">Deeply pondering the ledger...</div>;

    return (
        <div className="min-h-screen bg-parchment-100 p-6 md:p-12 font-sans text-ink-900">
            {/* Header */}
            <header className="mb-12 border-b-2 border-ink-800 pb-6 flex justify-between items-end">
                <div>
                    <h1 className="text-4xl font-serif text-ink-800 mb-2 italic">Admin Console</h1>
                    <p className="text-sm uppercase tracking-widest text-ink-800 opacity-70">Hyperledger Fabric Oversight & Governance</p>
                </div>
                <div className="text-right">
                    <span className="bg-wax-red text-white px-3 py-1 rounded text-xs font-bold uppercase tracking-tighter shadow-sm flex items-center gap-2">
                        <ShieldAlert size={14} /> System Secure
                    </span>
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
                    <h2 className="text-2xl font-serif text-ink-800 mb-6 flex items-center gap-2">
                        <Database size={20} className="text-bronze" /> Identity Oversight (Org1)
                    </h2>
                    <div className="space-y-3">
                        {users.slice(0, 5).map((id) => (
                            <div key={id.name} className="flex justify-between items-center p-4 bg-white/50 border border-ink-800/10 rounded hover:border-bronze transition-colors">
                                <div className="flex flex-col">
                                    <span className="font-medium tracking-tight text-ink-800">{id.name}</span>
                                    <span className="text-[10px] uppercase text-ink-800/50">Managed by Fabric-CA</span>
                                </div>
                                <span className="text-[10px] uppercase bg-parchment-200 px-3 py-1 rounded-full text-ink-800 font-bold border border-ink-800/10">
                                    {id.type}
                                </span>
                            </div>
                        ))}
                    </div>
                </section>

                {/* System Logs / Placeholder */}
                <section className="bg-parchment-50 border-2 border-ink-800/20 rounded-lg p-6 shadow-md shadow-ink-800/5">
                    <h2 className="text-2xl font-serif text-ink-800 mb-6 flex items-center gap-2">
                        <Layout size={20} className="text-bronze" /> Network Configuration
                    </h2>
                    <div className="flex flex-col gap-4 text-sm">
                        <div className="p-4 bg-ink-800 text-parchment-100 rounded font-mono text-xs overflow-x-auto whitespace-pre">
                            {`Channel: mychannel\nChaincode: basic v1.0\nPeers: 2 (Org1, Org2)\nTLS: Enabled`}
                        </div>
                        <div className="bg-white/50 border border-ink-800/10 p-6 rounded italic text-ink-800/60 leading-relaxed">
                            "The administrative layer ensures that every transaction committed to the ledger follows the immutable governance rules established at network genesis."
                        </div>
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
