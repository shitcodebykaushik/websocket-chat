import React, { useState, useEffect } from 'react';

const Chat = ({ username, socket }) => {
    const [input, setInput] = useState('');
    const [messages, setMessages] = useState('');

    const sendMessage = () => {
        if (socket && input.trim()) {
            const msg = {
                username: username,
                message: input,
            };
            socket.send(JSON.stringify(msg));
            setInput('');  // Clear input after sending
        } else {
            console.error('Cannot send empty message or socket is not connected.');
        }
    };

    useEffect(() => {
        if (socket) {
            socket.onmessage = (event) => {
                const msg = JSON.parse(event.data);
                setMessages(prev => `${prev}\n${msg.username} [${msg.ip}]: ${msg.message}`);
            };
        }
    }, [socket]);

    return (
        <div style={{ backgroundColor: '#f9f9f9', padding: '10px', borderRadius: '5px', boxShadow: '0 2px 4px rgba(0, 0, 0, 0.1)' }}>
            <h2>Chat as {username}</h2>
            <div style={{ backgroundColor: '#fff', padding: '10px', borderRadius: '5px', marginBottom: '10px' }}>
                <textarea value={messages} readOnly style={{ width: '100%', minHeight: '150px', padding: '5px' }} />
            </div>
            <input
                type="text"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                placeholder="Type your message..."
                style={{ width: '100%', padding: '5px', marginBottom: '10px', borderRadius: '5px', border: '1px solid #ccc' }}
            />
            <button style={{ padding: '5px 10px', backgroundColor: '#007bff', color: '#fff', border: 'none', borderRadius: '5px', cursor: 'pointer' }} onClick={sendMessage}>Send</button>
        </div>
    );
};

export default Chat;
