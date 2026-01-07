import React, { useEffect, useState } from 'react';
import { useParams, useNavigate, useLocation, Link } from 'react-router-dom';
import { fetchAssets, fetchAssetById, fetchHistory, proposeTransfer, acceptTransfer, updateAssetView, deleteAsset } from '../api/client';
import { ArrowLeft, ArrowRight, CheckCircle, Shield, History, Eye, EyeOff, Trash2, ExternalLink, Link as LinkIcon } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

const AssetDetails = () => {
    const { id } = useParams();
    const { user } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const [asset, setAsset] = useState(null);
    const [history, setHistory] = useState([]);
    const [loading, setLoading] = useState(true);

    const isFromAdmin = location.state?.from === 'admin';
    const userFullID = user ? `${user.org}::${user.username}` : '';

    // Action State
    const [transferTarget, setTransferTarget] = useState('');
    const [actionLoading, setActionLoading] = useState(false);

    useEffect(() => {
        loadData();
    }, [id]);

    const loadData = async () => {
        try {
            const [a, h] = await Promise.all([fetchAssetById(id), fetchHistory(id)]);
            setAsset(a);
            setHistory(h);
        } catch (err) {
            alert("Error loading asset");
        } finally {
            setLoading(false);
        }
    };

    const handlePropose = async () => {
        if (!transferTarget) return;
        setActionLoading(true);
        try {
            await proposeTransfer(id, transferTarget);
            await loadData();
            setTransferTarget('');
        } catch (err) {
            alert(err.message);
        } finally {
            setActionLoading(false);
        }
    };

    const handleAccept = async () => {
        setActionLoading(true);
        try {
            await acceptTransfer(id);
            await loadData();
        } catch (err) {
            alert(err.message);
        } finally {
            setActionLoading(false);
        }
    };

    const handleUpdateView = async (newView) => {
        setActionLoading(true);
        try {
            await updateAssetView(id, newView);
            await loadData();
        } catch (err) {
            alert(err.message);
        } finally {
            setActionLoading(false);
        }
    };

    const handleDelete = async () => {
        if (!window.confirm("Are you certain you wish to permanently delete this artifact from the ledger? This action is immutable.")) return;

        setActionLoading(true);
        try {
            await deleteAsset(id);
            alert("Artifact successfully purged.");
            navigate(isFromAdmin ? "/admin/assets" : "/");
        } catch (err) {
            alert("Deletion failed: " + (err.response?.data || err.message));
        } finally {
            setActionLoading(false);
        }
    };

    if (loading || !asset) return <div className="p-10 text-center">Loading...</div>;

    const isOwner = asset.ownerId === userFullID;
    const isProposedRecipient = asset.proposedOwnerId === userFullID;
    const isPendingTransfer = asset.status === 'PENDING_TRANSFER';

    // Future-proof storage resolution
    const sourceUrl = asset.imageUrl || '';
    const displayUrl = sourceUrl.startsWith('ipfs://')
        ? sourceUrl.replace('ipfs://', 'https://ipfs.io/ipfs/')
        : sourceUrl;

    return (
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">

            {/* Left Column: Image & Actions */}
            <div className="lg:col-span-1 space-y-6">
                <Link
                    to={isFromAdmin ? "/admin/assets" : "/"}
                    className="flex items-center gap-2 text-ink-900/50 hover:text-ink-900 mb-2 no-underline"
                >
                    <ArrowLeft className="w-4 h-4" /> {isFromAdmin ? "Back to Assets" : "Back to Registry"}
                </Link>

                <div className="group relative bg-white p-2 border border-ink-900/10 rounded-xl shadow-sm overflow-hidden">
                    <a
                        href={displayUrl}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="block aspect-square bg-parchment-200 rounded-lg overflow-hidden relative cursor-zoom-in"
                        title="Open Source Record"
                    >
                        <img
                            src={displayUrl}
                            className="w-full h-full object-cover grayscale-[20%] sepia-[10%] group-hover:grayscale-0 group-hover:sepia-0 transition-all duration-500"
                            alt={asset.name}
                        />
                        <div className="absolute inset-0 bg-ink-900/0 group-hover:bg-ink-900/20 transition-all flex items-center justify-center">
                            <ExternalLink className="text-white opacity-0 group-hover:opacity-100 transition-all transform scale-50 group-hover:scale-100" />
                        </div>
                    </a>

                    {sourceUrl.startsWith('ipfs://') && (
                        <div className="mt-4 px-3 py-2 bg-parchment-50 rounded border border-ink-900/5 flex items-center justify-between">
                            <div className="flex items-center gap-2 overflow-hidden">
                                <LinkIcon size={12} className="text-bronze shrink-0" />
                                <span className="text-[10px] font-mono text-ink-900/40 truncate">{sourceUrl}</span>
                            </div>
                            <a
                                href={displayUrl}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-[10px] font-bold text-bronze hover:underline shrink-0"
                            >
                                VIEW CID
                            </a>
                        </div>
                    )}
                </div>

                {/* Actions Panel */}
                <div className="bg-white p-6 rounded-xl border border-ink-900/10 shadow-sm">
                    <h3 className="font-serif font-bold text-lg mb-4 text-ink-900">Ownership Actions</h3>

                    {isOwner && !isPendingTransfer && asset.status === 'ACTIVE' && (
                        <div className="space-y-3">
                            <label className="text-xs font-bold uppercase text-ink-900/40">Propose Transfer</label>
                            <div className="flex gap-2">
                                <input
                                    type="text"
                                    placeholder="Username (e.g. charlie)"
                                    className="flex-1 p-2 bg-parchment-50 border border-ink-900/20 rounded text-sm"
                                    value={transferTarget}
                                    onChange={e => setTransferTarget(e.target.value)}
                                />
                                <button
                                    onClick={handlePropose}
                                    disabled={actionLoading || !transferTarget}
                                    className="bg-ink-900 text-white px-4 rounded hover:bg-ink-800 disabled:opacity-50 transition-colors"
                                >
                                    <ArrowRight className="w-4 h-4" />
                                </button>
                            </div>
                        </div>
                    )}

                    {isPendingTransfer && (
                        <div className="bg-amber-50 border border-amber-200 p-4 rounded-lg">
                            <div className="text-amber-800 font-bold text-sm mb-2">Transfer Pending</div>
                            <div className="text-xs text-ink-900/70 mb-3">
                                {isProposedRecipient ?
                                    "You have been proposed as the new owner." :
                                    `Proposed Recipient: ${asset.proposedOwnerId}`
                                }
                            </div>

                            {isProposedRecipient && (
                                <button
                                    onClick={handleAccept}
                                    disabled={actionLoading}
                                    className="w-full flex justify-center items-center gap-2 bg-wax-red text-white py-2 rounded hover:bg-red-900 shadow-sm font-bold text-sm transition-colors"
                                >
                                    <CheckCircle className="w-4 h-4" /> Accept Transfer
                                </button>
                            )}
                        </div>
                    )}

                    {!isOwner && !isProposedRecipient && !isPendingTransfer && (
                        <p className="text-sm text-ink-900/40 italic">You do not have administrative rights over this artifact.</p>
                    )}

                    {isOwner && (
                        <div className="mt-6 pt-6 border-t border-ink-900/10 space-y-3">
                            <label className="text-xs font-bold uppercase text-ink-900/40">Visibility Control</label>
                            <button
                                onClick={() => handleUpdateView(asset.view?.toUpperCase() === 'PUBLIC' ? 'PRIVATE' : 'PUBLIC')}
                                disabled={actionLoading}
                                className={`w-full py-2.5 px-4 flex items-center justify-center gap-2 text-xs font-bold rounded border transition-all 
                                    ${asset.view?.toUpperCase() === 'PUBLIC'
                                        ? 'bg-parchment-50 text-ink-800 border-ink-900/20 hover:bg-ink-900 hover:text-white hover:border-ink-900'
                                        : 'bg-ink-900 text-white border-ink-900 hover:bg-ink-800'}`}
                            >
                                {asset.view?.toUpperCase() === 'PUBLIC' ? (
                                    <><EyeOff className="w-4 h-4" /> Make Private</>
                                ) : (
                                    <><Eye className="w-4 h-4" /> Make Public</>
                                )}
                            </button>
                            <p className="text-[10px] text-ink-900/40 text-center">
                                {asset.view?.toUpperCase() === 'PUBLIC'
                                    ? "Currently visible to all authenticated users."
                                    : "Currently restricted to owner and administrators."}
                            </p>

                            <button
                                onClick={handleDelete}
                                disabled={actionLoading}
                                className="w-full mt-4 py-2.5 px-4 flex items-center justify-center gap-2 text-xs font-bold rounded border border-wax-red text-wax-red hover:bg-wax-red hover:text-white transition-all disabled:opacity-50"
                            >
                                <Trash2 className="w-4 h-4" /> Delete Artifact
                            </button>
                        </div>
                    )}
                </div>
            </div>

            {/* Right Column: Details & History */}
            <div className="lg:col-span-2 space-y-8">
                <div>
                    <div className="flex justify-between items-start">
                        <div>
                            <h1 className="text-4xl font-serif font-bold text-ink-900 mb-2">{asset.name}</h1>
                            <span className="font-mono text-sm bg-parchment-200 px-2 py-1 rounded text-ink-900/60">{asset.ID}</span>
                        </div>
                        <div className="text-right">
                            <div className="text-xs text-ink-900/40 uppercase tracking-widest mb-1">Current Owner</div>
                            <div className="flex items-center gap-2 font-bold text-lg text-ink-900">
                                <Shield className="w-5 h-5 text-wax-red" />
                                {asset.ownerId}
                            </div>
                        </div>
                    </div>
                    <p className="mt-6 text-lg text-ink-900/80 leading-relaxed font-serif">
                        {asset.description || 'No description provided.'}
                    </p>
                </div>

                <div className="border-t border-ink-900/10 pt-8">
                    <h3 className="flex items-center gap-2 font-serif font-bold text-xl text-ink-900 mb-6">
                        <History className="w-5 h-5" /> Provenance History
                    </h3>

                    <div className="space-y-0 relative border-l-2 border-parchment-300 ml-3">
                        {history.map((record, idx) => (
                            <div key={record.txId || idx} className="relative pl-8 pb-8 last:pb-0">
                                <div className="absolute -left-[9px] top-0 w-4 h-4 bg-parchment-100 border-2 border-bronze rounded-full"></div>
                                <div className="bg-white p-4 rounded-lg border border-ink-900/5 shadow-sm">
                                    <div className="flex justify-between items-start mb-1">
                                        <span className="font-bold text-ink-900 text-sm">{(record.actionType || 'UNKNOWN').replace(/_/g, ' ')}</span>
                                        <span className="text-xs text-ink-900/40 text-right">
                                            {record.timestamp ? new Date(record.timestamp).toLocaleString() : 'N/A'}
                                        </span>
                                    </div>
                                    <div className="text-xs text-ink-900/60">
                                        Actor: <span className="font-mono text-wax-red">{record.actorId}</span>
                                    </div>
                                    <div className="text-[10px] font-mono text-ink-900/30 mt-2 truncate">
                                        TX: {record.txId}
                                    </div>
                                </div>
                            </div>
                        ))}
                    </div>
                </div>
            </div>
        </div>
    );
};

export default AssetDetails;
