import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { Layers, Box, Globe, Database, HardDrive, Layout, Users, ShieldCheck, Clock, User, RefreshCcw, Snowflake, ShieldAlert, CheckCircle, Trash2 } from 'lucide-react';
import Pagination from '../components/Pagination';

const AdminNavLink = ({ to, label, icon, active }) => {
    return (
        <Link
            to={to}
            className={`flex items-center gap-2 px-4 py-2 rounded-md text-[10px] font-bold uppercase tracking-wider transition-all ${active ? 'bg-ink-800 text-parchment-100 shadow-md' : 'text-ink-800/60 hover:text-ink-800'}`}
        >
            {icon} {label}
        </Link>
    );
};

const AdminAssets = () => {
    const { token } = useAuth();
    const [assets, setAssets] = useState([]);
    const [source, setSource] = useState('blockchain');
    const [loading, setLoading] = useState(true);
    const [syncing, setSyncing] = useState(false);
    const [note, setNote] = useState('');
    const [currentPage, setCurrentPage] = useState(1);
    const itemsPerPage = 8;

    useEffect(() => {
        setCurrentPage(1);
        fetchData();
    }, [source, token]);

    const fetchData = async () => {
        setLoading(true);
        try {
            const res = await api.get(`/admin/assets?source=${source}`, {
                headers: { Authorization: `Bearer ${token}` }
            });
            setAssets(res.data.assets || []);
            setNote(res.data.note || '');
        } catch (err) {
            console.error("Failed to fetch admin assets", err);
        } finally {
            setLoading(false);
        }
    };

    // Pagination Logic
    const indexOfLastItem = currentPage * itemsPerPage;
    const indexOfFirstItem = indexOfLastItem - itemsPerPage;
    const currentAssets = assets.slice(indexOfFirstItem, indexOfLastItem);
    const totalPages = Math.ceil(assets.length / itemsPerPage);

    const handleSync = async () => {
        setSyncing(true);
        try {
            await api.post('/admin/sync', {}, {
                headers: { Authorization: `Bearer ${token}` }
            });
            alert("Ledger synchronization successful!");
            if (source === 'database') fetchData();
        } catch (err) {
            console.error("Sync failed", err);
            alert("Synchronization failed: " + (err.response?.data?.error || err.message));
        } finally {
            setSyncing(false);
        }
    };

    const handleUpdateStatus = async (id, status) => {
        setSyncing(true);
        try {
            await api.post(`/admin/assets/${id}/status`, { status }, {
                headers: { Authorization: `Bearer ${token}` }
            });
            alert(`Asset ${id} status updated to ${status}`);
            fetchData();
        } catch (err) {
            console.error("Management action failed", err);
            alert("Action failed: " + (err.response?.data?.error || err.message));
        } finally {
            setSyncing(false);
        }
    };

    return (
        <div className="space-y-8 text-ink-900">
            <div className="flex justify-between items-end border-b border-ink-800/20 pb-4">
                <div>
                    <h2 className="text-3xl font-serif text-ink-800">Global Inventory</h2>
                    <p className="text-xs uppercase tracking-widest text-ink-800/50">Unified Ledger Oversight</p>
                </div>

                <div className="flex gap-4 items-center">
                    {/* Sync Button */}
                    <button
                        onClick={handleSync}
                        disabled={syncing}
                        className={`flex items-center gap-2 px-4 py-2 rounded-lg text-[10px] font-bold uppercase tracking-widest transition-all border border-bronze text-bronze hover:bg-bronze hover:text-white disabled:opacity-50 shadow-sm`}
                    >
                        <RefreshCcw size={14} className={syncing ? 'animate-spin' : ''} />
                        {syncing ? 'Syncing...' : 'Sync Ledger'}
                    </button>

                    {/* Source Switcher */}
                    <div className="flex bg-parchment-200 p-1 rounded-lg border border-ink-800/10 h-10">
                        <button
                            onClick={() => setSource('blockchain')}
                            className={`flex items-center gap-2 px-4 rounded-md text-[9px] font-bold uppercase tracking-wider transition-all ${source === 'blockchain' ? 'bg-ink-800 text-parchment-100 shadow-md' : 'text-ink-800/60 hover:text-ink-800'}`}
                        >
                            <Globe size={12} /> Blockchain
                        </button>
                        <button
                            onClick={() => setSource('database')}
                            className={`flex items-center gap-2 px-4 rounded-md text-[9px] font-bold uppercase tracking-wider transition-all ${source === 'database' ? 'bg-ink-800 text-parchment-100 shadow-md' : 'text-ink-800/60 hover:text-ink-800'}`}
                        >
                            <Database size={12} /> Database
                        </button>
                    </div>
                </div>
            </div>

            {/* Admin Nav */}
            <div className="flex gap-4">
                <AdminNavLink to="/admin" label="Overview" icon={<Layout size={14} />} />
                <AdminNavLink to="/admin/users" label="Identity Audit" icon={<Users size={14} />} />
                <AdminNavLink to="/admin/assets" label="Global Inventory" icon={<Database size={14} />} active />
            </div>

            {loading ? (
                <div className="p-12 text-center font-serif italic text-ink-800/40 border-2 border-dashed border-ink-800/10 rounded-xl">
                    Sifting through the {source} records...
                </div>
            ) : source === 'database' && assets.length === 0 ? (
                <div className="p-12 border-2 border-dashed border-ink-800/20 rounded-xl flex flex-col items-center text-center max-w-2xl mx-auto bg-white/30">
                    <HardDrive size={48} className="text-ink-800/20 mb-4" />
                    <h3 className="text-xl font-serif text-ink-800 mb-2">Database Cache Empty</h3>
                    <p className="text-sm text-ink-800/60 mb-6 italic leading-relaxed">
                        {note || "Off-chain storage hasn't been synchronized with the ledger yet. Asset metadata is currently only available directly on-chain."}
                    </p>
                    <button
                        onClick={() => setSource('blockchain')}
                        className="text-[10px] font-bold uppercase tracking-widest text-bronze border-b border-bronze pb-1 hover:text-ink-800 hover:border-ink-800 transition-all font-sans"
                    >
                        Return to Blockchain Source
                    </button>
                </div>
            ) : (
                <div className="bg-white border-2 border-ink-800/10 rounded-lg overflow-hidden shadow-sm">
                    <table className="w-full text-left border-collapse">
                        <thead className="bg-parchment-200 text-ink-800 uppercase text-[10px] font-bold tracking-widest">
                            <tr>
                                <th className="px-6 py-4">Asset ID</th>
                                <th className="px-6 py-4">Designation</th>
                                <th className="px-6 py-4 text-center">Custodian</th>
                                <th className="px-6 py-4 text-center">Status</th>
                                <th className="px-6 py-4 text-right">Administrative Actions</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-ink-800/5">
                            {currentAssets.map((asset) => (
                                <tr key={asset.ID} className="hover:bg-parchment-50 transition-colors group text-sm">
                                    <td className="px-6 py-4">
                                        <div className="flex items-center gap-2">
                                            <div className="p-1.5 bg-parchment-100 rounded text-ink-800/40 group-hover:text-bronze transition-colors">
                                                <Box size={14} />
                                            </div>
                                            <span className="font-mono text-xs font-bold tracking-tighter">{asset.ID}</span>
                                        </div>
                                    </td>
                                    <td className="px-6 py-4">
                                        <div>
                                            <div className="font-serif font-bold text-ink-800">{asset.Name}</div>
                                            <div className="text-[10px] text-ink-800/40 truncate max-w-[200px] italic">
                                                {asset.Description || 'No description encrypted in record'}
                                            </div>
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 text-center">
                                        <div className="flex items-center justify-center gap-2 text-[10px] font-bold uppercase tracking-tight text-ink-800/60">
                                            <User size={12} className="text-bronze" />
                                            {asset.OwnerID}
                                        </div>
                                    </td>
                                    <td className="px-6 py-4 text-center">
                                        <StatusBadge status={asset.Status} />
                                    </td>
                                    <td className="px-6 py-4">
                                        <div className="flex justify-end items-center gap-3">
                                            {asset.Status !== 'FROZEN' && asset.Status !== 'DELETED' && (
                                                <button
                                                    onClick={() => handleUpdateStatus(asset.ID, 'FROZEN')}
                                                    disabled={syncing}
                                                    title="Freeze Asset"
                                                    className="p-2 text-blue-600 hover:bg-blue-50 rounded-lg transition-colors border border-blue-100"
                                                >
                                                    <Snowflake size={14} />
                                                </button>
                                            )}
                                            {asset.Status === 'FROZEN' && (
                                                <button
                                                    onClick={() => handleUpdateStatus(asset.ID, 'ACTIVE')}
                                                    disabled={syncing}
                                                    title="Unfreeze Asset"
                                                    className="p-2 text-green-600 hover:bg-green-50 rounded-lg transition-colors border border-green-100"
                                                >
                                                    <CheckCircle size={14} />
                                                </button>
                                            )}
                                            {asset.Status !== 'DELETED' && (
                                                <button
                                                    onClick={() => handleUpdateStatus(asset.ID, 'DELETED')}
                                                    disabled={syncing}
                                                    title="Revoke/Delete Asset"
                                                    className="p-2 text-red-600 hover:bg-red-50 rounded-lg transition-colors border border-red-100"
                                                >
                                                    <Trash2 size={14} />
                                                </button>
                                            )}
                                            <Link
                                                to={`/assets/${asset.ID}`}
                                                state={{ from: 'admin' }}
                                                className="px-3 py-1.5 bg-ink-800 text-parchment-100 rounded text-[9px] font-bold uppercase tracking-widest hover:bg-black transition-all shadow-sm"
                                            >
                                                Audit
                                            </Link>
                                        </div>
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
            )}

            {/* Source Info Card */}
            <div className={`p-6 bg-parchment-200 rounded-lg border-l-4 transition-colors ${source === 'blockchain' ? 'border-bronze' : 'border-ink-800'} shadow-sm`}>
                <div className="flex gap-4 items-center">
                    <ShieldAlert className={source === 'blockchain' ? 'text-bronze' : 'text-ink-800'} size={24} />
                    <div className="space-y-1">
                        <h5 className="text-sm font-bold uppercase tracking-widest text-ink-800">Governance Console: Oversight Active</h5>
                        <p className="text-[11px] text-ink-800/60 leading-normal italic font-serif">
                            {source === 'blockchain'
                                ? "Interacting directly with Hyperledger Fabric World State. Administrative actions (Freeze/Revoke) generate a permanent, immutable record in the ledger history."
                                : "Viewing optimized off-chain cache. Note: Administrative status changes will be written both to the Blockchain and the Database for near-real-time consistency."}
                        </p>
                    </div>
                </div>
            </div>
        </div>
    );
};

const StatusBadge = ({ status }) => {
    const isActive = status === 'ACTIVE';
    const isPending = status === 'PENDING_TRANSFER';
    const isFrozen = status === 'FROZEN';
    const isDeleted = status === 'DELETED';

    return (
        <span className={`inline-flex items-center gap-1.5 text-[9px] font-bold px-2.5 py-1 rounded-full uppercase border tracking-wider ${isActive ? 'bg-green-50 text-green-700 border-green-200' :
            isPending ? 'bg-amber-50 text-amber-700 border-amber-200' :
                isFrozen ? 'bg-blue-50 text-blue-700 border-blue-200' :
                    'bg-red-50 text-red-700 border-red-200'
            }`}>
            {isActive && <ShieldCheck size={10} />}
            {isPending && <Clock size={10} />}
            {isFrozen && <Snowflake size={10} />}
            {isDeleted && <Trash2 size={10} />}
            {status}
        </span>
    );
};

export default AdminAssets;
