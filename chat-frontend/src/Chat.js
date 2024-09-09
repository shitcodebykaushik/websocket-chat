import React, { useState, useEffect } from 'react';

const Chat = ({ username, socket, setMessage, message }) => {
    const [newMessage, setNewMessage] = useState('');

    // Handle message sending
    const handleSendMessage = (e) => {
        e.preventDefault();
        if (socket) {
            socket.send(JSON.stringify({
                username,
                message: newMessage,
                'user-id': username
            }));
            setNewMessage('');
        }
    };

    // Handle received messages
    useEffect(() => {
        if (socket) {
            socket.onmessage = (event) => {
                const msg = JSON.parse(event.data);
                console.log('Received message:', msg); // Debug log
                setMessage(prev => `${prev}\n${msg.username}: ${msg.message}`);
            };

            socket.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

            socket.onclose = () => {
                console.log('WebSocket connection closed');
            };

            return () => socket.close();
        }
    }, [socket]);

    return (
        <div>
            <h2>Chat Room</h2>
            <form onSubmit={handleSendMessage}>
                <input
                    type="text"
                    placeholder="Type your message here"
                    value={newMessage}
                    onChange={(e) => setNewMessage(e.target.value)}
                />
                <button type="submit">Send Message</button>
            </form>
            <textarea
                readOnly
                value={message}
                rows="10"
                cols="50"
                style={{ width: '100%' }}
            ></textarea>
        </div>
    );
};

export default Chat;
