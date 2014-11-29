package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"gopkg.in/qml.v0"
	"os"
	"strings"
)

var (
	output chan string = make(chan string) //channel waitin on the user to type something
	listUsers []string = make([]string, 0, 30) //users in the chat
	myName string //name of the client
	testing bool = false
)

type Control struct {
	Root        qml.Object
	convstring  string
	userlist    string
	inputString string
}

//message sent out to the server
type Message struct {
	Kind      string //type of message ("CONNECT","PRIVATE","PUBLIC","DISCONNECT","ADD")
	Username  string //my username
	Receiver  string //if its a private message specifies the receiver of the message.
	MSG       string //message
	Usernames []string
}

//start the connection, introduces the user to the chat and creates graphical interface.
func main() {
	myName= os.Args[2]
	//starting graphics
	qml.Init(nil)
	engine := qml.NewEngine()
	ctrl := Control{convstring: ""}
	ctrl.convstring = ""
	context := engine.Context()
	context.SetVar("ctrl", &ctrl)
	component, err := engine.LoadFile("chat.qml")
	if err != nil {
		fmt.Println("no file to load for ui")
		fmt.Println(err.Error())
		os.Exit(0)
	}
	win := component.CreateWindow(nil)
	ctrl.Root = win.Root()
	service:= os.Args[1]+":12100"
	tcpAddr, err := net.ResolveTCPAddr("tcp", service)
	handleErr(err)
	conn, err := net.DialTCP("tcp", nil, tcpAddr)
	handleErr(err)
	
	enc := json.NewEncoder(conn)
	dec := json.NewDecoder(conn)
	
	win.Show() //show window
	ctrl.updateText("Hello "+myName+".\nFor private messages, type the message followed by * and the name of the receiver.\n To leave the conversation type disconnect")
	
	introduceMyself(*enc)
	
	go send(*enc)
	go receive(*dec, &ctrl)
	
	win.Wait()
}

//sends message to the server
func send(enc json.Encoder){
	if testing {log.Println("send")}
	msg:=new(Message)
	for {
		message:= <-output
		whatever:=strings.Split(message,"*")
		msg.Username=myName
		if message=="disconnect"{
			msg.Kind="DISCONNECT"
			enc.Encode(msg)
			break
		} else if len(whatever)>1 {
			msg.Kind="PRIVATE"
			msg.Receiver=whatever[1]
			msg.MSG=whatever[0]
		} else {
			msg.Kind="PUBLIC"
			msg.MSG=whatever[0]
		}
		enc.Encode(msg)
	}
	os.Exit(1)
}

//receives message from server
func receive(dec json.Decoder, ctrl *Control){
	if testing {log.Println("receive")}
	msg:= new(Message)
	for {
		if err := dec.Decode(msg);err != nil {
			fmt.Println("Something went wrong, closing connection")
			panic(err)
			return
		}
		if msg.Kind=="PRIVATE"{
			ctrl.updateText(msg.Username+" wispers: "+msg.MSG)
		}else if msg.Kind=="PUBLIC"{
			ctrl.updateText(msg.Username+": "+msg.MSG)
		}else if msg.Kind=="ADD" || msg.Kind=="DISCONNECT"{
			ctrl.updateText(msg.MSG)
			ctrl.updateList(msg.Usernames)
		}else if msg.Kind=="SAMENAME"{
			myName=msg.Username
			ctrl.updateText(msg.MSG)
		}	
	}
}

//introduces client to the chat
func introduceMyself(enc json.Encoder){
	if testing {log.Println("introduceMyself")}
	msg:=new(Message)
	msg.Kind="CONNECT"
	msg.Username=myName
	enc.Encode(msg)
	
}

func handleErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

//Graphics methods

func (ctrl *Control) TextEntered(text qml.Object) {
	//this method is called whenever a return key is typed in the text entry field.  The qml object calls this function
	ctrl.inputString = text.String("text") //the ctrl's inputString field holds the message
	//you will want to send it to the server
	//but for now just send it back to the conv field
	//ctrl.updateText(ctrl.inputString)
	output <- ctrl.inputString

}

func (ctrl *Control) updateText(toAdd string) {
	//call this method whenever you want to add text to the qml object's conv field
	ctrl.convstring = ctrl.convstring + toAdd + "\n" //also keep track of everything in that field
	ctrl.Root.ObjectByName("conv").Set("text", ctrl.convstring)
	qml.Changed(ctrl, &ctrl.convstring)
}

func (ctrl *Control) updateList(list []string) {
	ctrl.userlist = ""
	for _, user := range list {
		ctrl.userlist += user + "\n"
	}
	ctrl.Root.ObjectByName("userlist").Set("text", ctrl.userlist)
	qml.Changed(ctrl, &ctrl.userlist)
}

