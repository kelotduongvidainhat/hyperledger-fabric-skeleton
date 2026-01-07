import React from 'react';
import { Shield, ArrowRightLeft, Clock } from 'lucide-react';
import { Link } from 'react-router-dom';

const AssetCard = ({ asset }) => {
    const isPending = asset.status === 'PENDING_TRANSFER';
    const isDeleted = asset.status === 'DELETED';

    // Resolve IPFS URL
    const getDisplayImage = (url) => {
        if (!url) return null;
        return url.replace('ipfs://', 'https://ipfs.io/ipfs/');
    };

    return (
        <div className={`bg-white rounded-2xl shadow-sm hover:shadow-xl transition-all duration-300 border border-ink-900/10 overflow-hidden group flex flex-col h-full ${isDeleted ? 'opacity-60 grayscale' : ''}`}>
            {/* Image Area */}
            <div className="h-52 bg-parchment-200 relative overflow-hidden flex items-center justify-center">
                {asset.imageUrl ? (
                    <img
                        src={getDisplayImage(asset.imageUrl)}
                        alt={asset.name}
                        className="w-full h-full object-cover group-hover:scale-110 transition-transform duration-700 ease-out"
                    />
                ) : (
                    <div className="flex flex-col items-center gap-2 opacity-20 text-ink-900">
                        <Shield size={40} />
                        <span className="text-[10px] uppercase font-bold tracking-[0.2em]">No Image</span>
                    </div>
                )}

                {/* Status Badge */}
                <div className={`absolute top-4 right-4 px-3 py-1 rounded-full text-[9px] font-black tracking-widest uppercase border shadow-lg backdrop-blur-md ${asset.status === 'ACTIVE' ? 'bg-green-500/90 text-white border-green-400' :
                        asset.status === 'PENDING_TRANSFER' ? 'bg-amber-500/90 text-white border-amber-400' :
                            asset.status === 'FROZEN' ? 'bg-blue-600/90 text-white border-blue-400' :
                                'bg-wax-red text-white border-red-400'
                    }`}>
                    {(asset.status || 'UNKNOWN').replace('_', ' ')}
                </div>

                <div className="absolute inset-0 bg-gradient-to-t from-ink-900/40 to-transparent opacity-0 group-hover:opacity-100 transition-opacity duration-300" />
            </div>

            {/* Content */}
            <div className="p-6 flex-grow flex flex-col">
                <div className="flex justify-between items-start mb-4">
                    <div className="space-y-1">
                        <div className="text-[8px] uppercase font-bold text-ink-900/40 tracking-widest flex items-center gap-1.5 leading-none">
                            <span className="w-1.5 h-1.5 rounded-full bg-bronze animate-pulse" /> Asset ID
                        </div>
                        <h3 className="font-serif text-xl font-bold text-ink-900 leading-tight" title={asset.name}>
                            {asset.name || 'Unnamed Artifact'}
                        </h3>
                    </div>
                </div>

                <div className="space-y-3 mb-6 flex-grow">
                    {isPending && (
                        <div className="p-2.5 bg-amber-50 rounded-lg border border-amber-200/50 flex items-center gap-3">
                            <ArrowRightLeft className="w-4 h-4 text-amber-600 animate-pulse" />
                            <div className="flex flex-col">
                                <span className="text-[8px] uppercase font-bold text-amber-900/40">Proposed Transfer</span>
                                <span className="text-[10px] font-bold text-amber-900 truncate max-w-[120px]">@{asset.proposedOwnerId?.split('::')[1]}</span>
                            </div>
                        </div>
                    )}
                </div>

                <Link to={`/assets/${asset.ID}`} className="group/btn relative overflow-hidden w-full text-center py-3 bg-ink-900 text-white rounded-xl hover:bg-bronze transition-all duration-300 font-serif font-bold text-sm shadow-md">
                    <span className="relative z-10 flex items-center justify-center gap-2">
                        Inspect Record <ArrowRightLeft size={14} className="opacity-0 group-hover/btn:opacity-100 -translate-x-2 group-hover/btn:translate-x-0 transition-all" />
                    </span>
                </Link>
            </div>

            {/* Footer Metadata */}
            <div className="px-6 py-4 bg-parchment-50 border-t border-ink-900/5 flex items-center justify-between text-[10px] font-bold uppercase tracking-widest text-ink-900/40">
                <span className="flex items-center gap-2">
                    <Clock className="w-3.5 h-3.5 text-bronze/60" /> Updated
                </span>
                <span className="text-ink-900/60">{asset.lastUpdatedAt ? new Date(asset.lastUpdatedAt).toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' }) : 'N/A'}</span>
            </div>
        </div>
    );
};

export default AssetCard;
