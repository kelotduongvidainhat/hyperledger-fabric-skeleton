import React from 'react';
import { fetchStorageURL } from '../api/client';
import { Globe, Clock, User, Paperclip } from 'lucide-react';
import { Link } from 'react-router-dom';

const GalleryAssetCard = ({ asset }) => {
    const hasAttachment = asset.attachment && asset.attachment.file_name;
    const [displayUrl, setDisplayUrl] = React.useState('');

    React.useEffect(() => {
        const getUrl = async () => {
            if (asset.imageUrl) {
                try {
                    const url = await fetchStorageURL(asset.imageUrl);
                    setDisplayUrl(url);
                } catch (e) {
                    if (asset.imageHash) setDisplayUrl(`https://ipfs.io/ipfs/${asset.imageHash}`);
                }
            }
        };
        getUrl();
    }, [asset.imageUrl, asset.imageHash]);
    return (
        <div className="bg-white rounded-xl shadow-sm hover:shadow-md transition-all duration-200 border border-ink-900/10 overflow-hidden group flex flex-col h-full">
            {/* Image Area */}
            <div className="h-48 bg-parchment-200 relative overflow-hidden flex items-center justify-center">
                {displayUrl ? (
                    <img
                        src={displayUrl}
                        alt={asset.name}
                        className="w-full h-full object-cover group-hover:scale-105 transition-transform duration-500"
                    />
                ) : (
                    <div className="text-ink-900/20 text-4xl font-serif">?</div>
                )}

                {/* Attachment Indicator */}
                {hasAttachment && (
                    <div className="absolute top-3 left-3 p-1.5 rounded-full bg-white/90 text-bronze shadow-sm backdrop-blur-md border border-ink-900/10">
                        <Paperclip size={10} className="animate-bounce" />
                    </div>
                )}

                {/* Public Badge */}
                <div className="absolute top-3 right-3 px-3 py-1 rounded-full text-[10px] font-bold tracking-widest uppercase bg-bronze/10 text-bronze border border-bronze/20 shadow-sm flex items-center gap-1.5">
                    <Globe className="w-3 h-3" />
                    Public
                </div>
            </div>

            {/* Content */}
            <div className="p-5 flex-grow">
                <div className="flex justify-between items-start mb-3">
                    <h3 className="font-serif text-lg font-bold text-ink-900 line-clamp-1" title={asset.name}>
                        {asset.name}
                    </h3>
                </div>

                <div className="space-y-3 mb-5">
                    <div className="flex items-center gap-2 text-ink-900/60">
                        <User className="w-4 h-4 text-ink-900/40" />
                        <span className="text-xs truncate" title={asset.ownerId}>
                            Owned by <span className="font-bold">{asset.ownerId?.split('::')[1] || 'Unknown'}</span>
                        </span>
                    </div>
                </div>

                <Link to={`/gallery/${asset.ID}`} className="block w-full text-center py-2.5 bg-ink-900 text-white rounded hover:bg-ink-800 transition-colors duration-200 font-serif text-sm">
                    View Public Record
                </Link>
            </div>

            {/* Footer Metadata */}
            <div className="px-5 py-3 bg-parchment-50 border-t border-ink-900/5 flex items-center justify-between text-xs text-ink-900/40">
                <span className="flex items-center gap-1">
                    <Clock className="w-3 h-3" /> Published
                </span>
                <span>{asset.lastUpdatedAt ? new Date(asset.lastUpdatedAt).toLocaleDateString() : 'N/A'}</span>
            </div>
        </div>
    );
};

export default GalleryAssetCard;
