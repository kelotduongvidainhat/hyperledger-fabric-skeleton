import React from 'react';
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { AuthProvider, useAuth } from './context/AuthContext';
import Layout from './components/Layout';
import Dashboard from './pages/Dashboard';
import Gallery from './pages/Gallery';
import GalleryAssetDetails from './pages/GalleryAssetDetails';
import CreateAsset from './pages/CreateAsset';
import AssetDetails from './pages/AssetDetails';
import Login from './pages/Login';
import Register from './pages/Register';
import AdminDashboard from './pages/AdminDashboard';
import AdminUsers from './pages/AdminUsers';
import AdminAssets from './pages/AdminAssets';

const ProtectedRoute = ({ children }) => {
  const { isAuthenticated, loading } = useAuth();
  const location = useLocation();

  if (loading) {
    return <div className="p-10 text-center">Loading registry...</div>;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" state={{ from: location }} replace />;
  }

  return children;
};

const AdminRoute = ({ children }) => {
  const { role, loading, isAuthenticated } = useAuth();

  if (loading) return <div className="p-10 text-center">Loading board...</div>;

  if (!isAuthenticated || role !== 'admin') {
    return <Navigate to="/" replace />;
  }

  return children;
};

function App() {
  return (
    <AuthProvider>
      <BrowserRouter>
        <Routes>
          <Route path="/login" element={<Login />} />
          <Route path="/register" element={<Register />} />

          <Route path="/" element={
            <ProtectedRoute>
              <Layout>
                <Dashboard />
              </Layout>
            </ProtectedRoute>
          } />

          <Route path="/gallery" element={
            <ProtectedRoute>
              <Layout>
                <Gallery />
              </Layout>
            </ProtectedRoute>
          } />

          <Route path="/gallery/:id" element={
            <ProtectedRoute>
              <Layout>
                <GalleryAssetDetails />
              </Layout>
            </ProtectedRoute>
          } />

          <Route path="/create" element={
            <ProtectedRoute>
              <Layout>
                <CreateAsset />
              </Layout>
            </ProtectedRoute>
          } />

          <Route path="/assets/:id" element={
            <ProtectedRoute>
              <Layout>
                <AssetDetails />
              </Layout>
            </ProtectedRoute>
          } />

          <Route path="/admin" element={
            <AdminRoute>
              <Layout>
                <AdminDashboard />
              </Layout>
            </AdminRoute>
          } />

          <Route path="/admin/users" element={
            <AdminRoute>
              <Layout>
                <AdminUsers />
              </Layout>
            </AdminRoute>
          } />

          <Route path="/admin/assets" element={
            <AdminRoute>
              <Layout>
                <AdminAssets />
              </Layout>
            </AdminRoute>
          } />
        </Routes>
      </BrowserRouter>
    </AuthProvider>
  );
}

export default App;
