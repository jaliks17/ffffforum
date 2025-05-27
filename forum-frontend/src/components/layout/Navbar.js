import React from 'react';
import '/Users/Лика/Desktop/ffffffffor/forum-frontend/src/components/MainLayout.css'; // Импортируйте CSS здесь
import { Link, useNavigate } from 'react-router-dom';
import { Button } from '@mui/material';

const Navbar = () => {
    const navigate = useNavigate();
    const token = localStorage.getItem('token');
    const userRole = localStorage.getItem('userRole');

    const handleLogout = () => {
        localStorage.removeItem('token');
        localStorage.removeItem('userId');
        localStorage.removeItem('userRole');
        navigate('/login');
    };

    return (
        <nav className="navbar">
            <h1 className="navTitle">Forum</h1>
            <div className="navLinks">
                <Link to="/posts">Посты</Link>
                <Link to="/chat">Чат</Link>
                {userRole === 'admin' && (
                    <Link to="/admin">Админ-панель</Link>
                )}
            </div>
            <div className="navLinks">
                {token ? (
                    <Button onClick={handleLogout} color="inherit">
                        Выйти
                    </Button>
                ) : (
                    <>
                        <Link to="/login">Войти</Link>
                        <Link to="/register">Регистрация</Link>
                    </>
                )}
            </div>
        </nav>
    );
};

export default Navbar;