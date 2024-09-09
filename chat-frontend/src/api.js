import axios from 'axios';

const API_URL = 'http://localhost:8080';

export const registerUser = async (username, password) => {
  try {
    const response = await axios.post(`${API_URL}/register`, { username, password });
    return response.data;
  } catch (error) {
    console.error('Error registering user:', error.response?.data || error.message);
    throw error;
  }
};

export const loginUser = async (username, password) => {
  try {
    const response = await axios.post(`${API_URL}/login`, { username, password });
    return response.data;
  } catch (error) {
    console.error('Error logging in:', error.response?.data || error.message);
    throw error;
  }
};

export const connectToWebSocket = (username) => {
  const socket = new WebSocket(`ws://localhost:8080/ws`);
  socket.onopen = () => {
    console.log('Connected to WebSocket');
  };
  socket.onmessage = (event) => {
    console.log('Message received:', JSON.parse(event.data));
  };
  socket.onerror = (error) => {
    console.error('WebSocket error:', error);
  };
  return socket;
};
