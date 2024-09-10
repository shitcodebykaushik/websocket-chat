import React, { useState, useEffect, useCallback, useMemo } from 'react';

const Chat = ({ username, socket, setMessage, message }) => {
    const [newMessage, setNewMessage] = useState('');

    const messageObject = useMemo(() => ({
        username,
        message: newMessage,
        'user-id': username
    }), [username, newMessage]);

    const handleSendMessage = useCallback((e) => {
        e.preventDefault();
        if (socket) {
            socket.send(JSON.stringify(messageObject));
            setNewMessage('');
        }
    }, [socket, messageObject]);

    const handleMessage = useCallback((event) => {
        const msg = JSON.parse(event.data);
        console.log('Received message:', msg); // Debug log
        setMessage(prev => `${prev}\n${msg.username}: ${msg.message}`);
    }, [setMessage]);

    useEffect(() => {
        if (socket) {
            socket.onmessage = handleMessage;

            socket.onerror = (error) => {
                console.error('WebSocket error:', error);
            };

            socket.onclose = () => {
                console.log('WebSocket connection closed');
            };

            return () => {
                socket.onmessage = null;
                socket.close();
            };
        }
    }, [socket, handleMessage]);

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
