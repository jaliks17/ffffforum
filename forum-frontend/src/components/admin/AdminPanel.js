import React, { useState, useEffect } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';
import {
  Box,
  Tabs,
  Tab,
  Typography,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Button,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Alert
} from '@mui/material';

function TabPanel({ children, value, index }) {
  return (
    <div hidden={value !== index}>
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

function AdminPanel() {
  const [value, setValue] = useState(0);
  const [users, setUsers] = useState([]);
  const [posts, setPosts] = useState([]);
  const [comments, setComments] = useState([]);
  const [selectedUser, setSelectedUser] = useState(null);
  const [openDialog, setOpenDialog] = useState(false);
  const [error, setError] = useState(null);
  const navigate = useNavigate();
  const token = localStorage.getItem('token');

  useEffect(() => {
    if (!token) {
      navigate('/login');
      return;
    }
    fetchData();
  }, [token, navigate]);

  const fetchData = async () => {
    try {
      // Получаем данные пользователей
      const usersResponse = await axios.get('http://localhost:50051/api/v1/users', {
        headers: { Authorization: `Bearer ${token}` }
      });
      setUsers(usersResponse.data.users);

      // Получаем все посты
      const postsResponse = await axios.get('http://localhost:8080/api/v1/posts', {
        headers: { Authorization: `Bearer ${token}` }
      });
      setPosts(postsResponse.data.posts);

      // Получаем все комментарии
      const commentsResponse = await axios.get('http://localhost:8080/api/v1/comments', {
        headers: { Authorization: `Bearer ${token}` }
      });
      setComments(commentsResponse.data.comments);
    } catch (err) {
      setError('Ошибка при загрузке данных');
      console.error('Error fetching data:', err);
    }
  };

  const handleChangeRole = async (userId, newRole) => {
    try {
      await axios.put(`http://localhost:50051/api/v1/users/${userId}/role`, 
        { role: newRole },
        { headers: { Authorization: `Bearer ${token}` } }
      );
      fetchData();
    } catch (err) {
      setError('Ошибка при изменении роли');
    }
  };

  const handleDeleteUser = async (userId) => {
    if (window.confirm('Вы уверены, что хотите удалить этого пользователя?')) {
      try {
        await axios.delete(`http://localhost:50051/api/v1/users/${userId}`, {
          headers: { Authorization: `Bearer ${token}` }
        });
        fetchData();
      } catch (err) {
        setError('Ошибка при удалении пользователя');
      }
    }
  };

  const handleDeletePost = async (postId) => {
    if (window.confirm('Вы уверены, что хотите удалить этот пост?')) {
      try {
        await axios.delete(`http://localhost:8080/api/v1/posts/${postId}`, {
          headers: { Authorization: `Bearer ${token}` }
        });
        fetchData();
      } catch (err) {
        setError('Ошибка при удалении поста');
      }
    }
  };

  const handleDeleteComment = async (commentId) => {
    if (window.confirm('Вы уверены, что хотите удалить этот комментарий?')) {
      try {
        await axios.delete(`http://localhost:8080/api/v1/comments/${commentId}`, {
          headers: { Authorization: `Bearer ${token}` }
        });
        fetchData();
      } catch (err) {
        setError('Ошибка при удалении комментария');
      }
    }
  };

  return (
    <Box sx={{ width: '100%', maxWidth: 1200, margin: '0 auto', padding: 3 }}>
      <Typography variant="h4" gutterBottom>
        Панель администратора
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      <Box sx={{ borderBottom: 1, borderColor: 'divider' }}>
        <Tabs value={value} onChange={(e, newValue) => setValue(newValue)}>
          <Tab label="Пользователи" />
          <Tab label="Посты" />
          <Tab label="Комментарии" />
        </Tabs>
      </Box>

      <TabPanel value={value} index={0}>
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>ID</TableCell>
                <TableCell>Имя пользователя</TableCell>
                <TableCell>Роль</TableCell>
                <TableCell>Действия</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {users.map((user) => (
                <TableRow key={user.id}>
                  <TableCell>{user.id}</TableCell>
                  <TableCell>{user.username}</TableCell>
                  <TableCell>{user.role}</TableCell>
                  <TableCell>
                    <Button
                      onClick={() => handleChangeRole(user.id, user.role === 'admin' ? 'user' : 'admin')}
                      color="primary"
                      size="small"
                      sx={{ mr: 1 }}
                    >
                      {user.role === 'admin' ? 'Сделать пользователем' : 'Сделать админом'}
                    </Button>
                    <Button
                      onClick={() => handleDeleteUser(user.id)}
                      color="error"
                      size="small"
                    >
                      Удалить
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </TabPanel>

      <TabPanel value={value} index={1}>
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>ID</TableCell>
                <TableCell>Заголовок</TableCell>
                <TableCell>Автор</TableCell>
                <TableCell>Дата создания</TableCell>
                <TableCell>Действия</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {posts.map((post) => (
                <TableRow key={post.id}>
                  <TableCell>{post.id}</TableCell>
                  <TableCell>{post.title}</TableCell>
                  <TableCell>{post.author_name}</TableCell>
                  <TableCell>{new Date(post.created_at).toLocaleString()}</TableCell>
                  <TableCell>
                    <Button
                      onClick={() => handleDeletePost(post.id)}
                      color="error"
                      size="small"
                    >
                      Удалить
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </TabPanel>

      <TabPanel value={value} index={2}>
        <TableContainer component={Paper}>
          <Table>
            <TableHead>
              <TableRow>
                <TableCell>ID</TableCell>
                <TableCell>Содержание</TableCell>
                <TableCell>Автор</TableCell>
                <TableCell>Пост</TableCell>
                <TableCell>Действия</TableCell>
              </TableRow>
            </TableHead>
            <TableBody>
              {comments.map((comment) => (
                <TableRow key={comment.id}>
                  <TableCell>{comment.id}</TableCell>
                  <TableCell>{comment.content}</TableCell>
                  <TableCell>{comment.author_name}</TableCell>
                  <TableCell>{comment.post_id}</TableCell>
                  <TableCell>
                    <Button
                      onClick={() => handleDeleteComment(comment.id)}
                      color="error"
                      size="small"
                    >
                      Удалить
                    </Button>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </TableContainer>
      </TabPanel>
    </Box>
  );
}

export default AdminPanel; 