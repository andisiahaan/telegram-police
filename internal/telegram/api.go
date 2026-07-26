package telegram

import (
	"encoding/json"
	"fmt"
)

// API exposes the Telegram Bot API methods used by this bot.
type API struct {
	client *Client
}

// NewAPI creates an API bound to the given client.
func NewAPI(client *Client) *API {
	return &API{client: client}
}

// DeleteMessage deletes a message from a chat.
func (a *API) DeleteMessage(chatID int64, messageID int) error {
	_, err := a.client.Do("deleteMessage", map[string]any{
		"chat_id":    chatID,
		"message_id": messageID,
	})
	return err
}

// BanChatMember bans a user from a chat until untilDate (Unix timestamp).
// Pass 0 for a permanent ban.
func (a *API) BanChatMember(chatID, userID int64, untilDate int64) error {
	_, err := a.client.Do("banChatMember", map[string]any{
		"chat_id":    chatID,
		"user_id":    userID,
		"until_date": untilDate,
	})
	return err
}

// GetChatMember returns membership information for a user in a chat.
func (a *API) GetChatMember(chatID, userID int64) (*ChatMember, error) {
	result, err := a.client.Do("getChatMember", map[string]any{
		"chat_id": chatID,
		"user_id": userID,
	})
	if err != nil {
		return nil, err
	}

	var member ChatMember
	if err := json.Unmarshal(result, &member); err != nil {
		return nil, fmt.Errorf("telegram api: getChatMember unmarshal: %w", err)
	}
	return &member, nil
}

// SendMessage sends a text message to a chat and returns the new message's ID,
// which is used by WarnAction for auto-deletion.
func (a *API) SendMessage(chatID int64, text string) (int, error) {
	result, err := a.client.Do("sendMessage", map[string]any{
		"chat_id": chatID,
		"text":    text,
	})
	if err != nil {
		return 0, err
	}

	var msg Message
	if err := json.Unmarshal(result, &msg); err != nil {
		return 0, fmt.Errorf("telegram api: sendMessage unmarshal: %w", err)
	}
	return msg.MessageID, nil
}
// GetUpdates fetches pending updates from Telegram using long-polling.
// offset should be the last update_id+1 to acknowledge previous updates.
// timeout is the long-poll wait in seconds (Telegram recommends 20–60).
func (a *API) GetUpdates(offset, timeout int) ([]Update, error) {
	result, err := a.client.Do("getUpdates", map[string]any{
		"offset":  offset,
		"timeout": timeout,
		"limit":   100,
	})
	if err != nil {
		return nil, err
	}

	var updates []Update
	if err := json.Unmarshal(result, &updates); err != nil {
		return nil, fmt.Errorf("telegram api: getUpdates unmarshal: %w", err)
	}
	return updates, nil
}

// DeleteWebhook removes the current webhook integration.
func (a *API) DeleteWebhook() error {
	_, err := a.client.Do("deleteWebhook", map[string]any{})
	return err
}

// SetWebhook registers the webhook URL with Telegram.
func (a *API) SetWebhook(url, secretToken string) error {
	params := map[string]any{
		"url": url,
	}
	if secretToken != "" {
		params["secret_token"] = secretToken
	}
	_, err := a.client.Do("setWebhook", params)
	return err
}
