import React, { useState } from 'react';
import axios from 'axios';

export default function CreatePost({ onCreate }) {
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const [error, setError] = useState(null);

  const handleSubmit = async () => {
    setError(null);

    const token = localStorage.getItem('token');
    if (!token) {
      setError('You need to be logged in to create a post.');
      return;
    }

    if (!title || !content) {
      setError('Title and content cannot be empty.');
      return;
    }

    try {
      const response = await axios.post(
        'http://localhost:8080/api/v1/posts',
        {
          title: title,
          content: content,
        },
        {
          headers: {
            Authorization: `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }
      );

      console.log('Post created successfully:', response.data);
      setTitle('');
      setContent('');
      if (onCreate) onCreate(response.data.post);

    } catch (err) {
      console.error('Error creating post:', err);
      setError(err.response?.data?.error || err.message || 'Failed to create post');
    }
  };

  return (
    <div className="create-post-container">
      <h3>Создать пост</h3>
      {error && <div style={{ color: 'red' }}>{error}</div>}
      <input 
        placeholder="Заголовок" 
        value={title} 
        onChange={e => setTitle(e.target.value)} 
        className="create-post-container input"
      />
      <br />
      <textarea 
        placeholder="Текст" 
        value={content} 
        onChange={e => setContent(e.target.value)} 
        className="create-post-container textarea"
      />
      <br />
      <button 
        onClick={handleSubmit} 
        className="create-post-container button"
      >
        Опубликовать
      </button>
    </div>
  );
}
