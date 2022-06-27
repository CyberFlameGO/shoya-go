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
	Type    NotificationType
	Details interface{}
}
