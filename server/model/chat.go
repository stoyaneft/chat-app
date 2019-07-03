package model

type ChatRoom struct {
	UID          string
	Users        map[User]struct{}
	MessageQueue chan ChatMessage
	SystemQueue  chan SystemMessage
}

type SystemMessage struct {
	Type    SystemMessageType
	UserID  int64
	ChatUID string
}

type SystemMessageType int

const SystemUserAdded SystemMessageType = iota

type Message struct {
	Email        string     `json:"email"`
	Username     string     `json:"username"`
	Password     string     `json:"password"`
	Message      string     `json:"message"`
	Type         message    `json:"type"`
	ChatUID      string     `json:"chatUID"`
	Participants []string   `json:"participants"`
	Chats        []ChatInfo `json:"chats"`
}

type message string

const (
	RegistrationMessage  message = "registration"
	LoginMessage         message = "login"
	CreateChatMessage    message = "createChat"
	ChatSelectionMessage message = "chatSelection"
	AddUserMessage       message = "addUser"
	SendMessage          message = "sendMessage"
)

type ChatInfo struct {
	UID          string
	Participants []string
	Messages     []ChatMessage
}

type Chat struct {
	UID  string
	Name string
}

type ChatMessage struct {
	Username  string
	ChatUID   string
	Message   string
	CreatedAt int64
}
