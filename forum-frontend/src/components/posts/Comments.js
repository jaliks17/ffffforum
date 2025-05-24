import React, { useState, useEffect, useCallback } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';
import '/Users/Лика/Desktop/ffffffffor/forum-frontend/src/components/MainLayout.css';

const Comments = ({ postId }) => {
    const [comments, setComments] = useState([]);
    const [newComment, setNewComment] = useState('');
    const [loading, setLoading] = useState(false);
    const [error, setError] = useState(null);
    const [submitting, setSubmitting] = useState(false);
    
    // User data from localStorage
    const token = localStorage.getItem('token');
    const userId = localStorage.getItem('userId');
    const username = localStorage.getItem('username');
    const avatar = localStorage.getItem('avatar') || '/default-avatar.png';
    
    const currentUser = userId ? {
        id: parseInt(userId, 10),
        username: username || `User #${userId}`,
        avatar: avatar
    } : null;
    
    const isAuthenticated = !!token && !!currentUser;
    const navigate = useNavigate();

    const fetchComments = useCallback(async () => {
        try {
            setLoading(true);
            setError(null);

            const config = token ? {
                headers: { 
                    'Authorization': `Bearer ${token}`
                }
            } : {};

            const response = await axios.get(
                `http://localhost:8081/api/v1/posts/${postId}/comments`,
                config
            );

            const processedComments = (response.data?.data || response.data?.comments || response.data || [])
                .map(comment => ({
                    id: parseInt(comment.id, 10),
                    author_id: parseInt(comment.author_id, 10),
                    post_id: parseInt(comment.post_id, 10),
                    content: comment.content || '',
                    author_name: comment.author_name || `User #${comment.author_id}`,
                    author_avatar: comment.author_avatar || '/default-avatar.png',
                    created_at: comment.created_at || new Date().toISOString(),
                    likes: comment.likes || 0,
                    is_liked: comment.is_liked || false
                }))
                .sort((a, b) => new Date(b.created_at) - new Date(a.created_at));

            setComments(processedComments);
        } catch (err) {
            setError(err.response?.data?.error || err.message || 'Failed to load comments');
        } finally {
            setLoading(false);
        }
    }, [postId, token]);

    const handleSubmitComment = async (e) => {
        e.preventDefault();
        
        if (!isAuthenticated) {
            navigate('/login');
            return;
        }

        if (!newComment.trim() || submitting) return;

        try {
            setSubmitting(true);
            const response = await axios.post(
                `http://localhost:8081/api/v1/posts/${postId}/comments`,
                { 
                    content: newComment,
                    author_id: currentUser.id
                },
                { 
                    headers: { 
                        Authorization: `Bearer ${token}`,
                        'Content-Type': 'application/json' 
                    }
                }
            );

            const newCommentData = {
                ...response.data,
                author_name: currentUser.username,
                author_avatar: currentUser.avatar,
                created_at: new Date().toISOString(),
                content: newComment,
                author_id: currentUser.id,
                post_id: postId,
                likes: 0,
                is_liked: false
            };

            setComments(prev => [newCommentData, ...prev]);
            setNewComment('');
            setError(null);
        } catch (err) {
            setError(err.response?.data?.error || err.message || 'Failed to post comment');
        } finally {
            setSubmitting(false);
        }
    };

    const handleLikeComment = async (commentId) => {
        if (!isAuthenticated) {
            navigate('/login');
            return;
        }

        try {
            const response = await axios.post(
                `http://localhost:8081/api/v1/comments/${commentId}/like`,
                {},
                { 
                    headers: { 
                        Authorization: `Bearer ${token}`,
                        'Content-Type': 'application/json' 
                    }
                }
            );

            setComments(prev => prev.map(comment => 
                comment.id === commentId 
                    ? { 
                        ...comment, 
                        likes: response.data.likes,
                        is_liked: response.data.is_liked
                      } 
                    : comment
            ));
        } catch (err) {
            setError(err.response?.data?.error || err.message || 'Failed to like comment');
        }
    };

    useEffect(() => {
        if (postId) {
            fetchComments();
        }
    }, [postId, fetchComments]);

    const formatDate = (dateString) => {
        const date = new Date(dateString);
        const now = new Date();
        const diffInHours = (now - date) / (1000 * 60 * 60);

        if (diffInHours < 24) {
            return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
        } else {
            return date.toLocaleDateString([], { month: 'short', day: 'numeric' });
        }
    };

    return (
        <div className="comments-section">
            <h3 className="comments-title">
                Discussion ({comments.length})
                {loading && <span className="loading-indicator">Loading...</span>}
            </h3>
            
            {isAuthenticated ? (
                <form onSubmit={handleSubmitComment} className="comment-form">
                    <div className="comment-input-container">
                        <img 
                            src={currentUser.avatar} 
                            alt={currentUser.username} 
                            className="comment-avatar" 
                        />
                        <textarea
                            value={newComment}
                            onChange={(e) => setNewComment(e.target.value)}
                            placeholder="Share your thoughts..."
                            rows="3"
                            disabled={submitting}
                            required
                            className="comment-textarea"
                        />
                    </div>
                    <div className="comment-submit-container">
                        <button 
                            type="submit" 
                            className="submit-comment-btn"
                            disabled={submitting || !newComment.trim()}
                        >
                            {submitting ? (
                                <span className="spinner"></span>
                            ) : (
                                'Post Comment'
                            )}
                        </button>
                    </div>
                </form>
            ) : (
                <div className="login-prompt">
                    <p>
                        <a href="/login" className="login-link">Sign in</a> to join the conversation
                    </p>
                </div>
            )}

            {error && (
                <div className="error-message">
                    <span className="error-icon">⚠️</span> {error}
                </div>
            )}

            <div className="comments-list">
                {comments.length === 0 && !loading ? (
                    <div className="no-comments">
                        <p>No comments yet. Be the first to share your thoughts!</p>
                    </div>
                ) : (
                    comments.map(comment => (
                        <div key={comment.id} className={`comment-item ${comment.author_id === currentUser?.id ? 'own-comment' : ''}`}>
                            <div className="comment-avatar-container">
                                <img 
                                    src={comment.author_avatar} 
                                    alt={comment.author_name} 
                                    className="comment-avatar" 
                                />
                            </div>
                            <div className="comment-content-container">
                                <div className="comment-header">
                                    <span className="comment-author">
                                        {comment.author_name}
                                    </span>
                                    <span className="comment-timestamp">
                                        {formatDate(comment.created_at)}
                                    </span>
                                </div>
                                <div className="comment-text">
                                    {comment.content.split('\n').map((line, i) => (
                                        <p key={i}>{line}</p>
                                    ))}
                                </div>
                                <div className="comment-actions">
                                    <button 
                                        className={`like-button ${comment.is_liked ? 'liked' : ''}`}
                                        onClick={() => handleLikeComment(comment.id)}
                                    >
                                        ❤️ {comment.likes > 0 && comment.likes}
                                    </button>
                                    <span className="comment-reply">Reply</span>
                                </div>
                            </div>
                        </div>
                    ))
                )}
            </div>
        </div>
    );
};

export default Comments;