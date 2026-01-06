import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { ShieldCheck, Database, Server, RefreshCcw, AlertTriangle, Layout, Users } from 'lucide-react';

const AdminUsers = () => {
    const { token } = useAuth();
    const [identities, setIdentities] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        const fetchUsers = async () => {
            try {
                const res = await api.get('/admin/users', { headers: { Authorization: `Bearer ${token}` } });
                setIdentities(res.data.identities);
            } catch (err) {
                console.error("Failed to fetch identities", err);
            } finally {
                setLoading(false);
            }
        };
        fetchUsers();
    }, [token]);

    if (loading) return <div className="p-8 font-serif italic">Consulting the registry...</div>;

    return (
        <div className="space-y-8">
            <div className="flex justify-between items-end border-b border-ink-800/20 pb-4">
                <div>
                    <h2 className="text-3xl font-serif text-ink-800">Identity Audit</h2>
                    <p className="text-xs uppercase tracking-widest text-ink-800/50">Cross-Referencing On-Chain & Off-Chain Data</p>
                </div>
                <div className="flex gap-2">
                    <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-1 bg-green-100 text-green-800 rounded border border-green-200">
                        <Server size={12} /> Fabric CA
                    </span>
                    <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-1 bg-amber-100 text-amber-800 rounded border border-amber-200">
                        <Database size={12} /> DB: Pending Sync
                    </span>
                </div>
            </div>

            {/* Admin Nav */}
            <div className="flex gap-4">
                <AdminNavLink to="/admin" label="Overview" icon={<Layout size={14} />} />
                <AdminNavLink to="/admin/users" label="Identity Audit" icon={<Users size={14} />} active />
                <AdminNavLink to="/admin/assets" label="Global Inventory" icon={<Database size={14} />} />
            </div>

            <div className="bg-white border-2 border-ink-800/10 rounded-lg overflow-hidden shadow-sm">
                <table className="w-full text-left border-collapse">
                    <thead className="bg-parchment-200 text-ink-800 uppercase text-[10px] font-bold tracking-widest">
                        <tr>
                            <th className="px-6 py-4">Identity Name</th>
                            <th className="px-6 py-4">Type</th>
                            <th className="px-6 py-4">On-Chain Status</th>
                            <th className="px-6 py-4">DB Profile</th>
                            <th className="px-6 py-4">Last Auth</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-ink-800/5">
                        {identities.map((id) => (
                            <tr key={id.name} className="hover:bg-parchment-50 transition-colors">
                                <td className="px-6 py-4 font-medium text-ink-900">{id.name}</td>
                                <td className="px-6 py-4">
                                    <span className="text-[10px] px-2 py-0.5 bg-parchment-100 border border-ink-800/10 rounded-full font-bold uppercase">
                                        {id.type}
                                    </span>
                                </td>
                                <td className="px-6 py-4">
                                    <div className="flex items-center gap-1.5 text-green-700 font-bold text-xs uppercase tracking-tighter">
                                        <ShieldCheck size={14} /> On-Chain
                                    </div>
                                </td>
                                <td className="px-6 py-4">
                                    <div className="flex flex-col">
                                        <div className="text-xs font-bold text-ink-800">{id.email}</div>
                                        <div className="text-[10px] text-ink-800/40 uppercase tracking-widest flex items-center gap-1">
                                            <div className="w-1.5 h-1.5 rounded-full bg-green-500"></div> {id.db_status}
                                        </div>
                                    </div>
                                </td>
                                <td className="px-6 py-4 text-xs text-ink-800/40 font-serif italic">
                                    {new Date().toLocaleDateString()}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
            </div>

            <div className="p-4 bg-amber-50 border border-amber-200 rounded-lg flex gap-3 items-start">
                <AlertTriangle className="text-amber-600 shrink-0" size={20} />
                <div className="text-xs text-amber-800 leading-relaxed">
                    <strong>Technical Note:</strong> Off-chain database profiles (PostgreSQL) are presently being implemented.
                    The table above shows real identities from the Fabric CA, but detailed user metadata (emails, full names, assigned roles)
                    will be populated once the synchronization worker is activated in <strong>Phase 2</strong>.
                </div>
            </div>
        </div>
    );
};

const AdminNavLink = ({ to, label, icon, active }) => (
    <Link
        to={to}
        className={`flex items-center gap-2 px-4 py-2 rounded-lg text-[10px] font-bold uppercase tracking-widest transition-all border ${active ? 'bg-ink-800 text-parchment-100 border-ink-800 shadow-md' : 'bg-white text-ink-800/60 border-ink-800/10 hover:border-bronze hover:text-bronze shadow-sm'}`}
    >
        {icon} {label}
    </Link>
);

export default AdminUsers;
