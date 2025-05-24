import React from 'react';
import { Navigate } from 'react-router-dom';

const Home = () => {
  return <Navigate to="/posts" replace />;
};

export default Home; 