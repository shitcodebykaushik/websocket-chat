package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var clients = make(map[*websocket.Conn]string) // Track client connections with usernames
var broadcast = make(chan Message)
var mu sync.Mutex
var mongoClient *mongo.Client
var usersCollection *mongo.Collection
var messagesCollection *mongo.Collection
var sosCollection *mongo.Collection // Dedicated SOS collection

type Message struct {
	Username  string    `json:"username"`
	Message   string    `json:"message"`
	UserId    string    `json:"user-id"`
	Timestamp string    `json:"timestamp"`
	IP        string    `json:"ip"`
	IsSOS     bool      `json:"is_sos"` // New field to indicate if this is an SOS alert
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

func hashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}

func formatTimestamp(ts time.Time) string {
	return ts.Format("2006-01-02 15:04:05")
}

func main() {
	// MongoDB connection setup
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	var err error
	mongoClient, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		panic("Error connecting to MongoDB: " + err.Error())
	}
	err = mongoClient.Ping(ctx, readpref.Primary())
	if err != nil {
		panic("Error pinging MongoDB: " + err.Error())
	}

	// Initialize MongoDB collections
	usersCollection = mongoClient.Database("chatdb").Collection("users")
	messagesCollection = mongoClient.Database("chatdb").Collection("messages")
	sosCollection = mongoClient.Database("chatdb").Collection("sos_alerts") // SOS collection

	// HTTP Handlers
	http.HandleFunc("/", homePage)
	http.HandleFunc("/ws", handleConnections)
	http.HandleFunc("/register", handleRegister)
	http.HandleFunc("/login", handleLogin)
	http.HandleFunc("/login/check", checkLogin)

	// Message handler goroutine
	go handleMessages()

	fmt.Println("Server started on :8080")
	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		panic("Error starting server: " + err.Error())
	}
}

func homePage(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the SOS Alert System!")
}

func handleRegister(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hashedPassword := hashPassword(user.Password)

	_, err = usersCollection.InsertOne(context.TODO(), bson.M{
		"username": user.Username,
		"password": hashedPassword,
	})
	if err != nil {
		http.Error(w, "Error registering user", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

	var credentials User
	err := json.NewDecoder(r.Body).Decode(&credentials)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	hashedPassword := hashPassword(credentials.Password)

	var result User
	err = usersCollection.FindOne(context.TODO(), bson.M{
		"username": credentials.Username,
		"password": hashedPassword,
	}).Decode(&result)
	if err != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "username",
		Value:  credentials.Username,
		Path:   "/",
		MaxAge: 3600,
	})
	w.WriteHeader(http.StatusOK)
}

func checkLogin(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("username")
	if err != nil || cookie.Value == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
	// Get the username from the cookie
	cookie, err := r.Cookie("username")
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	username := cookie.Value // Assign username from cookie

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("Error while upgrading connection:", err)
		return
	}
	defer conn.Close()

	// Add the connected user to the clients map
	mu.Lock()
	clients[conn] = username // Use the username in the map
	mu.Unlock()

	// Extract the IP address
	ipAddress := r.RemoteAddr

	// Send old messages to the new user
	sendOldMessages(conn, ipAddress)

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			fmt.Println("Error while reading message:", err)
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			return
		}

		// Populate the message with relevant details
		msg.Username = username // Use the username here
		msg.UserId = username
		msg.Timestamp = formatTimestamp(time.Now())
		msg.IP = ipAddress

		if msg.IsSOS {
			handleSOSAlert(msg)
		} else {
			saveMessageToDB(msg)
			broadcast <- msg
		}
	}
}

func sendOldMessages(conn *websocket.Conn, ipAddress string) {
	// Fetch old messages from the database with a limit
	cursor, err := messagesCollection.Find(context.TODO(), bson.M{}, options.Find().SetLimit(100).SetSort(bson.M{"timestamp": -1}))
	if err != nil {
		fmt.Println("Error fetching old messages:", err)
		return
	}
	defer cursor.Close(context.TODO())

	var messages []Message
	if err := cursor.All(context.TODO(), &messages); err != nil {
		fmt.Println("Error decoding old messages:", err)
		return
	}

	// Send old messages to the new user
	for _, msg := range messages {
		if err := conn.WriteJSON(msg); err != nil {
			fmt.Println("Error while sending old message:", err)
			return
		}
	}
}

func handleMessages() {
	for {
		msg := <-broadcast

		fmt.Printf("Broadcasting message from %s: %s\n", msg.Username, msg.Message)

		mu.Lock()
		for client, username := range clients {
			fmt.Printf("Sending message to %s\n", username)
			if err := client.WriteJSON(msg); err != nil {
				fmt.Println("Error while writing message:", err)
				client.Close()
				delete(clients, client)
			}
		}
		mu.Unlock()
	}
}

func saveMessageToDB(msg Message) {
	_, err := messagesCollection.InsertOne(context.TODO(), bson.M{
		"username":  msg.Username,
		"message":   msg.Message,
		"user-id":   msg.UserId,
		"timestamp": msg.Timestamp,
		"ip":        msg.IP,
		"is_sos":    msg.IsSOS,
	})
	if err != nil {
		fmt.Println("Error saving message to DB:", err)
	}
}

func handleSOSAlert(msg Message) {
	// Save SOS alert to a dedicated collection
	_, err := sosCollection.InsertOne(context.TODO(), bson.M{
		"username":  msg.Username,
		"message":   msg.Message,
		"user-id":   msg.UserId,
		"timestamp": msg.Timestamp,
		"ip":        msg.IP,
		"is_sos":    true,
	})
	if err != nil {
		fmt.Println("Error saving SOS message to DB:", err)
		return
	}

	// Broadcast SOS to all connected clients
	fmt.Printf("SOS Alert from %s: %s\n", msg.Username, msg.Message)

	mu.Lock()
	for client, username := range clients {
		if err := client.WriteJSON(msg); err != nil {
			fmt.Println("Error while sending SOS message to", username, ":", err)
			client.Close()
			delete(clients, client)
		}
	}
	mu.Unlock()
}
