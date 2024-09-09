package main

import (
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
    "context"
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true
    },
}

var clients = make(map[*websocket.Conn]bool)
var broadcast = make(chan Message)
var mu sync.Mutex
var mongoClient *mongo.Client
var usersCollection *mongo.Collection
var messagesCollection *mongo.Collection

type Message struct {
    Username  string `json:"username"`
    Message   string `json:"message"`
    UserId    string `json:"user-id"`
}

type User struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func hashPassword(password string) string {
    hash := sha256.Sum256([]byte(password))
    return hex.EncodeToString(hash[:])
}

func main() {
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

    usersCollection = mongoClient.Database("chatdb").Collection("users")
    messagesCollection = mongoClient.Database("chatdb").Collection("messages")

    http.HandleFunc("/", homePage)
    http.HandleFunc("/ws", handleConnections)
    http.HandleFunc("/register", handleRegister)
    http.HandleFunc("/login", handleLogin)
    http.HandleFunc("/login/check", checkLogin)

    go handleMessages()

    fmt.Println("Server started on :8080")
    err = http.ListenAndServe(":8080", nil)
    if err != nil {
        panic("Error starting server: " + err.Error())
    }
}

func homePage(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Welcome to the Chat Room!")
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
        Name:  "username",
        Value: credentials.Username,
        Path:  "/",
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
    cookie, err := r.Cookie("username")
    if err != nil {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    username := cookie.Value

    conn, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        fmt.Println("Error while upgrading connection:", err)
        return
    }
    defer conn.Close()

    mu.Lock()
    clients[conn] = true
    mu.Unlock()

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

        msg.Username = username
        msg.UserId = username

        saveMessageToDB(msg)

        broadcast <- msg
    }
}

func handleMessages() {
    for {
        msg := <-broadcast

        mu.Lock()
        for client := range clients {
            err := client.WriteJSON(msg)
            if err != nil {
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
        "timestamp": time.Now(),
    })
    if err != nil {
        fmt.Println("Error saving message to DB:", err)
    }
}
