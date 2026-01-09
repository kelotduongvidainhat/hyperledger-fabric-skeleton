import React, { useState, useEffect } from 'react';
import { Link, useParams, useNavigate, useLocation } from 'react-router-dom';
import { fetchAssets, fetchAssetById, fetchHistory, proposeTransfer, acceptTransfer, updateAssetView, deleteAsset, fetchBlockchainAsset, fetchStorageURL } from '../api/client';
import { ArrowLeft, ArrowRight, CheckCircle, Shield, History, Eye, EyeOff, Trash2, Paperclip, ExternalLink, Link as LinkIcon, Database, Verified, FileText, Download } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

const AssetDetails = () => {
    const { id } = useParams();
    const { user } = useAuth();
    const navigate = useNavigate();
    const location = useLocation();
    const [asset, setAsset] = useState(null);
    const [history, setHistory] = useState([]);
    const [loading, setLoading] = useState(true);
    const [blockchainData, setBlockchainData] = useState(null);
    const [showBlockchainModal, setShowBlockchainModal] = useState(false);
    const [displayUrl, setDisplayUrl] = useState('');
    const [attachmentUrl, setAttachmentUrl] = useState(''); // This will be the View URL
    const [attachmentDownloadUrl, setAttachmentDownloadUrl] = useState('');

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

            // Fetch Pre-signed URL for Main Image (MinIO first)
            if (a.imageUrl) {
                try {
                    const url = await fetchStorageURL(a.imageUrl);
                    setDisplayUrl(url);
                } catch (e) {
                    // Fallback to IPFS if MinIO fails and it looks like a CID
                    if (a.imageHash) setDisplayUrl(`https://ipfs.io/ipfs/${a.imageHash}`);
                }
            }

            // Fetch Pre-signed URLs for Attachment (View & Download)
            if (a.attachment?.storage_path) {
                try {
                    const viewUrl = await fetchStorageURL(a.attachment.storage_path, false);
                    const downloadUrl = await fetchStorageURL(a.attachment.storage_path, true);
                    setAttachmentUrl(viewUrl);
                    setAttachmentDownloadUrl(downloadUrl);
                } catch (e) {
                    console.error("Failed to fetch attachment URLs", e);
                }
            }
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

    const handleVerifyBlockchain = async () => {
        setActionLoading(true);
        try {
            const data = await fetchBlockchainAsset(id);
            setBlockchainData(data);
            setShowBlockchainModal(true);
        } catch (err) {
            alert("Verification Failed: " + (err.response?.data || err.message));
        } finally {
            setActionLoading(false);
        }
    };

    if (loading || !asset) return <div className="p-10 text-center">Loading...</div>;

    const isOwner = asset.ownerId === userFullID;
    const isProposedRecipient = asset.proposedOwnerId === userFullID;
    const isPendingTransfer = asset.status === 'PENDING_TRANSFER';

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
                        {displayUrl ? (
                            <img
                                src={displayUrl}
                                className="w-full h-full object-cover grayscale-[20%] sepia-[10%] group-hover:grayscale-0 group-hover:sepia-0 transition-all duration-500"
                                alt={asset.name}
                            />
                        ) : (
                            <div className="w-full h-full flex items-center justify-center text-ink-900/20 italic text-xs">No image provided</div>
                        )}
                        <div className="absolute inset-0 bg-ink-900/0 group-hover:bg-ink-900/20 transition-all flex items-center justify-center">
                            <ExternalLink className="text-white opacity-0 group-hover:opacity-100 transition-all transform scale-50 group-hover:scale-100" />
                        </div>
                    </a>

                    {asset.imageHash && (
                        <div className="mt-4 px-3 py-2 bg-parchment-50 rounded border border-ink-900/5 flex items-center justify-between">
                            <div className="flex items-center gap-2 overflow-hidden">
                                <LinkIcon size={12} className="text-bronze shrink-0" />
                                <span className="text-[10px] font-mono text-ink-900/40 truncate">ipfs://{asset.imageHash}</span>
                            </div>
                            <a
                                href={`https://ipfs.io/ipfs/${asset.imageHash}`}
                                target="_blank"
                                rel="noopener noreferrer"
                                className="text-[10px] font-bold text-bronze hover:underline shrink-0"
                            >
                                PROVENANCE
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

                {/* Attachment Card */}
                {
                    asset.attachment && asset.attachment.file_name && (
                        <div className="bg-white p-6 rounded-xl border border-ink-900/10 shadow-sm space-y-4">
                            <div className="flex items-center gap-2">
                                <Paperclip className="text-bronze w-5 h-5" />
                                <h3 className="font-serif font-bold text-lg text-ink-900">Supporting Document</h3>
                            </div>

                            <div className="p-3 bg-parchment-50 rounded border border-ink-900/5 space-y-2">
                                <div className="flex items-center gap-2">
                                    <FileText size={14} className="text-ink-900/40" />
                                    <span className="text-xs font-bold text-ink-900 truncate">{asset.attachment.file_name}</span>
                                </div>
                                <div className="grid grid-cols-2 gap-2 text-[10px] text-ink-900/60 uppercase font-bold tracking-tighter">
                                    <div>Size: {(asset.attachment.file_size / 1024).toFixed(2)} KB</div>
                                    <div>Type: {asset.attachment.storage_type}</div>
                                </div>
                                <div className="pt-1">
                                    <div className="text-[8px] uppercase text-ink-900/30">IPFS CID</div>
                                    <div className="text-[10px] font-mono text-bronze truncate">{asset.attachment.ipfs_cid}</div>
                                </div>
                            </div>

                            <div className="flex gap-2 group/btns">
                                {attachmentUrl && (
                                    <a
                                        href={attachmentUrl}
                                        target="_blank"
                                        rel="noopener noreferrer"
                                        className="flex-1 flex justify-center items-center gap-2 bg-parchment-200 text-ink-900 py-3 rounded-lg hover:bg-parchment-300 transition-all font-bold text-sm border border-ink-900/5 shadow-sm"
                                        title="View in browser"
                                    >
                                        <Eye className="w-4 h-4" /> View
                                    </a>
                                )}
                                {attachmentDownloadUrl && (
                                    <a
                                        href={attachmentDownloadUrl}
                                        className="flex-1 flex justify-center items-center gap-2 bg-bronze text-white py-3 rounded-lg hover:bg-ink-900 transition-all font-bold text-sm shadow-md"
                                        title="Download to device"
                                    >
                                        <Download className="w-4 h-4" /> Download
                                    </a>
                                )}
                            </div>
                        </div>
                    )
                }
            </div>

            {/* Right Column: Details & History */}
            <div className="lg:col-span-2 space-y-8">
                <div>
                    <div className="flex justify-between items-start">
                        <div>
                            <div className="flex items-center gap-3 mb-2">
                                <h1 className="text-4xl font-serif font-bold text-ink-900 m-0">{asset.name}</h1>
                                <button
                                    onClick={handleVerifyBlockchain}
                                    title="Verify on Blockchain"
                                    className="p-1.5 rounded-full bg-parchment-200 text-bronze hover:bg-bronze hover:text-white transition-all shadow-sm"
                                >
                                    <Shield className="w-5 h-5" />
                                </button>
                            </div>
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

            {/* Blockchain Data Modal */}
            {
                showBlockchainModal && (
                    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-ink-900/60 backdrop-blur-sm">
                        <div className="bg-white rounded-2xl shadow-2xl w-full max-w-2xl max-h-[75vh] flex flex-col overflow-hidden animate-in fade-in zoom-in duration-200">
                            <div className="p-6 border-b border-ink-900/10 flex items-center justify-between bg-parchment-50">
                                <div className="flex items-center gap-3">
                                    <div className="p-2 bg-bronze/10 rounded-lg">
                                        <Verified className="text-bronze w-6 h-6" />
                                    </div>
                                    <div>
                                        <h2 className="font-serif font-bold text-xl text-ink-900">Blockchain Verification</h2>
                                        <p className="text-xs text-ink-900/50">Direct "Ground Truth" Read from Hyperledger Fabric</p>
                                    </div>
                                </div>
                                <button
                                    onClick={() => setShowBlockchainModal(false)}
                                    className="text-ink-900/40 hover:text-ink-900 text-2xl font-light"
                                >
                                    &times;
                                </button>
                            </div>

                            <div className="flex-1 overflow-y-auto p-5 font-mono text-xs">
                                <div className="mb-4 p-4 bg-emerald-50 border border-emerald-100 rounded-lg flex items-center gap-3">
                                    <CheckCircle className="text-emerald-500 shrink-0" />
                                    <div className="text-emerald-900">
                                        <p className="font-bold">Cryptographically Verified</p>
                                        <p className="text-[10px] opacity-70">The record below was fetched directly from the immutable ledger using your organization certificate.</p>
                                    </div>
                                </div>

                                <div className="space-y-4">
                                    <div className="p-4 bg-ink-900 rounded-xl text-ink-50 overflow-x-auto shadow-inner">
                                        <pre className="m-0">{JSON.stringify(blockchainData, null, 2)}</pre>
                                    </div>

                                    <div className="grid grid-cols-2 gap-4">
                                        <div className="p-3 bg-parchment-100 rounded border border-ink-900/5">
                                            <div className="text-[9px] uppercase tracking-tighter text-ink-900/40 mb-1">Chaincode ID</div>
                                            <div className="font-bold truncate text-bronze">basic:1.0</div>
                                        </div>
                                        <div className="p-3 bg-parchment-100 rounded border border-ink-900/5">
                                            <div className="text-[9px] uppercase tracking-tighter text-ink-900/40 mb-1">Fetch Time</div>
                                            <div className="font-bold truncate text-bronze">{new Date().toLocaleString()}</div>
                                        </div>
                                    </div>

                                    <div className="p-4 rounded-lg bg-parchment-100 border border-ink-900/10 grayscale-[30%]">
                                        <p className="font-serif text-[13px] text-ink-900/70 italic leading-relaxed text-center">
                                            "This ledger entry serves as the single source of truth for the provenance and ownership of {asset.name}."
                                        </p>
                                    </div>
                                </div>
                            </div>

                            <div className="p-4 bg-parchment-50 border-t border-ink-900/10 text-center">
                                <button
                                    onClick={() => setShowBlockchainModal(false)}
                                    className="px-8 py-2 bg-ink-900 text-white rounded-full hover:bg-ink-800 transition-colors font-bold text-sm"
                                >
                                    Re-seal Report
                                </button>
                            </div>
                        </div>
                    </div>
                )
            }
        </div>
    );
};

export default AssetDetails;
