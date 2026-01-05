import React, { useState } from 'react';
import { createAsset } from '../api/client';
import { useNavigate } from 'react-router-dom';
import { Upload, X } from 'lucide-react';

const CreateAsset = () => {
    const navigate = useNavigate();
    const [loading, setLoading] = useState(false);
    const [form, setForm] = useState({
        id: `asset_${Date.now()}`,
        name: '',
        desc: '',
        image_url: 'https://images.unsplash.com/photo-1618005182384-a83a8bd57fbe?auto=format&fit=crop&q=80',
        image_hash: 'QmHashPlaceholder',
        view: 'Public'
    });

    const handleSubmit = async (e) => {
        e.preventDefault();
        setLoading(true);
        try {
            await createAsset(form);
            navigate('/');
        } catch (err) {
            alert("Failed to create asset: " + err.message);
        } finally {
            setLoading(false);
        }
    };

    return (
        <div className="max-w-2xl mx-auto">
            <div className="flex justify-between items-center mb-6">
                <h2 className="text-2xl font-serif text-ink-900">Mint New Artifact</h2>
                <button onClick={() => navigate('/')} className="p-2 hover:bg-ink-900/5 rounded-full">
                    <X className="w-5 h-5" />
                </button>
            </div>

            <form onSubmit={handleSubmit} className="bg-white p-8 rounded-xl shadow-sm border border-ink-900/10 space-y-6">

                {/* ID (Readonly or Generated) */}
                <div>
                    <label className="block text-sm font-bold text-ink-900 mb-1">Asset ID</label>
                    <input
                        type="text"
                        value={form.id}
                        onChange={e => setForm({ ...form, id: e.target.value })}
                        className="w-full p-2 bg-parchment-50 border border-ink-900/20 rounded font-mono text-sm text-ink-900/60"
                    />
                </div>

                {/* Name */}
                <div>
                    <label className="block text-sm font-bold text-ink-900 mb-1">Name</label>
                    <input
                        type="text"
                        required
                        value={form.name}
                        onChange={e => setForm({ ...form, name: e.target.value })}
                        className="w-full p-2 bg-white border border-ink-900/20 rounded focus:ring-1 focus:ring-bronze focus:border-bronze outline-none"
                        placeholder="e.g. The Crown Jewels"
                    />
                </div>

                {/* Description */}
                <div>
                    <label className="block text-sm font-bold text-ink-900 mb-1">Description</label>
                    <textarea
                        rows="3"
                        value={form.desc}
                        onChange={e => setForm({ ...form, desc: e.target.value })}
                        className="w-full p-2 bg-white border border-ink-900/20 rounded focus:ring-1 focus:ring-bronze focus:border-bronze outline-none"
                        placeholder="Provenance details..."
                    />
                </div>

                {/* Image Preview */}
                <div>
                    <label className="block text-sm font-bold text-ink-900 mb-1">Image URL</label>
                    <div className="flex gap-4">
                        <input
                            type="url"
                            value={form.image_url}
                            onChange={e => setForm({ ...form, image_url: e.target.value })}
                            className="flex-1 p-2 bg-white border border-ink-900/20 rounded focus:ring-1 focus:ring-bronze outline-none"
                        />
                    </div>
                    {form.image_url && (
                        <div className="mt-2 h-32 w-full bg-parchment-100 rounded overflow-hidden">
                            <img src={form.image_url} alt="Preview" className="h-full w-full object-cover opacity-80" />
                        </div>
                    )}
                </div>

                <button
                    type="submit"
                    disabled={loading}
                    className="w-full flex justify-center items-center gap-2 bg-ink-900 text-white py-3 rounded hover:bg-ink-800 transition-colors font-serif shadow-md disabled:opacity-50"
                >
                    {loading ? 'Minting on Blockchain...' : 'Commit to Ledger'}
                </button>
            </form>
        </div>
    );
};

export default CreateAsset;
