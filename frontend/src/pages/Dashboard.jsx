import React, { useEffect, useState } from 'react';
import { fetchAssets } from '../api/client';
import AssetCard from '../components/AssetCard';
import { Plus, Search, Filter, Briefcase, ArrowRightLeft } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const Dashboard = () => {
    const { user } = useAuth();
    const [assets, setAssets] = useState([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [viewFilter, setViewFilter] = useState('all'); // all, owned, pending

    useEffect(() => {
        loadAssets();
    }, []);

    const loadAssets = async () => {
        setLoading(true);
        try {
            const data = await fetchAssets();
            setAssets(Array.isArray(data) ? data : []);
        } catch (err) {
            console.error("Failed to load assets", err);
        } finally {
            setLoading(false);
        }
    };

    const userFullID = user ? `${user.org}::${user.username}` : '';

    // Filter logic for "My Collection"
    const myAssets = assets.filter(asset => {
        const isOwner = asset.ownerId === userFullID;
        const isProposed = asset.proposedOwnerId === userFullID;
        return isOwner || isProposed;
    });

    const applyViewFilter = (items) => {
        switch (viewFilter) {
            case 'owned':
                return items.filter(a => a.ownerId === userFullID && a.status !== 'PENDING_TRANSFER');
            case 'pending':
                return items.filter(a => a.status === 'PENDING_TRANSFER' || a.status === 'PENDING');
            default:
                return items;
        }
    };

    const filteredAssets = applyViewFilter(myAssets).filter(a =>
        a.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        a.ID?.toLowerCase().includes(searchTerm.toLowerCase())
    );

    const pendingCount = myAssets.filter(a => a.status?.includes('PENDING')).length;

    return (
        <div className="space-y-8">
            {/* Header Section */}
            <div className="flex flex-col md:flex-row justify-between items-end md:items-center gap-4 border-b border-ink-900/10 pb-6">
                <div>
                    <h2 className="text-3xl font-serif text-ink-900 mb-2">My Collection</h2>
                    <p className="text-ink-900/60 max-w-xl">
                        Manage your personal artifacts and track active ownership transfers.
                    </p>
                </div>
                <Link to="/create" className="flex items-center gap-2 bg-wax-red text-white py-2.5 px-5 rounded hover:bg-red-900 transition-colors shadow-sm font-serif">
                    <Plus className="w-4 h-4" />
                    <span>Mint New Asset</span>
                </Link>
            </div>

            {/* Quick Stats */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
                <div className="bg-white p-4 border border-ink-900/10 rounded-lg flex items-center gap-4">
                    <div className="p-3 bg-parchment-100 rounded-full text-ink-900">
                        <Briefcase className="w-5 h-5" />
                    </div>
                    <div>
                        <div className="text-2xl font-serif text-ink-900">{myAssets.length}</div>
                        <div className="text-xs uppercase font-bold text-ink-900/40">Total Items</div>
                    </div>
                </div>
                <div className="bg-white p-4 border border-ink-900/10 rounded-lg flex items-center gap-4">
                    <div className="p-3 bg-wax-red/10 rounded-full text-wax-red">
                        <ArrowRightLeft className="w-5 h-5" />
                    </div>
                    <div>
                        <div className="text-2xl font-serif text-ink-900">{pendingCount}</div>
                        <div className="text-xs uppercase font-bold text-ink-900/40">Active Transfers</div>
                    </div>
                </div>
            </div>

            {/* Controls Section */}
            <div className="flex flex-col md:flex-row gap-4">
                <div className="relative flex-grow max-w-md">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-ink-900/40" />
                    <input
                        type="text"
                        placeholder="Search your collection..."
                        className="w-full pl-10 pr-4 py-2 bg-white border border-ink-900/20 rounded focus:outline-none focus:ring-1 focus:ring-wax-red focus:border-wax-red transition-all"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>

                <div className="flex bg-white border border-ink-900/20 rounded p-1">
                    <button
                        onClick={() => setViewFilter('all')}
                        className={`px-4 py-1.5 text-xs font-bold rounded transition-all ${viewFilter === 'all' ? 'bg-ink-900 text-white shadow-sm' : 'text-ink-900/60 hover:text-ink-900'}`}
                    >
                        All Items
                    </button>
                    <button
                        onClick={() => setViewFilter('owned')}
                        className={`px-4 py-1.5 text-xs font-bold rounded transition-all ${viewFilter === 'owned' ? 'bg-ink-900 text-white shadow-sm' : 'text-ink-900/60 hover:text-ink-900'}`}
                    >
                        Owned
                    </button>
                    <button
                        onClick={() => setViewFilter('pending')}
                        className={`px-4 py-1.5 text-xs font-bold rounded transition-all ${viewFilter === 'pending' ? 'bg-ink-900 text-white shadow-sm' : 'text-ink-900/60 hover:text-ink-900'}`}
                    >
                        Transfers ({pendingCount})
                    </button>
                </div>
            </div>

            {/* Grid */}
            {loading ? (
                <div className="text-center py-20 text-ink-900/40 animate-pulse font-serif">Syncing with Registry...</div>
            ) : filteredAssets.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                    {filteredAssets.map(asset => (
                        <AssetCard key={asset.ID} asset={asset} />
                    ))}
                </div>
            ) : (
                <div className="text-center py-20 border-2 border-dashed border-ink-900/10 rounded-xl">
                    <h3 className="font-serif text-xl text-ink-900/40 mb-2">No assets in your collection</h3>
                    <p className="text-ink-900/30 text-sm">Create a new asset or discover artifacts in the <Link to="/gallery" className="text-wax-red hover:underline font-bold">Public Gallery</Link>.</p>
                </div>
            )}
        </div>
    );
};

export default Dashboard;
