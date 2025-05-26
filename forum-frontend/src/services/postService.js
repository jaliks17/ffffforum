const API_URL = 'http://localhost:8080/api/v1';

export async function fetchPosts() {
  try {
    const response = await fetch(`${API_URL}/posts`, {
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    if (!response.ok) {
      throw new Error('Failed to fetch posts');
    }
    
    return await response.json();
  } catch (error) {
    console.error('Fetch posts error:', error);
    throw error;
  }
}

export async function fetchPost(id) {
  try {
    const response = await fetch(`${API_URL}/posts/${id}`, {
      headers: {
        'Content-Type': 'application/json',
      },
    });
    
    if (!response.ok) {
      throw new Error('Failed to fetch post');
    }
    
    return await response.json();
  } catch (error) {
    console.error('Fetch post error:', error);
    throw error;
  }
}

export async function createPost(postData, token) {
  try {
    const response = await fetch(`${API_URL}/posts`, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify(postData),
    });
    
    if (!response.ok) {
      throw new Error('Failed to create post');
    }
    
    return await response.json();
  } catch (error) {
    console.error('Create post error:', error);
    throw error;
  }
}

export default { fetchPosts, fetchPost, createPost }; 