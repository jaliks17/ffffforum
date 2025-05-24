import React, { useState } from 'react';

export default function CreatePost({ onCreate }) {
  const [title, setTitle] = useState('');
  const [content, setContent] = useState('');
  const handleSubmit = () => {
    // TODO: Реализовать отправку поста через API
    if (onCreate) onCreate({ title, content });
    setTitle('');
    setContent('');
  };
  return (
    <div>
      <h3>Создать пост</h3>
      <input placeholder="Заголовок" value={title} onChange={e => setTitle(e.target.value)} />
      <br />
      <textarea placeholder="Текст" value={content} onChange={e => setContent(e.target.value)} />
      <br />
      <button onClick={handleSubmit}>Опубликовать</button>
    </div>
  );
}
