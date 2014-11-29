// server
package main

import (
	"encoding/json"
	"net"
	"log"
	"sync"
)

var (
	listUsers map[string]net.Conn = make(map[string]net.Conn)//list of users currently connected
	testing bool = false
	mutex = new(sync.Mutex)
)

//message sent between computers
type Message struct {
	Kind      string //type of message ("CONNECT","PRIVATE","PUBLIC","DISCONNECT","HEARTBEAT")
	Username  string //my username
	Receiver  string //if its a private message specifies the receiver of the message.
	MSG       string //message
	Usernames []string       
}

//waits for new connections
func main() {
	tcpAddr, err := net.ResolveTCPAddr("tcp4", ":12100")
	checkError(err)
	listener, err := net.ListenTCP("tcp", tcpAddr)
	checkError(err)
	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		go handleClient(conn)
	}
}

//handles a client connection by waiting on its messages
func handleClient(conn net.Conn) {
	if testing {log.Println("handleClient")}
	defer conn.Close()
	dec := json.NewDecoder(conn)
	var msg Message
	for {
		err := dec.Decode(&msg);
		if err != nil {
			unexpectedClosing(conn)
			return
		}
		msg.handle(conn)
	}
}

//decides what to do depending on the kind o message
func (msg *Message) handle(conn net.Conn) {
	if testing {log.Println("handle")}
	switch msg.Kind {
	case "CONNECT":
		addConnection(msg, conn)
	case "PRIVATE":
		msg.sendPrivate()
	case "PUBLIC":
		msg.sendPublic()
	case "DISCONNECT":
		removeConnection(msg)
	case "HEARTBEAT":
		//NOT SURE YET
	}
}

//sends the message to every member of the chat
func (msg *Message) sendPublic() {
	if testing {log.Println("sendPublic")}
	for _, user := range listUsers {
		enc := json.NewEncoder(user)
		err:= enc.Encode(msg)
		checkError(err)
	}
	log.Println(msg.Username+" sent a public message")
}

//sends a message just to the user save in msg.Receiver
func (msg *Message) sendPrivate() {
	if testing {log.Println("sendPrivate")}
	user:=listUsers[msg.Receiver]
	enc := json.NewEncoder(user)
	err:= enc.Encode(msg)
	checkError(err)
	log.Println(msg.Username+" sent a private message to "+msg.Receiver)
}

//adds a new member to the chat and alerts each chat member
func addConnection(msg *Message, conn net.Conn) {
	if testing {log.Println("addConnection")}
	newMessage:=new(Message)
	newMessage.Kind="ADD"
	newMessage.MSG=msg.Username+" joined the chat"
	handleSameUserName(msg,conn)
	mutex.Lock()
	listUsers[msg.Username]=conn
	mutex.Unlock()
	log.Println("listUsers",listUsers)
	for userName,_:= range listUsers {
		newMessage.Usernames = append(newMessage.Usernames, userName)
	}
	newMessage.sendPublic()
	log.Println(msg.Username+" joined the chat")
}

//removes one member from the chat and alerts every chat member
func removeConnection(msg *Message){
	if testing {log.Println("removeConnection")}
	mutex.Lock()
	delete(listUsers,msg.Username)
	mutex.Unlock()
	newMessage:=new(Message)
	newMessage.Kind="DISCONNECT"
	newMessage.MSG=msg.Username+" left the chat"
	for userName,_:= range listUsers {
		newMessage.Usernames = append(newMessage.Usernames, userName)
	}
	newMessage.sendPublic()
	log.Println(msg.Username+" left the chat") 
}

//handles when a user just shuts down the window
func unexpectedClosing(conn net.Conn){
	if testing {log.Println("unexpectedClosing")}
	for userName,connection:= range listUsers {
		if connection == conn {
			done:=new(Message)
			done.Username = userName
			removeConnection(done)
		} 
	}
}
//changes the userName to userName+1 if there is a user with that userName already
func handleSameUserName(msg *Message, conn net.Conn){
	if testing {log.Println("handleSameUser")}
	for userName,_:= range listUsers {
		if(userName==msg.Username){
			msg.Username=userName+"1"
			newMessage:=new(Message)
			newMessage.Username=msg.Username
			newMessage.Receiver=msg.Username
			newMessage.Kind="SAMENAME"
			newMessage.MSG="Your name was already in use, so we decided to update it to "+msg.Username
			mutex.Lock()
			listUsers[msg.Username]=conn
			mutex.Unlock()
			newMessage.sendPrivate()
			break	
		}
	}
}

//check if there is an error
func checkError(err error) {
	if err != nil {
		panic(err)
	}
}

