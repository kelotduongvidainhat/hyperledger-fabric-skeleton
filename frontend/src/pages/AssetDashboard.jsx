import React, { useState, useEffect } from 'react';
import { fetchAssets, createAsset, transferAsset, setAuthToken } from '../services/api';
import { Button } from '@/components/button';
import { Input } from '@/components/input';
import { Card, CardHeader, CardTitle, CardContent, CardFooter } from '@/components/card';
import { RefreshCw, Send, Plus } from 'lucide-react';

export default function AssetDashboard() {
    const [currentUser, setCurrentUser] = useState('admin');
    const [assets, setAssets] = useState([]);
    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState({
        id: '', color: '', size: '', owner: '', appraisedValue: ''
    });
    const [transferData, setTransferData] = useState({ id: '', newOwner: '' });

    useEffect(() => {
        setAuthToken(currentUser);
        loadAssets();
    }, [currentUser]);

    const loadAssets = async () => {
        setLoading(true);
        try {
            const data = await fetchAssets();
            setAssets(data);
        } catch (error) {
            console.error("Failed to fetch assets", error);
        } finally {
            setLoading(false);
        }
    };

    const handleCreate = async (e) => {
        e.preventDefault();
        try {
            await createAsset({
                ...formData,
                size: parseInt(formData.size),
                appraisedValue: parseInt(formData.appraisedValue)
            });
            setFormData({ id: '', color: '', size: '', owner: '', appraisedValue: '' });
            loadAssets();
        } catch (error) {
            alert('Failed to create asset: ' + error.response?.data?.error || error.message);
        }
    };

    const handleTransfer = async (e) => {
        e.preventDefault();
        try {
            await transferAsset(transferData.id, transferData.newOwner);
            setTransferData({ id: '', newOwner: '' });
            loadAssets();
        } catch (error) {
            alert('Failed to transfer asset: ' + error.response?.data?.error || error.message);
        }
    };

    return (
        <div className="container mx-auto p-4 space-y-8">
            <header className="flex justify-between items-center mb-8">
                <h1 className="text-3xl font-bold">Hyperledger Fabric Asset Manager</h1>
                <div className="flex items-center gap-4">
                    <span className="text-sm text-gray-500">Acting as:</span>
                    <select
                        className="border rounded p-2"
                        value={currentUser}
                        onChange={(e) => setCurrentUser(e.target.value)}
                    >
                        <option value="admin">Admin (Org1)</option>
                        <option value="user1">User1 (Org1)</option>
                    </select>
                </div>
            </header>

            <div className="grid md:grid-cols-2 gap-8">
                {/* Create Asset Form */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2"><Plus className="w-5 h-5" /> Create Asset</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleCreate} className="space-y-4">
                            <Input placeholder="Asset ID" value={formData.id} onChange={e => setFormData({ ...formData, id: e.target.value })} required />
                            <div className="grid grid-cols-2 gap-4">
                                <Input placeholder="Color" value={formData.color} onChange={e => setFormData({ ...formData, color: e.target.value })} required />
                                <Input type="number" placeholder="Size" value={formData.size} onChange={e => setFormData({ ...formData, size: e.target.value })} required />
                            </div>
                            <Input placeholder="Owner" value={formData.owner} onChange={e => setFormData({ ...formData, owner: e.target.value })} required />
                            <Input type="number" placeholder="Value" value={formData.appraisedValue} onChange={e => setFormData({ ...formData, appraisedValue: e.target.value })} required />
                            <Button type="submit" className="w-full">Create Asset</Button>
                        </form>
                    </CardContent>
                </Card>

                {/* Transfer Asset Form */}
                <Card>
                    <CardHeader>
                        <CardTitle className="flex items-center gap-2"><Send className="w-5 h-5" /> Transfer Asset</CardTitle>
                    </CardHeader>
                    <CardContent>
                        <form onSubmit={handleTransfer} className="space-y-4">
                            <Input placeholder="Asset ID to Transfer" value={transferData.id} onChange={e => setTransferData({ ...transferData, id: e.target.value })} required />
                            <Input placeholder="New Owner" value={transferData.newOwner} onChange={e => setTransferData({ ...transferData, newOwner: e.target.value })} required />
                            <Button type="submit" variant="secondary" className="w-full">Transfer Asset</Button>
                        </form>
                    </CardContent>
                </Card>
            </div>

            {/* Assets List */}
            <Card>
                <CardHeader className="flex flex-row items-center justify-between">
                    <CardTitle>Asset Inventory</CardTitle>
                    <Button variant="ghost" size="icon" onClick={loadAssets}><RefreshCw className={`w-5 h-5 ${loading ? 'animate-spin' : ''}`} /></Button>
                </CardHeader>
                <CardContent>
                    <div className="relative overflow-x-auto">
                        <table className="w-full text-sm text-left rtl:text-right text-gray-500">
                            <thead className="text-xs text-gray-700 uppercase bg-gray-50">
                                <tr>
                                    <th scope="col" className="px-6 py-3">ID</th>
                                    <th scope="col" className="px-6 py-3">Color</th>
                                    <th scope="col" className="px-6 py-3">Size</th>
                                    <th scope="col" className="px-6 py-3">Owner</th>
                                    <th scope="col" className="px-6 py-3">Value</th>
                                </tr>
                            </thead>
                            <tbody>
                                {assets.map((asset) => (
                                    <tr key={asset.ID} className="bg-white border-b">
                                        <td className="px-6 py-4 font-medium text-gray-900 whitespace-nowrap">{asset.ID}</td>
                                        <td className="px-6 py-4">{asset.Color}</td>
                                        <td className="px-6 py-4">{asset.Size}</td>
                                        <td className="px-6 py-4">{asset.Owner}</td>
                                        <td className="px-6 py-4">${asset.AppraisedValue}</td>
                                    </tr>
                                ))}
                                {assets.length === 0 && (
                                    <tr><td colSpan="5" className="text-center py-4">No assets found</td></tr>
                                )}
                            </tbody>
                        </table>
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
