import axios from 'axios';

const API_URL = 'http://localhost:8081/api/v1/auth';

export async function register(username, password, role = 'user') {
  try {
    const response = await fetch(`${API_URL}/signup`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password, role }),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Registration failed');
    }

    return await response.json();
  } catch (error) {
    console.error('Registration error:', error);
    throw error;
  }
}

export async function login(username, password) {
  try {
    const response = await fetch(`${API_URL}/signin`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({ username, password }),
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Login failed');
    }

    const data = await response.json();
    console.log('Данные от бэкенда после входа:', data);
    localStorage.setItem('token', data.access_token);
    localStorage.setItem('refreshToken', data.refresh_token);

    if (data.user) {
        localStorage.setItem('userId', data.user.id);
        localStorage.setItem('username', data.user.username);
        if (data.user.role) localStorage.setItem('userRole', data.user.role);
    } else {
        console.warn('Login response did not contain user data.');
    }

    return data;
  } catch (error) {
    console.error('Login error:', error);
    throw error;
  }
}

export function logout() {
  localStorage.removeItem('token');
  localStorage.removeItem('refreshToken');
}

export function getToken() {
  return localStorage.getItem('token');
}

export function getRefreshToken() {
  return localStorage.getItem('refreshToken');
}

export default { register, login, logout, getToken, getRefreshToken };
