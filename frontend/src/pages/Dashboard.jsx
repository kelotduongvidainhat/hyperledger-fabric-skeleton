import React, { useEffect, useState } from 'react';
import { fetchAssets } from '../api/client';
import AssetCard from '../components/AssetCard';
import { Plus, Search } from 'lucide-react';
import { Link } from 'react-router-dom';

const Dashboard = () => {
    const [assets, setAssets] = useState([]);
    const [loading, setLoading] = useState(true);
    const [filter, setFilter] = useState('');

    useEffect(() => {
        loadAssets();
    }, []);

    const loadAssets = async () => {
        try {
            const data = await fetchAssets();
            setAssets(Array.isArray(data) ? data : []);
        } catch (err) {
            console.error("Failed to load assets", err);
        } finally {
            setLoading(false);
        }
    };

    const filteredAssets = assets.filter(a =>
        a.Name.toLowerCase().includes(filter.toLowerCase()) ||
        a.ID.toLowerCase().includes(filter.toLowerCase())
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

            {/* Filter Bar */}
            <div className="relative max-w-md">
                <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-ink-900/40" />
                <input
                    type="text"
                    placeholder="Search by Name or ID..."
                    className="w-full pl-10 pr-4 py-2 bg-white border border-ink-900/20 rounded focus:outline-none focus:ring-1 focus:ring-wax-red focus:border-wax-red transition-all"
                    value={filter}
                    onChange={(e) => setFilter(e.target.value)}
                />
            </div>

            {/* Grid */}
            {loading ? (
                <div className="text-center py-20 text-ink-900/40 animate-pulse font-serif">Loading Ledger Data...</div>
            ) : filteredAssets.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                    {filteredAssets.map(asset => (
                        <AssetCard key={asset.ID} asset={asset} />
                    ))}
                </div>
            ) : (
                <div className="text-center py-20 border-2 border-dashed border-ink-900/10 rounded-xl">
                    <h3 className="font-serif text-xl text-ink-900/40 mb-2">The Registry is Empty</h3>
                    <p className="text-ink-900/30 text-sm">Or no assets matched your search.</p>
                </div>
            )}
        </div>
    );
};

export default Dashboard;
