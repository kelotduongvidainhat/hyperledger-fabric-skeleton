import { useState, useEffect } from 'react';
import { fetchAssets, fetchAssetsFromDB, createAsset, transferAsset, lockAsset, unlockAsset, setAuthToken, fetchIdentities, fetchAssetHistory } from '../services/api';
import { Button } from '@/components/button';
import { Input } from '@/components/input';
import { Card, CardHeader, CardTitle, CardContent } from '@/components/card';
import { RefreshCw, Send, Plus, History, X, Lock, Unlock, LogOut } from 'lucide-react';
import { useAuth } from '../context/AuthContext';

export default function AssetDashboard() {
    const { user, logout } = useAuth();
    const [dataSource, setDataSource] = useState('blockchain'); // 'blockchain' | 'database'
    const [assets, setAssets] = useState([]);
    const [loading, setLoading] = useState(false);
    const [formData, setFormData] = useState({
        id: '', name: '', category: '', owner: ''
    });
    const [transferData, setTransferData] = useState({ id: '', newOwner: '' });

    // History Modal State
    const [historyModalOpen, setHistoryModalOpen] = useState(false);
    const [selectedAssetHistory, setSelectedAssetHistory] = useState([]);
    const [selectedAssetId, setSelectedAssetId] = useState('');
    const [loadingHistory, setLoadingHistory] = useState(false);

    useEffect(() => {
        // Initial load
        loadAssets();
    }, [dataSource]); // Removed currentUser dependency

    const loadAssets = async () => {
        setLoading(true);
        try {
            let data;
            if (dataSource === 'database') {
                data = await fetchAssetsFromDB();
            } else {
                data = await fetchAssets();
            }
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
            await createAsset(formData);
            setFormData({ id: '', name: '', category: '', owner: '' });
            loadAssets();
        } catch (error) {
            alert('Failed to create asset: ' + error.response?.data?.error || error.message);
        }
    };

    const handleLock = async (id) => {
        try {
            await lockAsset(id);
            loadAssets();
        } catch (error) {
            alert('Failed to lock asset: ' + (error.response?.data?.error || error.message));
        }
    };

    const handleUnlock = async (id) => {
        try {
            await unlockAsset(id);
            loadAssets();
        } catch (error) {
            alert('Failed to unlock asset: ' + (error.response?.data?.error || error.message));
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

    const handleShowHistory = async (assetId) => {
        setSelectedAssetId(assetId);
        setHistoryModalOpen(true);
        setLoadingHistory(true);
        try {
            const history = await fetchAssetHistory(assetId);
            setSelectedAssetHistory(history);
        } catch (error) {
            console.error("Failed to fetch history", error);
            alert("Failed to fetch history: " + (error.response?.data?.error || error.message));
        } finally {
            setLoadingHistory(false);
        }
    };

    const formatTimestamp = (ts) => {
        if (!ts) return 'N/A';
        // Handle Go default time string: "2025-12-30 03:41:51.76... +0000 UTC"
        // Try to convert to ISO: Replace " +0000 UTC" with "Z" and space with "T"
        try {
            const iso = ts.replace(' +0000 UTC', 'Z').replace(' ', 'T');
            const d = new Date(iso);
            if (!isNaN(d.getTime())) return d.toLocaleString();
            return ts;
        } catch (e) {
            return ts;
        }
    };

    return (
        <div className="container mx-auto p-4 space-y-8">
            <header className="flex justify-between items-center mb-8">
                <h1 className="text-3xl font-bold">Hyperledger Fabric Asset Manager</h1>
                <div className="flex items-center gap-4">
                    <div className="flex items-center gap-2 mr-4">
                        <span className="text-sm font-medium">Source:</span>
                        <div className="flex bg-gray-100 p-1 rounded-lg">
                            <button
                                onClick={() => setDataSource('blockchain')}
                                className={`px-3 py-1 rounded-md text-sm font-medium transition-colors ${dataSource === 'blockchain' ? 'bg-white shadow text-green-600' : 'text-gray-500 hover:text-gray-900'}`}
                            >
                                Blockchain
                            </button>
                            <button
                                onClick={() => setDataSource('database')}
                                className={`px-3 py-1 rounded-md text-sm font-medium transition-colors ${dataSource === 'database' ? 'bg-white shadow text-blue-600' : 'text-gray-500 hover:text-gray-900'}`}
                            >
                                Database
                            </button>
                        </div>
                    </div>
                    <div className="flex items-center gap-4 text-sm">
                        <span className="text-gray-600">
                            User: <span className="font-bold text-gray-900">{user?.username || 'Unknown'}</span>
                        </span>
                        <Button variant="outline" size="sm" onClick={logout} className="flex items-center gap-1">
                            <LogOut className="w-4 h-4" /> Logout
                        </Button>
                    </div>
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
                                <Input placeholder="Name" value={formData.name} onChange={e => setFormData({ ...formData, name: e.target.value })} required />
                                <Input placeholder="Category" value={formData.category} onChange={e => setFormData({ ...formData, category: e.target.value })} required />
                            </div>
                            <Input placeholder="Owner" value={formData.owner} onChange={e => setFormData({ ...formData, owner: e.target.value })} required />
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
                                    <th scope="col" className="px-6 py-3">Name</th>
                                    <th scope="col" className="px-6 py-3">Category</th>
                                    <th scope="col" className="px-6 py-3">Owner</th>
                                    <th scope="col" className="px-6 py-3">Status</th>
                                    <th scope="col" className="px-6 py-3">Actions</th>
                                </tr>
                            </thead>
                            <tbody>
                                {assets.map((asset) => (
                                    <tr key={asset.ID} className="bg-white border-b">
                                        <td className="px-6 py-4 font-medium text-gray-900 whitespace-nowrap">{asset.ID}</td>
                                        <td className="px-6 py-4">{asset.Name}</td>
                                        <td className="px-6 py-4">{asset.Category}</td>
                                        <td className="px-6 py-4">{asset.Owner}</td>
                                        <td className="px-6 py-4">
                                            <span className={`px-2 py-1 rounded text-xs font-bold ${asset.Status === 'AVAILABLE' ? 'bg-green-100 text-green-800' : 'bg-red-100 text-red-800'}`}>
                                                {asset.Status}
                                            </span>
                                        </td>
                                        <td className="px-6 py-4 flex items-center gap-2">
                                            <Button variant="ghost" size="sm" onClick={() => handleShowHistory(asset.ID)} title="History">
                                                <History className="w-4 h-4" />
                                            </Button>
                                            {asset.Status === 'AVAILABLE' ? (
                                                <Button variant="ghost" size="sm" onClick={() => handleLock(asset.ID)} title="Lock" className="text-yellow-600 hover:text-yellow-800 hover:bg-yellow-50">
                                                    <Lock className="w-4 h-4" />
                                                </Button>
                                            ) : (
                                                <Button variant="ghost" size="sm" onClick={() => handleUnlock(asset.ID)} title="Unlock" className="text-green-600 hover:text-green-800 hover:bg-green-50">
                                                    <Unlock className="w-4 h-4" />
                                                </Button>
                                            )}
                                        </td>
                                    </tr>
                                ))}
                                {assets.length === 0 && (
                                    <tr><td colSpan="6" className="text-center py-4">No assets found</td></tr>
                                )}
                            </tbody>
                        </table>
                    </div>
                </CardContent>
            </Card>

            {/* History Modal */}
            {historyModalOpen && (
                <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center p-4 z-50">
                    <div className="bg-white rounded-lg shadow-xl w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col">
                        <div className="p-4 border-b flex justify-between items-center bg-gray-50">
                            <h2 className="text-xl font-bold flex items-center gap-2">
                                <History className="w-5 h-5" /> Asset History: {selectedAssetId}
                            </h2>
                            <Button variant="ghost" size="icon" onClick={() => setHistoryModalOpen(false)}>
                                <X className="w-5 h-5" />
                            </Button>
                        </div>
                        <div className="p-4 overflow-y-auto flex-1">
                            {loadingHistory ? (
                                <div className="flex justify-center p-8"><RefreshCw className="w-8 h-8 animate-spin text-blue-500" /></div>
                            ) : (
                                <div className="space-y-4">
                                    {selectedAssetHistory.length === 0 ? (
                                        <p className="text-center text-gray-500">No history found.</p>
                                    ) : (

                                        selectedAssetHistory.map((record, index) => (
                                            <div key={index} className="border rounded-lg p-3 hover:bg-gray-50 transition-colors">
                                                <div className="flex justify-between items-start mb-2">
                                                    <span className="text-xs font-mono bg-gray-200 px-2 py-1 rounded text-gray-700 truncate max-w-[200px]" title={record.txId}>
                                                        TX: {record.txId ? record.txId.substring(0, 10) : '???'}...
                                                    </span>
                                                    <span className="text-xs text-gray-500">{formatTimestamp(record.timestamp)}</span>
                                                </div>
                                                {record.record ? (
                                                    <div className="grid grid-cols-2 gap-2 text-sm">
                                                        <div><span className="font-semibold">Owner:</span> {record.record.Owner}</div>
                                                        <div><span className="font-semibold">Status:</span> {record.record.Status}</div>
                                                    </div>
                                                ) : (
                                                    <div className="text-sm text-gray-500 italic">No record data</div>
                                                )}
                                                {record.isDelete && <div className="mt-2 text-red-500 text-xs font-bold">ASSET DELETED</div>}
                                            </div>
                                        ))
                                    )}
                                </div>
                            )}
                        </div>
                    </div>
                </div>
            )}
        </div>
    );
}
