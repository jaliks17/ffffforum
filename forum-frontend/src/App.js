import React, { useState } from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import Login from './components/auth/Login';
import Register from './components/auth/Register';
import PostList from './components/posts/PostList';
import PostDetail from './components/posts/PostDetail';
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
    const [refreshPosts, setRefreshPosts] = useState(false);
    const isAuthenticated = !!localStorage.getItem('token');

    const onPostCreated = () => {
        setRefreshPosts(prev => !prev);
    };

  return (
        <Router future={{ v7_startTransition: true, v7_relativeSplatPath: true }}>
            <MainLayout>
          <Routes>
            <Route path="/" element={<Home />} />
                    <Route path="/register" element={<Register />} />
            <Route path="/login" element={<Login />} />
                    <Route path="/posts" element={
                        <>
                            {isAuthenticated && <CreatePost onPostCreated={onPostCreated} />}
                            <PostList key={refreshPosts} />
                        </>
                    } />
                    <Route path="/posts/:id" element={<PostDetail />} />
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