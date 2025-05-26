import React, { useEffect, useState, useCallback } from 'react';
import axios from 'axios';
import Comments from './Comments'; 
import '/Users/Лика/Desktop/ffffffffor/forum-frontend/src/components/MainLayout.css';
import { useNavigate } from 'react-router-dom';

const PostList = ({ refreshTrigger }) => {
    const [posts, setPosts] = useState([]);
    const [loading, setLoading] = useState(true);
    const [error, setError] = useState(null);
    const [editingPostId, setEditingPostId] = useState(null);
    const [editFormData, setEditFormData] = useState({
        title: '',
        content: ''
    });

    const currentUserId = localStorage.getItem('userId') ? parseInt(localStorage.getItem('userId'), 10) : null;
    const currentUserRole = localStorage.getItem('userRole');
    const token = localStorage.getItem('token');
    const isAuthenticated = !!token;
    const navigate = useNavigate();

    const fetchPosts = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);
            setEditingPostId(null);

            const response = await axios.get('http://localhost:8080/api/v1/posts', {
                headers: { 
                    'Accept': 'application/json',
                    ...(token && { 'Authorization': `Bearer ${token}` })
                }
            });

            const rawPosts = response.data.data || response.data;

            if (!Array.isArray(rawPosts)) {
                console.error("Received data is not an array:", rawPosts);
                throw new Error('Invalid posts data format');
            }

            const processedPosts = rawPosts.map(post => ({
                ...post,
                id: parseInt(post.id, 10),
                author_id: parseInt(post.author_id, 10),
                created_at: post.created_at || new Date().toISOString()
            }));

            processedPosts.sort((a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime());

            setPosts(processedPosts);
        } catch (err) {
            console.error("Fetch posts error:", err.response?.data || err.message);
            setError(err.response?.data?.error || err.response?.data?.message || err.message || 'Failed to load posts');
        } finally {
            setLoading(false);
        }
    }, [token, refreshTrigger]);

    useEffect(() => {
        fetchPosts();
    }, [fetchPosts, refreshTrigger]);

    const handleDeletePost = async (postId, authorId) => {
        if (!isAuthenticated) {
            navigate('/login');
            return;
        }

        const confirmDelete = window.confirm("Вы уверены, что хотите удалить этот пост?");
        if (!confirmDelete) return;

        try {
            await axios.delete(`http://localhost:8080/api/v1/posts/${postId}`, {
                headers: { 
                    'Authorization': `Bearer ${token}`,
                    'Content-Type': 'application/json'
                }
            });

            setPosts(prev => prev.filter(post => post.id !== postId));
        } catch (error) {
            console.error('Delete post error:', error);
            const errorMessage = error.response?.data?.error || 
                               error.response?.data?.message || 
                               'Не удалось удалить пост';
            
            alert(errorMessage);
        }
    };

    const startEditing = (post) => {
        if (!isAuthenticated) {
            navigate('/login');
            return;
        }
        if (currentUserId !== post.author_id && currentUserRole !== 'admin') {
             alert('Вы не можете редактировать этот пост.');
             return;
        }
        setEditingPostId(post.id);
        setEditFormData({
            title: post.title,
            content: post.content
        });
    };

    const cancelEditing = () => {
        setEditingPostId(null);
        setEditFormData({ title: '', content: '' });
    };

    const handleEditChange = (e) => {
        const { name, value } = e.target;
        setEditFormData(prev => ({
            ...prev,
            [name]: value
        }));
    };

    const handleUpdatePost = async (postId) => {
        if (!isAuthenticated) {
            navigate('/login');
            return;
        }
        
        const postToUpdate = posts.find(p => p.id === postId);
         if (!postToUpdate || (currentUserId !== postToUpdate.author_id && currentUserRole !== 'admin')) {
             alert('У вас нет прав на обновление этого поста.');
             return;
         }

        try {
            await axios.put(
                `http://localhost:8080/api/v1/posts/${postId}`,
                {
                    title: editFormData.title,
                    content: editFormData.content
                },
                {
                    headers: {
                        'Authorization': `Bearer ${token}`,
                        'Content-Type': 'application/json'
                    }
                }
            );
            
            setEditingPostId(null);
            setPosts(prev => prev.map(post =>
                 post.id === postId
                     ? { ...post, title: editFormData.title, content: editFormData.content }
                     : post
            ));
            setError(null);

        } catch (error) {
            console.error('Update post error:', error);
            const errorMessage = error.response?.data?.error || 
                               error.response?.data?.message || 
                               'Не удалось обновить пост';
            alert(errorMessage);
        }
    };

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        const options = { year: 'numeric', month: 'long', day: 'numeric', hour: '2-digit', minute: '2-digit' };
        return date.toLocaleDateString('ru-RU', options);
    };

    return (
        <div className="post-list-container">
            {loading && <div className="loading-indicator">Загрузка постов...</div>}
            {error && <div className="error-message">Ошибка: {error}</div>}

            {!loading && !error && posts.length === 0 && (
                <div className="no-posts">Постов пока нет.</div>
            )}

            {posts.map(post => (
                <div key={post.id} className="post-item">
                    {editingPostId === post.id ? (
                        <div className="edit-form">
                            <h4>Редактировать пост</h4>
                            <input
                                type="text"
                                name="title"
                                value={editFormData.title}
                                onChange={handleEditChange}
                                className="edit-title-input"
                                placeholder="Заголовок"
                            />
                            <textarea
                                name="content"
                                value={editFormData.content}
                                onChange={handleEditChange}
                                className="edit-content-input"
                                rows={5}
                                placeholder="Содержимое поста"
                            />
                            <div className="edit-actions">
                                <button 
                                    onClick={() => handleUpdatePost(post.id)}
                                    className="save-button"
                                    disabled={!editFormData.title.trim() || !editFormData.content.trim()}
                                >
                                    Сохранить
                                </button>
                                <button 
                                    onClick={cancelEditing}
                                    className="cancel-button"
                                >
                                    Отмена
                                </button>
                            </div>
                        </div>
                    ) : (
                        <>
                            <div className="post-header">
                                <h3 className="post-title">{post.title}</h3>
                                {isAuthenticated && (currentUserId === post.author_id || currentUserRole === 'admin') && (
                                    <div className="post-actions">
                                        <button
                                            onClick={() => startEditing(post)}
                                            className="edit-button"
                                            title="Редактировать пост"
                                        >
                                            ✎
                                        </button>
                                        <button
                                            onClick={() => handleDeletePost(post.id, post.author_id)}
                                            className="delete-button"
                                            title={currentUserRole === 'admin' 
                                                ? "Удалить пост (админ)" 
                                                : "Удалить ваш пост"}
                                        >
                                            ✕
                                        </button>
                                    </div>
                                )}
                            </div>
                            <div className="post-content">
                                {post.content.split('\n').map((p, i) => (
                                    <p key={i}>{p}</p>
                                ))}
                            </div>
                        </>
                    )}
                    
                    <div className="post-meta">
                        <span className="author">{post.author_name || `User #${post.author_id}`}</span>
                        <span className="separator">•</span>
                        <span className="timestamp">
                            {formatDate(post.created_at)}
                        </span>
                        {currentUserRole === 'admin' && post.author_id !== currentUserId && (
                            <span className="admin-badge">(Admin)</span>
                        )}
                    </div>
                    
                    <Comments postId={post.id} />
                </div>
            ))}
        </div>
    );
};

export default PostList;
