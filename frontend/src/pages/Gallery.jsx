import React, { useEffect, useState } from 'react';
import { fetchAssets } from '../api/client';
import GalleryAssetCard from '../components/GalleryAssetCard';
import { Search, Globe, Info } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

const Gallery = () => {
    const { user } = useAuth();
    const [assets, setAssets] = useState([]);
    const [loading, setLoading] = useState(true);
    const [searchTerm, setSearchTerm] = useState('');

    useEffect(() => {
        loadAssets();
    }, []);

    const loadAssets = async () => {
        setLoading(true);
        try {
            const data = await fetchAssets();
            setAssets(Array.isArray(data) ? data : []);
        } catch (err) {
            console.error("Failed to load gallery", err);
        } finally {
            setLoading(false);
        }
    };

    // Filter logic for "Public Gallery"
    const publicAssets = assets.filter(asset => {
        return asset.view?.toUpperCase() === 'PUBLIC';
    });

    const filteredAssets = publicAssets.filter(a =>
        a.name?.toLowerCase().includes(searchTerm.toLowerCase()) ||
        a.ID?.toLowerCase().includes(searchTerm.toLowerCase())
    );

    return (
        <div className="space-y-8">
            {/* Header Section */}
            <div className="border-b border-ink-900/10 pb-6">
                <div className="flex items-center gap-3 mb-2">
                    <Globe className="w-6 h-6 text-wax-red" />
                    <h2 className="text-3xl font-serif text-ink-900">Public Gallery</h2>
                </div>
                <p className="text-ink-900/60 max-w-2xl">
                    Discover artifacts made available to the public. These records are immutable and verified by the Hyperledger Fabric network.
                </p>
            </div>

            {/* Info Box */}
            <div className="bg-parchment-100/50 border border-ink-900/5 rounded-lg p-4 flex items-start gap-3">
                <Info className="w-5 h-5 text-ink-900/40 mt-0.5" />
                <p className="text-xs text-ink-900/60 leading-relaxed">
                    Artifacts in the gallery are visible to all authenticated users. Private tokens are restricted to their respective owners and will not appear here.
                </p>
            </div>

            {/* Controls Section */}
            <div className="flex flex-col md:flex-row gap-4">
                <div className="relative flex-grow max-w-md">
                    <Search className="absolute left-3 top-1/2 -translate-y-1/2 w-4 h-4 text-ink-900/40" />
                    <input
                        type="text"
                        placeholder="Search the gallery..."
                        className="w-full pl-10 pr-4 py-2 bg-white border border-ink-900/20 rounded focus:outline-none focus:ring-1 focus:ring-wax-red focus:border-wax-red transition-all"
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                    />
                </div>
                <div className="flex-grow"></div>
                <div className="text-xs font-bold text-ink-900/40 uppercase self-center">
                    Showing {filteredAssets.length} Artifacts
                </div>
            </div>

            {/* Grid */}
            {loading ? (
                <div className="text-center py-20 text-ink-900/40 animate-pulse font-serif">Querying Global Ledger...</div>
            ) : filteredAssets.length > 0 ? (
                <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
                    {filteredAssets.map(asset => (
                        <GalleryAssetCard key={asset.ID} asset={asset} />
                    ))}
                </div>
            ) : (
                <div className="text-center py-20 border-2 border-dashed border-ink-900/10 rounded-xl">
                    <h3 className="font-serif text-xl text-ink-900/40 mb-2">The gallery is empty</h3>
                    <p className="text-ink-900/30 text-sm">No public artifacts have been minted yet.</p>
                </div>
            )}
        </div>
    );
};

export default Gallery;
