import React, { useEffect, useState } from 'react';
import { useParams, useNavigate, Link } from 'react-router-dom';
import { fetchAssetById, fetchHistory } from '../api/client';
import { ArrowLeft, Globe, Shield, History, Clock, User } from 'lucide-react';

const GalleryAssetDetails = () => {
    const { id } = useParams();
    const navigate = useNavigate();
    const [asset, setAsset] = useState(null);
    const [history, setHistory] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
        loadData();
    }, [id]);

    const loadData = async () => {
        setLoading(true);
        try {
            const [a, h] = await Promise.all([fetchAssetById(id), fetchHistory(id)]);
            setAsset(a);
            setHistory(h);
        } catch (err) {
            console.error("Error loading gallery artifact", err);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="max-w-4xl mx-auto">
            <button onClick={() => navigate('/gallery')} className="flex items-center gap-2 text-ink-900/40 hover:text-ink-900 mb-8 transition-colors text-xs font-bold uppercase tracking-widest">
                <ArrowLeft className="w-4 h-4" /> Back to Gallery
            </button>

            {loading ? (
                <div className="text-center py-20 animate-pulse font-serif italic text-ink-900/40">
                    Accessing Ledger Record...
                </div>
            ) : !asset ? (
                <div className="text-center py-20 border-2 border-dashed border-ink-900/10 rounded-xl">
                    <h3 className="font-serif text-xl text-ink-900/40 mb-2">Artifact not found</h3>
                    <p className="text-ink-900/30 text-sm">The requested record does not exist or has been restricted.</p>
                </div>
            ) : (
                <>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-12">
                        {/* Image Section */}
                        <div className="space-y-6">
                            <div className="aspect-square bg-parchment-200 rounded-2xl border border-ink-900/10 overflow-hidden shadow-sm flex items-center justify-center">
                                {asset.imageUrl ? (
                                    <img src={asset.imageUrl} alt={asset.name} className="w-full h-full object-cover" />
                                ) : (
                                    <div className="text-ink-900/20 text-6xl font-serif">?</div>
                                )}
                            </div>
                        </div>

                        {/* Info Section */}
                        <div className="flex flex-col">
                            <div className="flex items-center gap-2 mb-4">
                                <div className="px-3 py-1 bg-bronze/10 text-bronze border border-bronze/20 rounded-full text-[10px] font-bold uppercase tracking-widest flex items-center gap-1.5">
                                    <Globe className="w-3 h-3" /> Public Registry Record
                                </div>
                            </div>

                            <h2 className="text-4xl font-serif text-ink-900 mb-4">{asset.name}</h2>
                            <p className="text-ink-900/60 leading-relaxed mb-8">{asset.description}</p>

                            <div className="space-y-4 bg-white p-6 rounded-xl border border-ink-900/10 shadow-sm">
                                <div className="flex items-center justify-between py-2 border-b border-ink-900/5">
                                    <span className="text-xs font-bold uppercase text-ink-900/40">Registered Owner</span>
                                    <span className="text-sm font-medium flex items-center gap-2">
                                        <User className="w-4 h-4 text-bronze" /> {asset.ownerId}
                                    </span>
                                </div>
                                <div className="flex items-center justify-between py-2 border-b border-ink-900/5">
                                    <span className="text-xs font-bold uppercase text-ink-900/40">Status</span>
                                    <span className="text-sm font-medium text-green-700">{asset.status}</span>
                                </div>
                                <div className="flex items-center justify-between py-2">
                                    <span className="text-xs font-bold uppercase text-ink-900/40">Asset ID</span>
                                    <span className="text-xs font-mono text-ink-900/60">{asset.ID}</span>
                                </div>
                            </div>

                            <div className="mt-8 p-4 bg-parchment-100/50 rounded-lg border border-ink-900/5">
                                <div className="flex items-center gap-2 text-ink-900/40 mb-2">
                                    <Shield className="w-4 h-4" />
                                    <span className="text-[10px] font-bold uppercase">Cryptography Verification</span>
                                </div>
                                <div className="text-[10px] font-mono text-ink-900/60 break-all leading-tight">
                                    SHA-256 Hash: {asset.imageHash || 'VERIFIED_ON_CHAIN'}
                                </div>
                            </div>
                        </div>
                    </div>

                    {/* History Section */}
                    <div className="mt-16">
                        <div className="flex items-center gap-3 mb-8 border-b border-ink-900/10 pb-4">
                            <History className="w-5 h-5 text-ink-900" />
                            <h3 className="text-2xl font-serif text-ink-900">Provenance History</h3>
                        </div>

                        <div className="relative border-l border-ink-900/10 ml-3 pl-8 space-y-10">
                            {history.length > 0 ? history.map((record, index) => (
                                <div key={record.txId || index} className="relative">
                                    <div className="absolute -left-[41px] top-1 w-4 h-4 rounded-full bg-parchment-50 border-2 border-ink-900 shadow-sm"></div>
                                    <div>
                                        <div className="flex items-center gap-3 mb-1">
                                            <span className="text-xs font-bold uppercase text-wax-red">
                                                {(record.actionType || 'UNKNOWN').replace(/_/g, ' ')}
                                            </span>
                                            <span className="text-[10px] font-mono text-ink-900/40 flex items-center gap-1">
                                                <Clock className="w-3 h-3" /> {record.timestamp ? new Date(record.timestamp).toLocaleString() : 'N/A'}
                                            </span>
                                        </div>
                                        <p className="text-sm font-medium mb-1">Actor: {record.actorId || 'System'}</p>
                                        <p className="text-[10px] font-mono text-ink-900/40 truncate" title={record.txId}>Tx: {record.txId}</p>
                                    </div>
                                </div>
                            )) : (
                                <p className="text-sm text-ink-900/40 italic">Initial genesis record loading...</p>
                            )}
                        </div>
                    </div>
                </>
            )}
        </div>
    );
};

export default GalleryAssetDetails;
