import React, { useEffect, useState } from 'react';
import { fetchAssets } from '../api/client';
import AssetCard from '../components/AssetCard';
import { Plus, Search, Filter } from 'lucide-react';
import { Link } from 'react-router-dom';
import { useAuth } from '../context/AuthContext';

const Dashboard = () => {
    const { user } = useAuth();
    const [assets, setAssets] = useState([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');
    const [viewFilter, setViewFilter] = useState('all'); // all, mine, public, incoming, outgoing

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

    const getVisibleAssets = () => {
        return assets.filter(asset => {
            const isPublic = asset.View === 'Public';
            const isOwner = asset.OwnerID === userFullID;
            const isProposed = asset.ProposedOwnerID === userFullID;
            return isPublic || isOwner || isProposed;
        });
    };

    const applyViewFilter = (visibleAssets) => {
        switch (viewFilter) {
            case 'mine':
                return visibleAssets.filter(a => a.OwnerID === userFullID);
            case 'public':
                return visibleAssets.filter(a => a.View === 'Public');
            case 'incoming':
                return visibleAssets.filter(a => a.ProposedOwnerID === userFullID && a.Status === 'PENDING_TRANSFER');
            case 'outgoing':
                return visibleAssets.filter(a => a.OwnerID === userFullID && a.Status === 'PENDING_TRANSFER');
            default:
                return visibleAssets;
        }
    };

    const visibleItems = getVisibleAssets();
    const filteredByView = applyViewFilter(visibleItems);

    const finalAssets = filteredByView.filter(a =>
        a.Name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        a.ID.toLowerCase().includes(searchTerm.toLowerCase())
    );

    return (
        <div className="space-y-8">
            {/* Header Section */}
            <div className="flex flex-col md:flex-row justify-between items-end md:items-center gap-4 border-b border-ink-900/10 pb-6">
                <div>
                    <h2 className="text-3xl font-serif text-ink-900 mb-2">Registry Artifacts</h2>
                    <p className="text-ink-900/60 max-w-xl">
                        Browse the immutable record of ownership secured by the Chaincode.
                    </p>
                </div>
                <Link to="/create" className="flex items-center gap-2 bg-wax-red text-white py-2.5 px-5 rounded hover:bg-red-900 transition-colors shadow-sm font-serif">
                    <Plus className="w-4 h-4" />
                    <span>Mint New Asset</span>
                </Link>
            </div>

            {/* Controls Section */}
            <div className="flex flex-col md:flex-row gap-4">
                {/* Search Bar */}
                <div className="relative flex-grow max-w-md">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-ink-900/40" />
                    <input
                        type="text"
                        placeholder="Search by Name or ID..."
                        className="w-full pl-10 pr-4 py-2 bg-white border border-ink-900/20 rounded focus:outline-none focus:ring-1 focus:ring-wax-red focus:border-wax-red transition-all"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>

                {/* Filter Select */}
                <div className="relative">
                    <Filter className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-ink-900/40" />
                    <select
                        className="pl-10 pr-4 py-2 bg-white border border-ink-900/20 rounded focus:outline-none focus:ring-1 focus:ring-wax-red transition-all appearance-none min-w-[180px]"
                        value={viewFilter}
                        onChange={(e) => setViewFilter(e.target.value)}
                    >
                        <option value="all">All Visible Assets</option>
                        <option value="mine">My Collection</option>
                        <option value="public">Public Gallery</option>
                        <option value="incoming">Pending (Transfer In)</option>
                        <option value="outgoing">Pending (Transfer Out)</option>
                    </select>
                </div>
            </div>

            {/* Grid */}
            {loading ? (
                <div className="text-center py-20 text-ink-900/40 animate-pulse font-serif">Loading Ledger Data...</div>
            ) : finalAssets.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                    {finalAssets.map(asset => (
                        <AssetCard key={asset.ID} asset={asset} />
                    ))}
                </div>
            ) : (
                <div className="text-center py-20 border-2 border-dashed border-ink-900/10 rounded-xl">
                    <h3 className="font-serif text-xl text-ink-900/40 mb-2">No artifacts found</h3>
                    <p className="text-ink-900/30 text-sm">Adjustment your filters or create a new asset.</p>
                </div>
            )}
        </div>
    );
};

export default Dashboard;
