import React, { useState, useEffect } from 'react';
import axios from 'axios';
import Register from './Register';
import Chat from './Chat';
import Login from './Login';

const App = () => {
    const [showPage, setShowPage] = useState('home'); // 'home', 'login', 'register', 'chat'
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [message, setMessage] = useState('');
    const [socket, setSocket] = useState(null);

    useEffect(() => {
        const checkLogin = async () => {
            try {
                const res = await axios.get('/login/check');
                if (res.status === 200) {
                    setShowPage('chat');
                } else {
                    setShowPage('home');
                }
            } catch (error) {
                setShowPage('home');
            }
        };

        checkLogin();
    }, []);

    useEffect(() => {
        if (showPage === 'chat' && !socket) {
            const ws = new WebSocket('ws://localhost:8080/ws');
            setSocket(ws);

            ws.onopen = () => {
                console.log('WebSocket connection established');
            };

            ws.onmessage = (event) => {
                const msg = JSON.parse(event.data);
                console.log('Received message:', msg); // Debug log
                setMessage(prev => `${prev}\n${msg.username}: ${msg.message}`);
            };

            ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

            ws.onclose = () => {
                console.log('WebSocket connection closed');
            };

            return () => ws.close();
        }
    }, [showPage, socket]);

    const handleLogin = async () => {
        try {
            await axios.post('/login', { username, password });
            setShowPage('chat');
        } catch (error) {
            console.error('Error logging in:', error);
            alert('Login failed!');
        }
    };

    const handleRegister = async (username, password) => {
        try {
            await axios.post('/register', { username, password });
            setShowPage('login');
        } catch (error) {
            console.error('Error registering:', error);
            alert('Registration failed!');
        }
    };

    return (
        <div>
            <h1>Chat Application</h1>
            {showPage === 'home' && (
                <div>
                    <button onClick={() => setShowPage('login')}>Login</button>
                    <button onClick={() => setShowPage('register')}>Sign Up</button>
                </div>
            )}
            {showPage === 'login' && (
                <Login
                    username={username}
                    password={password}
                    setUsername={setUsername}
                    setPassword={setPassword}
                    handleLogin={handleLogin}
                />
            )}
            {showPage === 'register' && (
                <Register handleRegister={handleRegister} />
            )}
            {showPage === 'chat' && (
                <Chat
                    username={username}
                    socket={socket}
                    setMessage={setMessage}
                    message={message}
                />
            )}
        </div>
    );
};

export default App;
