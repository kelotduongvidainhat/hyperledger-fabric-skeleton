import React, { useEffect, useState } from 'react';
import { useParams, useNavigate, useLocation, Link } from 'react-router-dom';
import { fetchAssets, fetchHistory, proposeTransfer, acceptTransfer } from '../api/client';
import { ArrowLeft, ArrowRight, CheckCircle, Shield, History } from 'lucide-react';
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
            const [a, h] = await Promise.all([fetchAssets(id), fetchHistory(id)]);
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

    if (loading || !asset) return <div className="p-10 text-center">Loading...</div>;

    const isOwner = asset.OwnerID === userFullID;
    const isProposedRecipient = asset.ProposedOwnerID === userFullID;
    const isPendingTransfer = asset.Status === 'PENDING_TRANSFER';

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

                <div className="bg-white p-2 border border-ink-900/10 rounded-xl shadow-sm">
                    <div className="aspect-square bg-parchment-200 rounded-lg overflow-hidden">
                        <img src={asset.ImageURL} className="w-full h-full object-cover grayscale-[20%] sepia-[10%]" />
                    </div>
                </div>

                {/* Actions Panel */}
                <div className="bg-white p-6 rounded-xl border border-ink-900/10 shadow-sm">
                    <h3 className="font-serif font-bold text-lg mb-4 text-ink-900">Ownership Actions</h3>

                    {isOwner && !isPendingTransfer && asset.Status === 'ACTIVE' && (
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
                                    `Proposed Recipient: ${asset.ProposedOwnerID}`
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
                </div>
            </div>

            {/* Right Column: Details & History */}
            <div className="lg:col-span-2 space-y-8">
                <div>
                    <div className="flex justify-between items-start">
                        <div>
                            <h1 className="text-4xl font-serif font-bold text-ink-900 mb-2">{asset.Name}</h1>
                            <span className="font-mono text-sm bg-parchment-200 px-2 py-1 rounded text-ink-900/60">{asset.ID}</span>
                        </div>
                        <div className="text-right">
                            <div className="text-xs text-ink-900/40 uppercase tracking-widest mb-1">Current Owner</div>
                            <div className="flex items-center gap-2 font-bold text-lg text-ink-900">
                                <Shield className="w-5 h-5 text-wax-red" />
                                {asset.OwnerID}
                            </div>
                        </div>
                    </div>
                    <p className="mt-6 text-lg text-ink-900/80 leading-relaxed font-serif">
                        {asset.Description}
                    </p>
                </div>

                <div className="border-t border-ink-900/10 pt-8">
                    <h3 className="flex items-center gap-2 font-serif font-bold text-xl text-ink-900 mb-6">
                        <History className="w-5 h-5" /> Provenance History
                    </h3>

                    <div className="space-y-0 relative border-l-2 border-parchment-300 ml-3">
                        {history.map((record, idx) => (
                            <div key={idx} className="relative pl-8 pb-8 last:pb-0">
                                <div className="absolute -left-[9px] top-0 w-4 h-4 bg-parchment-100 border-2 border-bronze rounded-full"></div>
                                <div className="bg-white p-4 rounded-lg border border-ink-900/5 shadow-sm">
                                    <div className="flex justify-between items-start mb-1">
                                        <span className="font-bold text-ink-900 text-sm">{record.ActionType.replace(/_/g, ' ')}</span>
                                        <span className="text-xs text-ink-900/40 text-right">
                                            {new Date(record.Timestamp).toLocaleString()}
                                        </span>
                                    </div>
                                    <div className="text-xs text-ink-900/60">
                                        Actor: <span className="font-mono text-wax-red">{record.ActorID}</span>
                                    </div>
                                    <div className="text-[10px] font-mono text-ink-900/30 mt-2 truncate">
                                        TX: {record.TxId}
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
