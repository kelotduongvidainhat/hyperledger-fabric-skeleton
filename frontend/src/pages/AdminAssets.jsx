import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { api } from '../api/client';
import { useAuth } from '../context/AuthContext';
import { Layers, Box, Globe, Database, HardDrive, Layout, Users, ShieldCheck, Clock, User, RefreshCcw } from 'lucide-react';

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

    useEffect(() => {
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
                                <th className="px-6 py-4">Current Custodian</th>
                                <th className="px-6 py-4">Status</th>
                                <th className="px-6 py-4 text-right">Details</th>
                            </tr>
                        </thead>
                        <tbody className="divide-y divide-ink-800/5">
                            {assets.map((asset) => (
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
                                    <td className="px-6 py-4">
                                        <div className="flex items-center gap-2 text-xs">
                                            <User size={12} className="text-bronze" />
                                            <span className="font-medium">{asset.OwnerID}</span>
                                        </div>
                                    </td>
                                    <td className="px-6 py-4">
                                        <StatusBadge status={asset.Status} />
                                    </td>
                                    <td className="px-6 py-4 text-right">
                                        <Link
                                            to={`/assets/${asset.ID}`}
                                            state={{ from: 'admin' }}
                                            className="text-[10px] uppercase font-bold text-bronze hover:text-ink-800 transition-colors border-b border-bronze/20"
                                        >
                                            View Audit
                                        </Link>
                                    </td>
                                </tr>
                            ))}
                        </tbody>
                    </table>
                </div>
            )}

            {/* Source Info Card */}
            <div className={`p-6 bg-parchment-200 rounded-lg border-l-4 transition-colors ${source === 'blockchain' ? 'border-bronze' : 'border-ink-800'}`}>
                <div className="flex gap-4 items-center">
                    <Layers className={source === 'blockchain' ? 'text-bronze' : 'text-ink-800'} size={24} />
                    <div className="space-y-1">
                        <h5 className="text-sm font-bold uppercase tracking-widest text-ink-800">Source: {source}</h5>
                        <p className="text-[11px] text-ink-800/60 leading-normal italic font-serif">
                            {source === 'blockchain'
                                ? "Displaying raw, immutable data fetched directly from the Hyperledger Fabric ledger (World State). Each entry is verified by the network endorsement policy."
                                : "Displaying optimized, off-chain data from PostgreSQL. Enhanced for rich querying and administrative performance metrics."}
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

    return (
        <span className={`inline-flex items-center gap-1.5 text-[9px] font-bold px-2.5 py-1 rounded-full uppercase border tracking-wider ${isActive ? 'bg-green-50 text-green-700 border-green-200' :
                isPending ? 'bg-amber-50 text-amber-700 border-amber-200' :
                    'bg-parchment-200 text-ink-800 border-ink-800/10'
            }`}>
            {isActive && <ShieldCheck size={10} />}
            {isPending && <Clock size={10} />}
            {status}
        </span>
    );
};

export default AdminAssets;
