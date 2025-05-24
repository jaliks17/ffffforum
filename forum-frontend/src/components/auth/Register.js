import React, { useState } from 'react';
import axios from 'axios';
import { useNavigate } from 'react-router-dom';
import '../MainLayout.css';
import authService from '../../services/authService';

const Register = () => {
    const [formData, setFormData] = useState({
        username: '',
        password: '',
        confirmPassword: '',
        role: 'user'
    });
    const [error, setError] = useState('');
    const [isLoading, setIsLoading] = useState(false);
    const navigate = useNavigate();

    const handleInputChange = (e) => {
        const { name, value } = e.target;
        setFormData(prev => ({
            ...prev,
            [name]: value
        }));
    };

    const validateForm = () => {
        if (!formData.username || !formData.password || !formData.confirmPassword) {
            setError('Все поля обязательны для заполнения');
            return false;
        }

        if (formData.password.length < 8) {
            setError('Пароль должен содержать минимум 8 символов');
            return false;
        }

        if (formData.password !== formData.confirmPassword) {
            setError('Пароли не совпадают');
            return false;
        }

        return true;
    };

    const handleSubmit = async (e) => {
        e.preventDefault();
        setError('');

        if (!validateForm()) return;

        setIsLoading(true);

        try {
            const response = await authService.register(
                formData.username,
                formData.password,
                formData.role
            );

            alert('Регистрация прошла успешно! Теперь вы можете войти.');
            navigate('/login');

        } catch (error) {
            console.error('Ошибка регистрации:', error);
            setError(error.message || 'Ошибка регистрации. Попробуйте снова');
        } finally {
            setIsLoading(false);
        }
    };

    return (
        <div className="register-page">
            <div className="register-container">
                <h2 className="register-title">Создать аккаунт</h2>
                
                <form className="register-form" onSubmit={handleSubmit}>
                    {error && <div className="error-message">{error}</div>}
                    
                    <div className="form-field">
                        <label htmlFor="username">Имя пользователя</label>
                        <input
                            type="text"
                            id="username"
                            name="username"
                            value={formData.username}
                            onChange={handleInputChange}
                            placeholder="Введите имя пользователя"
                        />
                    </div>
                    
                    <div className="form-field">
                        <label htmlFor="password">Пароль</label>
                        <input
                            type="password"
                            id="password"
                            name="password"
                            value={formData.password}
                            onChange={handleInputChange}
                            placeholder="********"
                        />
                    </div>
                    
                    <div className="form-field">
                        <label htmlFor="confirmPassword">Подтвердите пароль</label>
                        <input
                            type="password"
                            id="confirmPassword"
                            name="confirmPassword"
                            value={formData.confirmPassword}
                            onChange={handleInputChange}
                            placeholder="Подтвердите свой пароль"
                        />
                    </div>
                    
                    <button 
                        type="submit" 
                        className="register-button"
                        disabled={isLoading}
                    >
                        {isLoading ? 'Регистрация...' : 'Регистрация'}
                    </button>
                </form>
                
                <div className="login-redirect">
                    У вас уже есть аккаунт? <span onClick={() => navigate('/login')}>Вход</span>
                </div>
            </div>
        </div>
    );
};

export default Register;