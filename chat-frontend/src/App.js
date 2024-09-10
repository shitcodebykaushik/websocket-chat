import React, { useState, useEffect, useCallback } from 'react';
import axios from 'axios';
import Register from './Register';
import Chat from './Chat';
import Login from './Login';

const App = () => {
    const [showPage, setShowPage] = useState('home'); // 'home', 'login', 'register', 'chat'
    const [username, setUsername] = useState('');
    const [password, setPassword] = useState('');
    const [socket, setSocket] = useState(null);

    const connectWebSocket = useCallback(() => {
        if (!socket) {
            const ws = new WebSocket('ws://localhost:8080/ws');
            setSocket(ws);

            ws.onopen = () => {
                console.log('WebSocket connection established');
            };

            ws.onmessage = (event) => {
                const msg = JSON.parse(event.data);
                console.log('Received message:', msg);
                setSocket(ws);
            };

            ws.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

            ws.onclose = () => {
                console.log('WebSocket connection closed');
                setSocket(null);
            };
        }
    }, [socket]);

    useEffect(() => {
        if (showPage === 'chat') {
            connectWebSocket();
        }

        return () => {
            if (socket) {
                socket.close();
            }
        };
    }, [showPage, connectWebSocket, socket]);

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
            <h1>Broadcast channel</h1>
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
                />
            )}
        </div>
    );
};

export default App;
