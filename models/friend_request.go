package models

import (
	"gitlab.com/george/shoya-go/config"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// FriendRequestState is the state of a friend request.
type FriendRequestState string

const (
	FriendRequestStateSent     FriendRequestState = "sent"     // FriendRequestStateSent is the state of a friend request when it has been sent.
	FriendRequestStateAccepted FriendRequestState = "accepted" // FriendRequestStateAccepted is the state of a friend request when it has been accepted.
	FriendRequestStateIgnored  FriendRequestState = "ignored"  // FriendRequestStateIgnored is the state of a friend request when it has been ignored.
)

// FriendRequest represents a friend request between two users.
type FriendRequest struct {
	BaseModel
	FromID string             `json:"fromId"`
	From   User               `json:"fromUser"`
	ToID   string             `json:"toId"`
	To     User               `json:"toUser"`
	State  FriendRequestState `json:"state"`
}

// NewFriendRequest creates a new friend request between two users.
func NewFriendRequest(fromUser, toUser *User) *FriendRequest {
	return &FriendRequest{
		From:  *fromUser,
		To:    *toUser,
		State: FriendRequestStateSent,
	}
}

// Accept accepts a friend request.
func (f *FriendRequest) Accept() (bool, error) {
	if f.State == FriendRequestStateAccepted {
		return true, nil
	}

	changes := map[string]interface{}{
		"state": FriendRequestStateAccepted,
	}

	if tx := config.DB.Omit(clause.Associations).Model(&f).Updates(changes); tx.Error != nil {
		return false, tx.Error
	}

	return true, nil
}

// Deny denies a friend request and marks it as ignored.
func (f *FriendRequest) Deny() (bool, error) {
	if f.State == FriendRequestStateIgnored {
		return true, nil
	}

	changes := map[string]interface{}{
		"state": FriendRequestStateIgnored,
	}

	if tx := config.DB.Omit(clause.Associations).Model(&f).Updates(changes); tx.Error != nil {
		return false, tx.Error
	}

	return true, nil
}

// Delete deletes a friend request.
func (f *FriendRequest) Delete() (bool, error) {
	if tx := config.DB.Unscoped().Omit(clause.Associations).Delete(&f); tx.Error != nil {
		return false, tx.Error
	}

	return true, nil
}

// GetFriendRequestForUsers returns the friend request between two users.
func GetFriendRequestForUsers(u1, u2 string) (*FriendRequest, error) {
	var fr FriendRequest
	if tx := config.DB.Where("from_id = ? AND to_id = ?", u1, u2).Or("from_id = ? AND to_id = ?", u2, u1).First(&fr); tx.Error != nil {
		if tx.Error == gorm.ErrRecordNotFound {
			return nil, ErrNoFriendRequestFound
		}
		return nil, tx.Error
	}

	return &fr, nil
}
