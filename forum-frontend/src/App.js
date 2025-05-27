import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import Login from './components/auth/Login';
import Register from './components/auth/Register';
import PostList from './components/posts/PostList';
import CreatePost from './components/posts/CreatePost';
import MainLayout from './components/layout/MainLayout';
import Chat from './components/chat/Chat';
import Home from './components/Home';
import AdminPanel from './components/admin/AdminPanel';
import './components/MainLayout.css';

const PrivateRoute = ({ children, requireAdmin = false }) => {
  const isAuthenticated = !!localStorage.getItem('token');
  const userRole = localStorage.getItem('userRole');

  if (!isAuthenticated) {
    return <Navigate to="/login" />;
  }

  if (requireAdmin && userRole !== 'admin') {
    return <Navigate to="/" />;
  }

  return children;
};

const App = () => {
    const [refreshPostsTrigger, setRefreshPostsTrigger] = useState(0);
    const isAuthenticated = !!localStorage.getItem('token');

    const handlePostCreated = () => {
        setRefreshPostsTrigger(prev => prev + 1);
    };

  return (
        <Router future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
            <MainLayout>
          <Routes>
                    <Route path="/register" element={<Register />} />
            <Route path="/login" element={<Login />} />
                    <Route path="/posts" element={
                        <>
                            {isAuthenticated && <CreatePost onCreate={handlePostCreated} />}
                            <PostList refreshTrigger={refreshPostsTrigger} />
                        </>
                    } />
                    <Route path="/chat" element={<Chat />} />
                    <Route path="/admin" element={
                      <PrivateRoute requireAdmin={true}>
                        <AdminPanel />
                      </PrivateRoute>
                    } />
          </Routes>
            </MainLayout>
      </Router>
  );
};

export default App; 