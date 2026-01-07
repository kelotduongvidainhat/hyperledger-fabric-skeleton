import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { ShieldCheck, Database, Server, RefreshCcw, AlertTriangle, Layout, Users, Ban, CheckCircle, UserCheck, ShieldAlert } from 'lucide-react';
import Pagination from '../components/Pagination';
import { Box } from 'lucide-react';

const AdminUsers = () => {
    const { token } = useAuth();
    const [identities, setIdentities] = useState([]);
    const [loading, setLoading] = useState(true);
    const [actionLoading, setActionLoading] = useState(null);
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 8;

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

    useEffect(() => {
        fetchUsers();
    }, [token]);

    const handleUpdateStatus = async (username, status, role = "") => {
        setActionLoading(username);
        try {
            await api.post(`/admin/users/${username}/status`, { status, role }, {
                headers: { Authorization: `Bearer ${token}` }
            });
            await fetchUsers();
        } catch (err) {
            console.error("Update failed", err);
            alert("Failed to update user: " + (err.response?.data?.error || err.message));
        } finally {
            setActionLoading(null);
        }
    };

    if (loading) return <div className="p-8 font-serif italic">Consulting the registry...</div>;

    const indexOfLastItem = currentPage * itemsPerPage;
    const indexOfFirstItem = indexOfLastItem - itemsPerPage;
    const currentIdentities = identities.slice(indexOfFirstItem, indexOfLastItem);
    const totalPages = Math.ceil(identities.length / itemsPerPage);

    return (
        <div className="space-y-8">
            <div className="flex justify-between items-end border-b border-ink-800/20 pb-4">
                <div>
                    <h2 className="text-3xl font-serif text-ink-800">Identity Audit</h2>
                    <p className="text-xs uppercase tracking-widest text-ink-800/50">Unified Governance Control</p>
                </div>
                <div className="flex gap-2">
                    <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-1 bg-green-100 text-green-800 rounded border border-green-200 shadow-sm">
                        <Server size={12} /> Fabric CA: Online
                    </span>
                    <span className="flex items-center gap-1 text-[10px] font-bold px-2 py-1 bg-ink-800 text-parchment-100 rounded border border-ink-800 shadow-sm">
                        <Database size={12} /> DB: Integrated
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
                            <th className="px-6 py-4">Organization</th>
                            <th className="px-6 py-4">Account Status</th>
                            <th className="px-6 py-4 text-right">Administrative Actions</th>
                        </tr>
                    </thead>
                    <tbody className="divide-y divide-ink-800/5">
                        {currentIdentities.map((id) => (
                            <tr key={`${id.org}_${id.name}`} className="hover:bg-parchment-50 transition-colors group">
                                <td className="px-6 py-4">
                                    <div className="flex flex-col">
                                        <div className="flex items-center gap-2">
                                            <span className="font-bold text-ink-900">{id.name}</span>
                                            <span className={`text-[8px] px-1.5 py-0.5 rounded font-bold uppercase border ${id.role === 'admin' ? 'bg-bronze text-white border-bronze' : id.role === 'auditor' ? 'bg-blue-50 text-blue-700 border-blue-200' : 'bg-parchment-100 text-ink-800/60 border-ink-800/10'}`}>
                                                {id.role}
                                            </span>
                                        </div>
                                        <span className="text-[10px] text-ink-800/40 italic truncate max-w-[150px]">{id.email}</span>
                                    </div>
                                </td>
                                <td className="px-6 py-4">
                                    <div className={`inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-[10px] font-bold border ${id.org === 'Org1MSP'
                                        ? 'bg-blue-50 text-blue-700 border-blue-200'
                                        : 'bg-indigo-50 text-indigo-700 border-indigo-200'
                                        }`}>
                                        <Box size={10} />
                                        {id.org || 'Unknown Org'}
                                    </div>
                                </td>
                                <td className="px-6 py-4">
                                    <StatusBadge status={id.status} />
                                </td>
                                <td className="px-6 py-4 text-right">
                                    {id.name !== 'admin' && (
                                        <div className="flex justify-end gap-2 text-[9px] font-bold uppercase">
                                            {id.status === 'PENDING' && (
                                                <button
                                                    onClick={() => handleUpdateStatus(id.name, 'ACTIVE')}
                                                    disabled={actionLoading === id.name}
                                                    className="flex items-center gap-1.5 px-3 py-1 bg-green-600 text-white rounded hover:bg-green-700 transition-all disabled:opacity-50"
                                                >
                                                    <CheckCircle size={12} /> Approve
                                                </button>
                                            )}
                                            {id.status === 'ACTIVE' && (
                                                <>
                                                    <button
                                                        onClick={() => handleUpdateStatus(id.name, 'BANNED')}
                                                        disabled={actionLoading === id.name}
                                                        className="flex items-center gap-1.5 px-3 py-1 border border-red-200 text-red-600 rounded hover:bg-red-600 hover:text-white transition-all disabled:opacity-50"
                                                    >
                                                        <Ban size={12} /> Ban
                                                    </button>
                                                    {id.role !== 'auditor' && (
                                                        <button
                                                            onClick={() => handleUpdateStatus(id.name, '', 'auditor')}
                                                            disabled={actionLoading === id.name}
                                                            className="flex items-center gap-1.5 px-3 py-1 border border-blue-200 text-blue-600 rounded hover:bg-blue-600 hover:text-white transition-all disabled:opacity-50"
                                                        >
                                                            <UserCheck size={12} /> Promote
                                                        </button>
                                                    )}
                                                    {id.role === 'auditor' && (
                                                        <button
                                                            onClick={() => handleUpdateStatus(id.name, '', 'user')}
                                                            disabled={actionLoading === id.name}
                                                            className="flex items-center gap-1.5 px-3 py-1 border border-parchment-500 text-ink-800/60 rounded hover:bg-ink-800 hover:text-white transition-all disabled:opacity-50"
                                                        >
                                                            Demote
                                                        </button>
                                                    )}
                                                </>
                                            )}
                                            {id.status === 'BANNED' && (
                                                <button
                                                    onClick={() => handleUpdateStatus(id.name, 'ACTIVE')}
                                                    disabled={actionLoading === id.name}
                                                    className="flex items-center gap-1.5 px-3 py-1 bg-ink-800 text-white rounded text-[9px] font-bold uppercase hover:bg-black transition-all disabled:opacity-50"
                                                >
                                                    <RefreshCcw size={12} /> Re-Activate
                                                </button>
                                            )}
                                        </div>
                                    )}
                                    {id.name === 'admin' && <span className="text-[9px] font-bold text-ink-800/20 italic">Root Authority</span>}
                                </td>
                            </tr>
                        ))}
                    </tbody>
                </table>
                <Pagination
                    currentPage={currentPage}
                    totalPages={totalPages}
                    onPageChange={setCurrentPage}
                />
            </div>

            <div className="p-6 bg-parchment-200 rounded-lg border-l-4 border-bronze flex gap-4 items-start shadow-sm">
                <ShieldAlert className="text-bronze shrink-0" size={24} />
                <div className="space-y-1">
                    <h4 className="text-sm font-bold uppercase tracking-widest text-ink-800">Governance Security Policy</h4>
                    <p className="text-[11px] text-ink-800/70 leading-relaxed font-serif italic">
                        Banning a user blocks application-level access immediately via PostgreSQL status checks.
                        Root Admin identity is protected from self-modification. All status changes are logged
                        to the off-chain audit trail for compliance verification.
                    </p>
                </div>
            </div>
        </div>
    );
};

const StatusBadge = ({ status }) => {
    const isActive = status === 'ACTIVE';
    const isPending = status === 'PENDING';
    const isBanned = status === 'BANNED';
    const isDeleted = status === 'DELETED';

    return (
        <span className={`inline-flex items-center gap-1.5 text-[9px] font-bold px-2.5 py-1 rounded-full uppercase border tracking-wider ${isActive ? 'bg-green-50 text-green-700 border-green-200' :
                isPending ? 'bg-amber-50 text-amber-700 border-amber-200' :
                    isDeleted ? 'bg-slate-50 text-slate-700 border-slate-200' :
                        'bg-red-50 text-red-700 border-red-200'
            }`}>
            <div className={`w-1.5 h-1.5 rounded-full ${isActive ? 'bg-green-500' :
                    isPending ? 'bg-amber-500' :
                        isDeleted ? 'bg-slate-400' :
                            'bg-red-500'
                }`}></div>
            {status}
        </span>
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
