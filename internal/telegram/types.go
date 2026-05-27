package telegram

import "encoding/json"

// Update is the main payload Telegram sends to the webhook.
type Update struct {
	UpdateID      int      `json:"update_id"`
	Message       *Message `json:"message"`
	EditedMessage *Message `json:"edited_message"`
}

// Message holds the fields we care about from a Telegram message.
type Message struct {
	MessageID       int             `json:"message_id"`
	From            *User           `json:"from"`
	Chat            *Chat           `json:"chat"`
	Text            string          `json:"text"`
	Entities        []MessageEntity `json:"entities"`
	NewChatMembers  []User          `json:"new_chat_members"`
	LeftChatMember  *User           `json:"left_chat_member"`
	ForwardFrom     *User           `json:"forward_from"`
	ForwardFromChat *Chat           `json:"forward_from_chat"`
}

// User represents a Telegram user.
type User struct {
	ID        int64  `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Username  string `json:"username"`
}

// Chat represents a Telegram chat.
type Chat struct {
	ID   int64  `json:"id"`
	Type string `json:"type"`
}

// MessageEntity represents a special entity inside message text (mention, URL, etc.).
type MessageEntity struct {
	Type   string `json:"type"`   // mention, text_mention, url, ...
	Offset int    `json:"offset"`
	Length int    `json:"length"`
	User   *User  `json:"user"` // only set for text_mention
}

// ChatMember represents a chat member with their current status.
type ChatMember struct {
	User   User   `json:"user"`
	Status string `json:"status"` // creator, administrator, member, restricted, left, kicked
}

// APIResponse is the generic wrapper for every Telegram API response.
type APIResponse struct {
	OK          bool            `json:"ok"`
	Description string          `json:"description"`
	Result      json.RawMessage `json:"result"`
}
