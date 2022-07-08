package models

// NotificationType is the type of notification.
type NotificationType string

const (
	NotificationTypeAll                   NotificationType = "all"
	NotificationTypeFriendRequest         NotificationType = "friendRequest"
	NotificationTypeInvite                NotificationType = "invite"
	NotificationTypeInviteResponse        NotificationType = "inviteResponse"
	NotificationTypeRequestInvite         NotificationType = "requestInvite"
	NotificationTypeRequestInviteResponse NotificationType = "requestInviteResponse"
	NotificationTypeVoteToKick            NotificationType = "voteToKick"
)

func (n NotificationType) String() string {
	return string(n)
}

func ParseNotificationType(s string) NotificationType {
	switch s {
	case "all":
		return NotificationTypeAll
	case "friendRequest":
		return NotificationTypeFriendRequest
	case "invite":
		return NotificationTypeInvite
	case "inviteResponse":
		return NotificationTypeInviteResponse
	case "requestInvite":
		return NotificationTypeRequestInvite
	case "requestInviteResponse":
		return NotificationTypeRequestInviteResponse
	case "voteToKick":
		return NotificationTypeVoteToKick
	default:
		return NotificationTypeAll
	}
}

type Notification struct {
	CreatedAt      string           `json:"created_at"`
	ID             string           `json:"id"`
	Type           NotificationType `json:"type"`
	Details        interface{}      `json:"details"` // Can be a string, a map[string]interface{}, or nil.
	Message        string           `json:"message"`
	Seen           *bool            `json:"seen,omitempty"`           // This can be not included (e.g.: when sent over websocket).
	SenderId       *string          `json:"senderUserId,omitempty"`   // This can be not included (e.g.: when sent over websocket).
	SenderUsername *string          `json:"senderUsername,omitempty"` // This can be not included (e.g.: when sent over websocket).
	ReceiverUserId *string          `json:"receiverUserId,omitempty"` // This can be not included (e.g.: when sent over websocket).
}
