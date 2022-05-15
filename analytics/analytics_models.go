// analytics_models.go:
// While normally

package main

import (
	"encoding/json"
	"github.com/mitchellh/mapstructure"
	"gorm.io/gorm"
)

// EventSource indicates what endpoint an event was ingested from; The website and game use separate endpoints to ingest events.
type EventSource string

const (
	EventSourceWeb  EventSource = "web"
	EventSourceGame EventSource = "game"
)

type EventType string

const (
	EventTypeLoginLoginSuccess EventType = "Login_LoginSuccess"
	EventTypeLoginLoginFail    EventType = "Login_LoginFail"

	EventTypeAdminAppOpen  EventType = "Admin_AppOpen"
	EventTypeAdminAppClose EventType = "Admin_AppClose"
	EventTypeAdminAppError EventType = "Admin_AppError"

	EventTypeMainMenuAvatarsClickGetMoreFavorites        EventType = "MainMenu_Avatars_ClickGetMoreFavorites"
	EventTypeMainMenuUserDetailsClickVRChatPlusSupporter EventType = "MainMenu_UserDetails_ClickVRChatPlusSupporter"
	EventTypeMainMenuClickVRChatPlus                     EventType = "MainMenu_ClickVRChatPlus"
	EventTypeMainMenuVRChatPlusMoreInfo                  EventType = "MainMenu_VRChatPlus_MoreInfo"
	EventTypeMainMenuVRChatPlusClickMonthly              EventType = "MainMenu_VRChatPlus_ClickMonthly"
	EventTypeMainMenuVRChatPlusClickYearly               EventType = "MainMenu_VRChatPlus_ClickYearly"

	EventTypeSafetyChangeSafetyLevel                  EventType = "Safety_ChangeSafetyLevel"
	EventTypeSafetyPanicModeActivated                 EventType = "Safety_PanicModeActivated"
	EventTypeSafetyChangeAvatarPerfMinRatingToDisplay EventType = "Safety_ChangeAvatarPerfMinRatingToDisplay"

	EventTypeModerationShowUserAvatar EventType = "Moderation_ShowUserAvatar"
	EventTypeModerationHideUserAvatar EventType = "Moderation_HideUserAvatar"
	EventTypeModerationSendWarning    EventType = "Moderation_SendWarning"
	EventTypeModerationSendKick       EventType = "Moderation_SendKick"
	EventTypeModerationMuteUser       EventType = "Moderation_MuteUser"

	EventTypeSocialSendFriendRequest EventType = "Social_SendFriendRequest"
	EventTypeSocialUpdateStatus      EventType = "Social_UpdateStatus"
	EventTypeSocialUpdateBio         EventType = "Social_UpdateBio"

	EventTypeWorldEnterWorld EventType = "World_EnterWorld"

	EventTypeErrorLoadWorldFailed EventType = "Error_LoadWorldFailed"

	EventTypeSearchManualSearch      EventType = "Search_ManualSearch"
	EventTypeSearchAddSavedSearch    EventType = "Search_AddSavedSearch"
	EventTypeSearchRemoveSavedSearch EventType = "Search_RemoveSavedSearch"

	EventTypeSubProfileImageChangedClient EventType = "Sub_ProfileImageChangedClient"

	EventTypeMenuUserDetailsClickPlaylist EventType = "Menu_userDetailsClickPlaylist"

	EventTypeDialogMenuGiftVRChatPlusBuy EventType = "DialogMenu_GiftVRChatPlus_Buy"
)

// Event is the database struct.
type Event struct {
	gorm.Model
	BatchId string
	ApiKey  string
	Type    EventSource
	Data    string `gorm:"type:jsonb"`
}

func (e *Event) ToApi() *ApiAnalyticsEvent {
	var j map[string]interface{}
	var uid string
	var up ApiAnalyticsEventUserProperties

	err := json.Unmarshal([]byte(e.Data), &j)
	if err != nil {
		panic(err)
	}

	if _uid, ok := j["user_id"].(string); ok {
		uid = _uid
	}

	if u, ok := j["user_properties"]; ok {
		if _u, ok := u.(map[string]interface{}); ok {
			err := mapstructure.Decode(_u, &up)
			if err != nil {
				panic(err)
			}
		} else {
			panic(err) // invalid data?
		}
	}

	return &ApiAnalyticsEvent{
		Model: gorm.Model{
			ID:        e.ID,
			CreatedAt: e.CreatedAt,
			UpdatedAt: e.UpdatedAt,
			DeletedAt: e.DeletedAt,
		},
		SessionId:           int(j["session_id"].(float64)),
		UserId:              uid,
		EventSource:         e.Type,
		EventType:           "",
		EventUserProperties: up,
		EventData:           nil,
	}
}

type ApiAnalyticsEvent struct {
	gorm.Model
	SessionId           int
	UserId              string
	EventSource         EventSource
	EventType           EventType
	EventUserProperties ApiAnalyticsEventUserProperties
	EventData           map[string]interface{}
}

type ApiAnalyticsEventUserProperties struct {
	// General analytics
	Betas     []string `json:"betas"`
	UserIcons int      `json:"userIcons"`

	InVRMode     bool   `json:"inVRMode"`
	Platform     string `json:"platform"`
	InputType    string `json:"inputType"`
	Store        string `json:"store"`
	BuildType    string `json:"buildType"`
	BuildVersion string `json:"buildVersion"`

	// User-related (
	AcceptedTOSVersion int    `json:"acceptedTOSVersion"`
	AccountType        string `json:"accountType"`
	DisplayName        string `json:"displayName"`
	DeveloperType      string `json:"developerType"`
	SubscriptionType   string `json:"subscriptionType"`
	SubscriptionStore  string `json:"subscriptionStore"`

	FavoriteWorlds  int `json:"favoriteWorlds"`
	FavoriteAvatars int `json:"favoriteAvatars"`
	FavoriteFriends int `json:"favoriteFriends"`
	NumberOfFriends int `json:"numberOfFriends"`

	CurrentWorldId   string `json:"currentWorldId"`
	CurrentWorldName string `json:"currentWorldName"`

	DeviceID           string `json:"deviceId"`
	OperatingSystem    string `json:"operatingSystem"`
	ProcessorFrequency int    `json:"processorFrequency"`
	SystemMemorySize   int    `json:"systemMemorySize"`

	GraphicsDeviceName    string `json:"graphicsDeviceName"`
	GraphicsDeviceVendor  string `json:"graphicsDeviceVendor"`
	GraphicsDeviceVersion string `json:"graphicsDeviceVersion"`

	VRDeviceModel       string `json:"vrDeviceModel"`
	VRDeviceRefreshRate int    `json:"vrDeviceRefreshRate"`
	VRDeviceRenderScale int    `json:"vrDeviceRenderScale"`
	NumExtraVRTrackers  int    `json:"numExtraVRTrackers"`

	// Game Settings
	HudEnabled                                     bool   `json:"hudEnabled"`
	SafetyLevel                                    int    `json:"safetyLevel"`
	NameplatesEnabled                              bool   `json:"nameplatesEnabled"`
	SettingMicLevel                                int    `json:"setting_MicLevel"`
	SettingToggleTalk                              bool   `json:"setting_ToggleTalk"`
	SettingSafetyLevel                             int    `json:"setting_SafetyLevel"`
	SettingHeadLookWalk                            bool   `json:"setting_HeadLookWalk"`
	SettingShowTooltips                            bool   `json:"setting_ShowTooltips"`
	SettingViveAdvanced                            bool   `json:"setting_ViveAdvanced"`
	SettingInvertedMouse                           bool   `json:"setting_InvertedMouse"`
	SettingMicDeviceName                           string `json:"setting_MicDeviceName"`
	SettingPersonalSpace                           bool   `json:"setting_PersonalSpace"`
	SettingTalkDefaultOn                           bool   `json:"setting_TalkDefaultOn"`
	SettingComfortTurning                          bool   `json:"setting_ComfortTurning"`
	SettingDesktopReticle                          bool   `json:"setting_DesktopReticle"`
	SettingShowSocialRank                          bool   `json:"setting_ShowSocialRank"`
	SettingGraphicsQuality                         string `json:"setting_GraphicsQuality"`
	SettingDisableMicButton                        bool   `json:"setting_DisableMicButton"`
	SettingLocomotionMethod                        string `json:"setting_LocomotionMethod"`
	SettingMouseSensitivity                        int    `json:"setting_MouseSensitivity"`
	SettingUIHapticsEnabled                        bool   `json:"setting_UIHapticsEnabled"`
	SettingAllowUntrustedURL                       bool   `json:"setting_AllowUntrustedURL"`
	SettingClearCacheOnStart                       bool   `json:"setting_ClearCacheOnStart"`
	SettingShowCommunityLabs                       bool   `json:"setting_ShowCommunityLabs"`
	SettingAllowAvatarCopying                      bool   `json:"setting_AllowAvatarCopying"`
	SettingSkipGoButtonInLoad                      bool   `json:"setting_SkipGoButtonInLoad"`
	SettingStreamerModeEnabled                     bool   `json:"setting_StreamerModeEnabled"`
	SettingThirdPersonRotation                     bool   `json:"setting_ThirdPersonRotation"`
	SettingVoicePrioritization                     bool   `json:"setting_VoicePrioritization"`
	AvatarPerfMinRatingToDisplay                   int    `json:"avatarPerfMinRatingToDisplay"`
	SettingLimitDynamicBoneUsage                   bool   `json:"setting_LimitDynamicBoneUsage"`
	SettingSelectedNetworkRegion                   string `json:"setting_SelectedNetworkRegion"`
	SettingHideNotificationPhotos                  bool   `json:"setting_HideNotificationPhotos"`
	SettingMaximumAvatarDownloadSize               int    `json:"setting_MaximumAvatarDownloadSize"`
	SettingAdvancedGraphicsAntialiasing            int    `json:"setting_AdvancedGraphicsAntialiasing"`
	SettingShowCommunityLabsInWorldSearch          bool   `json:"setting_ShowCommunityLabsInWorldSearch"`
	SettingAvatarPerformanceRatingMinimumToDisplay int    `json:"setting_AvatarPerformanceRatingMinimumToDisplay"`
}
