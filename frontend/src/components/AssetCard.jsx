import React from 'react';
import { Shield, ArrowRightLeft, Clock } from 'lucide-react';
import { Link } from 'react-router-dom';

const AssetCard = ({ asset }) => {
    const isPending = asset.Status === 'PENDING_TRANSFER';

    return (
        <div className="bg-white rounded-xl shadow-sm hover:shadow-md transition-all duration-200 border border-ink-900/10 overflow-hidden group">
            {/* Image Area */}
            <div className="h-48 bg-parchment-200 relative overflow-hidden flex items-center justify-center">
                {asset.ImageURL && !asset.ImageURL.includes('ipfs') ? (
                    <img src={asset.ImageURL} alt={asset.Name} className="w-full h-full object-cover grayscale-[20%] sepia-[30%] group-hover:grayscale-0 group-hover:sepia-0 transition-all duration-500" />
                ) : (
                    <div className="text-ink-900/20 text-4xl font-serif">?</div>
                )}

                {/* Status Badge */}
                <div className={`absolute top-3 right-3 px-3 py-1 rounded-full text-xs font-bold tracking-wider uppercase border ${isPending ? 'bg-bronze/10 text-bronze border-bronze/20' : 'bg-green-900/5 text-green-900 border-green-900/10'}`}>
                    {asset.Status.replace('_', ' ')}
                </div>
            </div>

            {/* Content */}
            <div className="p-5">
                <div className="flex justify-between items-start mb-2">
                    <h3 className="font-serif text-lg font-bold text-ink-900 line-clamp-1">{asset.Name}</h3>
                    <span className="text-xs font-mono text-ink-900/40 bg-parchment-100 px-1.5 py-0.5 rounded">{asset.ID}</span>
                </div>

                <div className="space-y-2 text-sm text-ink-900/70 mb-4">
                    <div className="flex items-center gap-2">
                        <Shield className="w-4 h-4 text-wax-red" />
                        <span className="font-medium">Owner:</span> {asset.OwnerID}
                    </div>
                    {isPending && (
                        <div className="flex items-center gap-2 text-bronze">
                            <ArrowRightLeft className="w-4 h-4" />
                            <span>To: {asset.ProposedOwnerID}</span>
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
                <span>{new Date(asset.LastUpdatedAt).toLocaleDateString()}</span>
            </div>
        </div>
    );
};

export default AssetCard;
