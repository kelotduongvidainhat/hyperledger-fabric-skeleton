import React from 'react';
import { Shield, ArrowRightLeft, Clock } from 'lucide-react';
import { Link } from 'react-router-dom';

const AssetCard = ({ asset }) => {
    const isPending = asset.status === 'PENDING_TRANSFER';

    return (
        <div className="bg-white rounded-xl shadow-sm hover:shadow-md transition-all duration-200 border border-ink-900/10 overflow-hidden group">
            {/* Image Area */}
            <div className="h-48 bg-parchment-200 relative overflow-hidden flex items-center justify-center">
                {asset.imageUrl && !asset.imageUrl.includes('ipfs') ? (
                    <img src={asset.imageUrl} alt={asset.name} className="w-full h-full object-cover grayscale-[20%] sepia-[30%] group-hover:grayscale-0 group-hover:sepia-0 transition-all duration-500" />
                ) : (
                    <div className="text-ink-900/20 text-4xl font-serif">?</div>
                )}

                {/* Status Badge */}
                <div className={`absolute top-3 right-3 px-3 py-1 rounded-full text-[10px] font-bold tracking-widest uppercase border shadow-sm ${asset.status === 'ACTIVE' ? 'bg-green-50 text-green-700 border-green-200' :
                    asset.status === 'PENDING_TRANSFER' ? 'bg-amber-50 text-amber-700 border-amber-200' :
                        asset.status === 'FROZEN' ? 'bg-blue-50 text-blue-700 border-blue-200' :
                            'bg-red-50 text-red-700 border-red-200'
                    }`}>
                    {(asset.status || 'UNKNOWN').replace('_', ' ')}
                </div>
            </div>

            {/* Content */}
            <div className="p-5">
                <div className="flex justify-between items-start mb-2">
                    <h3 className="font-serif text-lg font-bold text-ink-900 line-clamp-1" title={asset.name}>
                        {asset.name?.length > 15 ? `${asset.name.substring(0, 15)}...` : (asset.name || 'Untitled')}
                    </h3>
                    <span className="text-xs font-mono text-ink-900/40 bg-parchment-100 px-1.5 py-0.5 rounded" title={asset.ID}>
                        {asset.ID?.length > 7 ? `${asset.ID.substring(0, 7)}...` : asset.ID}
                    </span>
                </div>

                <div className="space-y-2 text-sm text-ink-900/70 mb-4">
                    {isPending && (
                        <div className="flex items-center gap-2 text-amber-700">
                            <ArrowRightLeft className="w-4 h-4" />
                            <span className="font-medium tracking-tight">To: {asset.proposedOwnerId}</span>
                        </div>
                    )}
                </div>

                <Link to={`/assets/${asset.ID}`} className="block w-full text-center py-2 border border-ink-900/20 rounded hover:bg-ink-900 hover:text-white transition-colors duration-200 font-serif text-sm">
                    View Ledger Record
                </Link>
            </div>

            {/* Footer Metadata */}
            <div className="px-5 py-3 bg-parchment-50 border-t border-ink-900/5 flex items-center justify-between text-xs text-ink-900/40">
                <span className="flex items-center gap-1">
                    <Clock className="w-3 h-3" />Updated
                </span>
                <span>{asset.lastUpdatedAt ? new Date(asset.lastUpdatedAt).toLocaleDateString() : 'N/A'}</span>
            </div>
        </div>
    );
};

export default AssetCard;
