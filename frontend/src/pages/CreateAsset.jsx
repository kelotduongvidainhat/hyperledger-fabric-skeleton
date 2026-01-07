import React, { useState } from 'react';
import { createAsset, uploadToIPFS } from '../api/client';
import { useNavigate } from 'react-router-dom';
import { Upload, X, CheckCircle, Loader2, Image as ImageIcon } from 'lucide-react';

const CreateAsset = () => {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false);
    const [uploading, setUploading] = useState(false);
    const [form, setForm] = useState({
        id: `artifact_${Date.now()}`,
        name: '',
        desc: '',
        image_url: '',
        image_hash: '',
        view: 'Public'
    });

    const handleFileChange = async (e) => {
        const file = e.target.files[0];
        if (!file) return;

        setUploading(true);
        try {
            const result = await uploadToIPFS(file);
            // Result contains { cid, url, message }
            setForm(prev => ({
                ...prev,
                image_url: result.url, // ipfs://CID
                image_hash: result.cid
            }));
        } catch (err) {
            console.error("Upload failed", err);
            alert("IPFS upload failed: " + (err.response?.data?.error || err.message));
        } finally {
            setUploading(false);
        }
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        if (!form.image_url) {
            alert("Please upload an artifact image to IPFS first.");
            return;
        }
        setLoading(true);
        try {
            await createAsset(form);
            navigate('/');
        } catch (err) {
            alert("Failed to commit artifact: " + err.message);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="max-w-2xl mx-auto py-4">
            <div className="flex justify-between items-center mb-8 border-b border-ink-900/10 pb-4">
                <div>
                    <h2 className="text-3xl font-serif text-ink-900 italic">Mint New Artifact</h2>
                    <p className="text-[10px] uppercase tracking-[0.2em] text-ink-900/40 mt-1 font-bold">Anchoring Metadata to Blockchain & IPFS</p>
                </div>
                <button onClick={() => navigate('/')} className="p-2 text-ink-900/40 hover:text-wax-red transition-colors">
                    <X className="w-6 h-6" />
                </button>
            </div>

            <form onSubmit={handleSubmit} className="space-y-8">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-8">
                    <div className="space-y-6">
                        {/* ID (Readonly or Generated) */}
                        <div>
                            <label className="block text-[10px] uppercase font-bold text-ink-900/40 mb-2 tracking-widest">Registry Identifier</label>
                            <input
                                type="text"
                                value={form.id}
                                onChange={e => setForm({ ...form, id: e.target.value })}
                                className="w-full p-2.5 bg-parchment-100 border border-ink-900/10 rounded font-mono text-xs text-ink-900/60 outline-none"
                            />
                        </div>

                        {/* Name */}
                        <div>
                            <label className="block text-[10px] uppercase font-bold text-ink-900/40 mb-2 tracking-widest">Artifact Name</label>
                            <input
                                type="text"
                                required
                                value={form.name}
                                onChange={e => setForm({ ...form, name: e.target.value })}
                                className="w-full p-2.5 bg-white border border-ink-900/10 rounded focus:border-bronze outline-none transition-all shadow-sm"
                                placeholder="e.g. Royal Decree of 1852"
                            />
                        </div>

                        {/* Description */}
                        <div>
                            <label className="block text-[10px] uppercase font-bold text-ink-900/40 mb-2 tracking-widest">Provenance Description</label>
                            <textarea
                                rows="4"
                                value={form.desc}
                                onChange={e => setForm({ ...form, desc: e.target.value })}
                                className="w-full p-2.5 bg-white border border-ink-900/10 rounded focus:border-bronze outline-none transition-all shadow-sm resize-none"
                                placeholder="Historical significance and origin..."
                            />
                        </div>
                    </div>

                    <div className="space-y-6">
                        {/* Image / IPFS Section */}
                        <div>
                            <label className="block text-[10px] uppercase font-bold text-ink-900/40 mb-2 tracking-widest">Decentralized Storage (IPFS)</label>

                            <div className={`relative h-56 rounded-xl border-2 border-dashed transition-all flex flex-col items-center justify-center overflow-hidden
                                ${form.image_url ? 'border-bronze bg-parchment-50' : 'border-ink-900/10 bg-white hover:border-bronze/50'}`}>

                                {form.image_url ? (
                                    <>
                                        <img
                                            src={form.image_url.replace('ipfs://', 'https://ipfs.io/ipfs/')}
                                            alt="Preview"
                                            className="absolute inset-0 w-full h-full object-cover opacity-90"
                                        />
                                        <div className="absolute inset-0 bg-ink-900/20 group hover:bg-ink-900/40 transition-all flex items-center justify-center">
                                            <button
                                                type="button"
                                                onClick={() => setForm({ ...form, image_url: '', image_hash: '' })}
                                                className="bg-white/90 text-wax-red p-2 rounded-full opacity-0 group-hover:opacity-100 transition-all shadow-xl"
                                            >
                                                <X className="w-5 h-5" />
                                            </button>
                                        </div>
                                    </>
                                ) : (
                                    <div className="p-6 text-center">
                                        {uploading ? (
                                            <div className="flex flex-col items-center gap-3">
                                                <Loader2 className="w-10 h-10 text-bronze animate-spin" />
                                                <span className="text-xs font-bold text-ink-900/60 uppercase">Uploading to IPFS...</span>
                                            </div>
                                        ) : (
                                            <>
                                                <Upload className="w-10 h-10 text-ink-900/20 mb-3 mx-auto" />
                                                <p className="text-xs text-ink-900/40 mb-4 font-serif italic">Drag and drop file or click to select</p>
                                                <label className="cursor-pointer px-4 py-2 bg-bronze text-white text-[10px] font-bold uppercase rounded tracking-widest hover:bg-ink-800 transition-colors shadow-sm">
                                                    Select Artifact
                                                    <input type="file" className="hidden" onChange={handleFileChange} accept="image/*" />
                                                </label>
                                            </>
                                        )}
                                    </div>
                                )}
                            </div>

                            {form.image_hash && (
                                <div className="mt-3 p-3 bg-parchment-100 rounded border border-ink-900/5 flex items-center gap-3">
                                    <CheckCircle size={14} className="text-green-600" />
                                    <div className="flex-1 min-w-0">
                                        <div className="text-[8px] uppercase font-bold text-ink-900/40 tracking-wider">Content ID (CID)</div>
                                        <div className="text-[10px] font-mono text-bronze truncate">{form.image_hash}</div>
                                    </div>
                                </div>
                            )}
                        </div>

                        <div className="p-4 bg-blue-50/50 rounded-lg border border-blue-900/5 flex gap-3">
                            <ImageIcon className="text-blue-900/40 shrink-0" size={16} />
                            <p className="text-[10px] text-blue-900/60 leading-relaxed font-serif italic">
                                Images uploaded here are permanently stored on the decentralized IPFS network.
                                The CID will be recorded on the Hyperledger Fabric ledger to prove authenticity.
                            </p>
                        </div>
                    </div>
                </div>

                <div className="pt-6 border-t border-ink-900/10">
                    <button
                        type="submit"
                        disabled={loading || uploading || !form.image_url}
                        className="w-full flex justify-center items-center gap-2 bg-ink-900 text-white py-4 rounded-xl hover:bg-bronze transition-all font-serif text-xl font-bold shadow-xl disabled:opacity-50 disabled:grayscale"
                    >
                        {loading ? 'Minting on Blockchain...' : 'Commit Artifact to Ledger'}
                    </button>
                    <p className="text-center text-[10px] text-ink-900/30 mt-4 uppercase tracking-[0.3em] font-bold">Endorsement Required by Org1 & Org2</p>
                </div>
            </form>
        </div>
    );
};

export default CreateAsset;
