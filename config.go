package main

import (
	"context"
	"encoding/json"
	"fmt"
	hsync "github.com/gtsatsis/harvester/sync"
	"sync"
)

// Config holds the configuration for database & redis connections, fiber, and the harvester polling interval.
type Config struct {
	Database               DBConfig     `json:"database"`
	Redis                  RedisConfig  `json:"redis"`
	Server                 ServerConfig `json:"server"`
	ApiConfigRefreshRateMs int          `json:"apiConfigRefreshRateMs"`
}

// DBConfig holds the configuration for the database.
type DBConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"username"`
	Password string `json:"password"`
	Database string `json:"database"`
}

// RedisConfig holds the configuration for the redis connection.
type RedisConfig struct {
	Host     string `json:"host"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

// ServerConfig holds the configuration for Fiber.
type ServerConfig struct {
	Address string `json:"address"`
	Port    int    `json:"port"`
}

// ApiConfig holds the dynamic configuration for the API.
type ApiConfig struct {
	// Internal Configuration
	InfoPushes   ApiInfoPushesList `seed:"[]" json:"infoPushes" redis:"{config}:infoPushes"`
	JwtSecret    hsync.Secret      `json:"-" seed:"INSECURE_CHANGEME" redis:"{config}:jwtSecret"`
	PhotonSecret hsync.Secret      `json:"-" seed:"INSECURE_CHANGEME" redis:"{config}:photonSecret"`
	// Photon Room Settings
	PhotonSettingMaxAccountsPerIpAddress hsync.Int64 `seed:"5" json:"maxAccountsPerIp" redis:"{config}:photonSettingMaxAccountsPerIp"`
	// Connector Mod AutoConfig Functionality
	AutoConfigApiUrl         hsync.String `json:"autoConfigApiUrl" seed:"" redis:"{config}:autoConfigApiUrl"`
	AutoConfigWebsocketUrl   hsync.String `json:"autoConfigWebsocketUrl" seed:"" redis:"{config}:autoConfigWebsocketUrl"`
	AutoConfigNameServerHost hsync.String `json:"autoConfigNameServerHost" seed:"" redis:"{config}:autoConfigNameServerHost"`
	// External Configuration (VRChat-specifics)
	Address                       hsync.String           `seed:"" json:"address" redis:"{config}:address"`                                                  // Address is the physical address of the corporate entity.
	Announcements                 ApiAnnouncementsList   `seed:"[]" json:"announcements" redis:"{config}:announcements"`                                    // Announcements is a list of announcements to be displayed to the user upon world load.
	ApiKey                        hsync.String           `seed:"" json:"apiKey" redis:"{config}:apiKey"`                                                    // ApiKey is the API key used to authenticate requests.
	AppName                       hsync.String           `seed:"" json:"appName" redis:"{config}:appName"`                                                  // AppName is the name of the application.
	BuildVersionTag               hsync.String           `seed:"" json:"buildVersionTag" env:"BUILD_VERSION_TAG"`                                           // BuildVersionTag is the tag used to identify which API build is currently running.
	CaptchaPercentage             hsync.Int64            `seed:"0" json:"captchaPercentage" redis:"{config}:captchaPercentage"`                             // CaptchaPercentage is the percentage of suspicious world joins that will be required to pass a captcha.
	ClientApiKey                  hsync.String           `seed:"" json:"clientApiKey" redis:"{config}:clientApiKey"`                                        // ClientApiKey is the API key used to authenticate requests from the client. Should be the same as ApiKey.
	ClientBPSCeiling              hsync.Int64            `seed:"18432" json:"clientBPSCeiling" redis:"{config}:clientBPSCeiling"`                           // ClientBPSCeiling is direct client configuration.
	ClientDisconnectTimeout       hsync.Int64            `seed:"30000" json:"clientDisconnectTimeout" redis:"{config}:clientDisconnectTimeout"`             // ClientDisconnectTimeout is direct client configuration
	ClientReservedPlayerBPS       hsync.Int64            `seed:"7168" json:"clientReservedPlayerBPS" redis:"{config}:clientReservedPlayerBPS"`              // ClientReservedPlayerBPS is direct client configuration
	ClientSentCountAllowance      hsync.Int64            `seed:"15" json:"clientSentCountAllowance" redis:"{config}:clientSentCountAllowance"`              // ClientSentCountAllowance is direct client configuration
	ContactEmail                  hsync.String           `seed:"" json:"contactEmail" redis:"{config}:contactEmail"`                                        // ContactEmail is the email address to be used for contact requests.
	CopyrightEmail                hsync.String           `seed:"" json:"copyrightEmail" redis:"{config}:copyrightEmail"`                                    // CopyrightEmail is the email address to be used for copyright requests.
	CurrentTOSVersion             hsync.Int64            `seed:"1" json:"currentTOSVersion" redis:"{config}:currentTOSVersion"`                             // CurrentTOSVersion is the current version of the Terms of Service.
	DefaultAvatar                 hsync.String           `seed:"" json:"defaultAvatar" redis:"{config}:defaultAvatar"`                                      // DefaultAvatar is the default avatar id to be used for new users.
	DeploymentGroup               hsync.String           `seed:"" json:"deploymentGroup" env:"DEPLOYMENT_GROUP"`                                            // DeploymentGroup is the name of the deployment group. (blue/green)
	DevAppVersionStandalone       hsync.String           `seed:"" json:"devAppVersionStandalone" redis:"{config}:devAppVersionStandalone"`                  // DevAppVersionStandalone is not used.
	DevDownloadLinkWindows        hsync.String           `seed:"" json:"devDownloadLinkWindows" redis:"{config}:devDownloadLinkWindows"`                    // DevDownloadLinkWindows is not used.
	DevSdkUrl                     hsync.String           `seed:"" json:"devSdkUrl" redis:"{config}:devSdkUrl"`                                              // DevSdkUrl is not used.
	DevServerVersionStandalone    hsync.String           `seed:"" json:"devServerVersionStandalone" redis:"{config}:devServerVersionStandalone"`            // DevServerVersionStandalone is not used.
	DisCountdown                  hsync.String           `seed:"" json:"disCountdown" redis:"{config}:disCountdown"`                                        // DisCountdown not used.
	DisableAvatarCopying          hsync.Bool             `seed:"false" json:"disableAvatarCopying" redis:"{config}:disableAvatarCopying"`                   // DisableAvatarCopying is a flag indicating whether to allow avatar copying.
	DisableAvatarGating           hsync.Bool             `seed:"false" json:"disableAvatarGating" redis:"{config}:disableAvatarGating"`                     // DisableAvatarGating is a flag indicating whether gate avatar uploads.
	DisableCaptcha                hsync.Bool             `seed:"false" json:"disableCaptcha" redis:"{config}:disableCaptcha"`                               // DisableCaptcha is a flag indicating whether to use captcha for world joins.
	DisableCommunityLabs          hsync.Bool             `seed:"false" json:"disableCommunityLabs" redis:"{config}:disableCommunityLabs"`                   // DisableCommunityLabs is a flag indicating whether to disable the Community Labs feature.
	DisableCommunityLabsPromotion hsync.Bool             `seed:"false" json:"disableCommunityLabsPromotion" redis:"{config}:disableCommunityLabsPromotion"` // DisableCommunityLabsPromotion is a flag indicating whether to disable promotion *out* of the Community Labs.
	DisableEmail                  hsync.Bool             `seed:"false" json:"disableEmail" redis:"{config}:disableEmail"`                                   // DisableEmail is a flag indicating whether to disable email sending.
	DisableEventStream            hsync.Bool             `seed:"false" json:"disableEventStream" redis:"{config}:disableEventStream"`                       // DisableEventStream is a flag indicating whether to disable the event stream. (Pipeline?)
	DisableFeedbackGating         hsync.Bool             `seed:"false" json:"disableFeedbackGating" redis:"{config}:disableFeedbackGating"`                 // DisableFeedbackGating is a flag indicating whether to gate feedback requests.
	DisableFrontendBuilds         hsync.Bool             `seed:"false" json:"disableFrontendBuilds" redis:"{config}:disableFrontendBuilds"`                 // DisableFrontendBuilds is a flag indicating whether to disable frontend builds.
	DisableHello                  hsync.Bool             `seed:"false" json:"disableHello" redis:"{config}:disableHello"`                                   // DisableHello is a flag indicating whether to disable an unknown feature.
	DisableOculusSubs             hsync.Bool             `seed:"true" json:"disableOculusSubs"`                                                             // DisableOculusSubs is a flag indicating whether to disable the Oculus subscriptions.
	DisableRegistration           hsync.Bool             `seed:"false" json:"disableRegistration" redis:"{config}:disableRegistration"`                     // DisableRegistration is a flag indicating whether to disable registration.
	DisableSteamNetworking        hsync.Bool             `seed:"true" json:"disableSteamNetworking"`                                                        // DisableSteamNetworking is a flag indicating whether to disable Steam networking.
	DisableTwoFactorAuth          hsync.Bool             `seed:"false" json:"disableTwoFactorAuth" redis:"{config}:disableTwoFactorAuth"`                   // DisableTwoFactorAuth is a flag indicating whether to disable two-factor authentication.
	DisableUdon                   hsync.Bool             `seed:"false" json:"disableUdon" redis:"{config}:disableUdon"`                                     // DisableUdon is a flag indicating whether to disable Udon.
	DisableUpgradeAccount         hsync.Bool             `seed:"true" json:"disableUpgradeAccount"`                                                         // DisableUpgradeAccount is a flag indicating whether to disable the account upgrade feature
	DownloadUrls                  ApiDownloadUrls        `seed:"{}" json:"downloadUrls" redis:"{config}:downloadUrls"`                                      // DownloadUrls is a map of SDK download urls.
	DynamicWorldRows              ApiDynamicWorldRowList `seed:"[]" json:"dynamicWorldRows" redis:"{config}:dynamicWorldRows"`                              // DynamicWorldRows is a list of dynamic world rows.
	// Events is a struct containing client event configuration data. (Comment is here for cleanliness of other comments)
	Events                                    ApiEvents               `seed:"{\"distanceClose\":2,\"distanceFactor\":100,\"distanceFar\":80,\"groupDistance\":3,\"maximumBunchSize\":247,\"notVisibleFactor\":100,\"playerOrderBucketSize\":5,\"playerOrderFactor\":55,\"slowUpdateFactorThreshold\":25,\"viewSegmentLength\":5}" json:"events" redis:"{config}:events"`
	FrontendBuildBranch                       hsync.String            `seed:"main" redis:"{config}:frontendBuildBranch"`
	GearDemoRoomId                            hsync.String            `seed:"0" json:"gearDemoRoomId"`                                                                                          // GearDemoRoomId is not used.
	HomeWorldId                               hsync.String            `seed:"" json:"homeWorldId" redis:"{config}:homeWorldId"`                                                                 // HomeWorldId is the id of the default home world.
	HomepageRedirectTarget                    hsync.String            `seed:"" json:"homepageRedirectTarget" redis:"{config}:homepageRedirectTarget"`                                           // HomepageRedirectTarget is the target of the homepage redirect.
	HubWorldId                                hsync.String            `seed:"" json:"hubWorldId"`                                                                                               // HubWorldId is the id of the default hub world. (not used anymore)
	JobsEmail                                 hsync.String            `seed:"" json:"jobsEmail" redis:"{config}:jobsEmail"`                                                                     // JobsEmail is the email address for the jobs email.
	MessageOfTheDay                           hsync.String            `seed:"" json:"messageOfTheDay" redis:"{config}:messageOfTheDay"`                                                         // MessageOfTheDay is the message of the day.
	ModerationEmail                           hsync.String            `seed:"" json:"moderationEmail" redis:"{config}:moderationEmail"`                                                         // ModerationEmail is the email address for the moderation email.
	ModerationQueryPeriod                     hsync.Int64             `seed:"60" json:"moderationQueryPeriod" redis:"{config}:moderationQueryPeriod"`                                           // ModerationQueryPeriod is the period in seconds for querying moderation data.
	NotAllowedToSelectAvatarInPrivateWorldMsg hsync.String            `seed:"" json:"notAllowedToSelectAvatarInPrivateWorldMessage" redis:"{config}:notAllowedToSelectAvatarInPrivateWorldMsg"` // NotAllowedToSelectAvatarInPrivateWorldMsg is the message to display when a user tries to select an avatar in a private world but their rank does not allow them to do so.
	Plugin                                    hsync.String            `seed:"naoka" json:"plugin" redis:"{config}:plugin"`                                                                      // Plugin is the name of the Photon plugin.
	ReleaseAppVersionStandalone               hsync.String            `seed:"" json:"releaseAppVersionStandalone" redis:"{config}:releaseAppVersionStandalone"`                                 // ReleaseAppVersionStandalone is the version of the standalone client.
	ReleaseSdkUrl                             hsync.String            `seed:"" json:"releaseSdkUrl" redis:"{config}:releaseSdkUrl"`                                                             // ReleaseSdkUrl is the url of the release SDK.
	ReleaseSdkVersion                         hsync.String            `seed:"" json:"releaseSdkVersion" redis:"{config}:releaseSdkVersion"`                                                     // ReleaseSdkVersion is the version of the release SDK.
	ReleaseServerVersionStandalone            hsync.String            `seed:"private_server_01" json:"releaseServerVersionStandalone"`                                                          // ReleaseServerVersionStandalone is not used.
	SdkDeveloperFaqUrl                        hsync.String            `seed:"" json:"sdkDeveloperFaqUrl" redis:"{config}:sdkDeveloperFaqUrl"`                                                   // SdkDeveloperFaqUrl is the url of the SDK developer faq.
	SdkDiscordUrl                             hsync.String            `seed:"" json:"sdkDiscordUrl" redis:"{config}:sdkDiscordUrl"`                                                             // SdkDiscordUrl is the url of the SDK discord.
	SdkNotAllowedToPublishMsg                 hsync.String            `seed:"" json:"sdkNotAllowedToPublishMessage" redis:"{config}:sdkNotAllowedToPublishMsg"`                                 // SdkNotAllowedToPublishMsg is the message to display when a user tries to publish content but their rank does not allow them to do so.
	SdkUnityVersion                           hsync.String            `seed:"" json:"sdkUnityVersion" redis:"{config}:sdkUnityVersion"`                                                         // SdkUnityVersion is the version of Unity supported by the SDK.
	ServerName                                hsync.String            `seed:"" json:"serverName" env:"SERVER_NAME"`                                                                             // ServerName is the name of the current API instance.
	SupportEmail                              hsync.String            `seed:"" json:"supportEmail" redis:"{config}:supportEmail"`                                                               // SupportEmail is the email address for the support email.
	TimeoutWorldId                            hsync.String            `seed:"" json:"timeOutWorldId" redis:"{config}:timeOutWorldId"`                                                           // TimeoutWorldId is the id of the timeout world.
	TutorialWorldId                           hsync.String            `seed:"" json:"tutorialWorldId" redis:"{config}:tutorialWorldId"`                                                         // TutorialWorldId is the id of the tutorial world.
	UpdateRateMsMaximum                       hsync.Int64             `seed:"1000" json:"updateRateMsMaximum" redis:"{config}:updateRateMsMaximum"`                                             // UpdateRateMsMaximum is direct client configuration.
	UpdateRateMsMinimum                       hsync.Int64             `seed:"50" json:"updateRateMsMinimum" redis:"{config}:updateRateMsMinimum"`                                               // UpdateRateMsMinimum is direct client configuration.
	UpdateRateMsNormal                        hsync.Int64             `seed:"200" json:"updateRateMsNormal" redis:"{config}:updateRateMsNormal"`                                                // UpdateRateMsNormal is direct client configuration.
	UpdateRateMsUdonManual                    hsync.Int64             `seed:"50" json:"updateRateMsUdonManual" redis:"{config}:updateRateMsUdonManual"`                                         // UpdateRateMsUdonManual is direct client configuration.
	UploadAnalysisPercent                     hsync.Int64             `seed:"0" json:"uploadAnalysisPercent"`                                                                                   // UploadAnalysisPercent is not used.
	UrlList                                   UrlList                 `seed:"[]" json:"urlList" redis:"{config}:urlList"`                                                                       // UrlList is a whitelist of URLs that can be accessed by video players from within the client with "Allow Untrusted URLs" off.
	UseReliableUdpForVoice                    hsync.Bool              `seed:"false" json:"useReliableUdpForVoice" redis:"{config}:useReliableUdpForVoice"`                                      // UseReliableUdpForVoice is whether to use reliable UDP for voice.
	UserUpdatePeriod                          hsync.Int64             `seed:"60" json:"userUpdatePeriod" redis:"{config}:userUpdatePeriod"`                                                     // UserUpdatePeriod ???
	UserVerificationDelay                     hsync.Int64             `seed:"5" json:"userVerificationDelay" redis:"{config}:userVerificationDelay"`                                            // UserVerificationDelay ???
	UserVerificationRetry                     hsync.Int64             `seed:"30" json:"userVerificationRetry" redis:"{config}:userVerificationRetry"`                                           // UserVerificationRetry ???
	UserVerificationTimeout                   hsync.Int64             `seed:"60" json:"userVerificationTimeout" redis:"{config}:userVerificationTimeout"`                                       // UserVerificationTimeout ???
	ViveWindowsUrl                            hsync.String            `seed:"" json:"viveWindowsUrl"`                                                                                           // ViveWindowsUrl is the url of the Vive Windows Client.
	WhitelistedAssetUrls                      WhitelistedAssetUrlList `seed:"[]" json:"whiteListedAssetUrls" redis:"{config}:whiteListedAssetUrls"`                                             // WhitelistedAssetUrls is a whitelist of URLs that the client can retrieve assets from.
	WorldUpdatePeriod                         hsync.Int64             `seed:"60" json:"worldUpdatePeriod" redis:"{config}:worldUpdatePeriod"`                                                   // WorldUpdatePeriod ???
	YoutubeDLHash                             hsync.String            `seed:"" json:"youtubedl-hash" redis:"{config}:youtubedl-hash"`                                                           // YoutubeDLHash is the hash of the youtube-dl binary.
	YoutubeDLVersion                          hsync.String            `seed:"" json:"youtubedl-version" redis:"{config}:youtubedl-version"`                                                     // YoutubeDLVersion is the version of youtube-dl.
}

func (a *ApiConfig) Update(u map[string]interface{}) error {
	p := RedisClient.Pipeline()

	for key, value := range u {
		p.Set(context.Background(), fmt.Sprintf("{config}:%s", key), value, 0)
	}

	_, err := p.Exec(context.Background())
	if err != nil {
		return err
	}
	return nil
}

// ApiAnnouncementsList is a list of ApiAnnouncement
type ApiAnnouncementsList struct {
	m    sync.RWMutex
	List []ApiAnnouncement
}

// SetString sets the value of ApiAnnouncementsList based on a JSON string.
func (a *ApiAnnouncementsList) SetString(s string) error {
	a.m.Lock()
	defer a.m.Unlock()
	return json.Unmarshal([]byte(s), &a.List)
}

// Get returns a list of ApiAnnouncement
func (a *ApiAnnouncementsList) Get() []ApiAnnouncement {
	a.m.RLock()
	defer a.m.RUnlock()
	return a.List
}

// String returns a stringified version of the list.
func (a *ApiAnnouncementsList) String() string {
	a.m.RLock()
	defer a.m.RUnlock()
	b, _ := json.Marshal(a.List)
	return string(b)
}

// ApiAnnouncement is an announcement for the API.
type ApiAnnouncement struct {
	Name string `json:"name"` // Name is the name of the announcement.
	Text string `json:"text"` // Text is the text of the announcement.
}

// ApiDownloadUrls is a struct that contains the download URLs for the SDKs.
type ApiDownloadUrls struct {
	m           sync.RWMutex
	Sdk2        string `json:"sdk2"`         // Sdk2 is a URL pointing to the latest release of SDKv2.
	Sdk3Avatars string `json:"sdk3-avatars"` // Sdk3Avatars is a URL pointing to the latest release of SDKv3 for avatars.
	Sdk3Worlds  string `json:"sdk3-worlds"`  // Sdk3Worlds is a URL pointing to the latest release of SDKv3 for worlds.
}

// ApiDownloadUrlsResponse is a struct used to convert the ApiDownloadUrls struct to JSON.
type ApiDownloadUrlsResponse struct {
	Sdk2        string `json:"sdk2"`         // Sdk2 is a URL pointing to the latest release of SDKv2.
	Sdk3Avatars string `json:"sdk3-avatars"` // Sdk3Avatars is a URL pointing to the latest release of SDKv3 for avatars.
	Sdk3Worlds  string `json:"sdk3-worlds"`  // Sdk3Worlds is a URL pointing to the latest release of SDKv3 for worlds.
}

// SetString sets the value of ApiDownloadUrls based on a JSON string.
func (a *ApiDownloadUrls) SetString(s string) error {
	a.m.Lock()
	defer a.m.Unlock()
	return json.Unmarshal([]byte(s), a)
}

// Get returns the ApiDownloadUrls struct as ApiDownloadUrlsResponse.
func (a *ApiDownloadUrls) Get() ApiDownloadUrlsResponse {
	return ApiDownloadUrlsResponse{
		Sdk2:        a.Sdk2,
		Sdk3Avatars: a.Sdk3Avatars,
		Sdk3Worlds:  a.Sdk3Worlds,
	}
}

// String returns a stringified version of the struct.
func (a *ApiDownloadUrls) String() string {
	a.m.RLock()
	defer a.m.RUnlock()
	b, _ := json.Marshal(a)
	return string(b)
}

type ApiDynamicWorldRowList struct {
	m    sync.RWMutex
	List []ApiDynamicWorldRow
}

func (a *ApiDynamicWorldRowList) SetString(s string) error {
	a.m.Lock()
	defer a.m.Unlock()
	return json.Unmarshal([]byte(s), &a.List)
}

func (a *ApiDynamicWorldRowList) Get() []ApiDynamicWorldRow {
	a.m.RLock()
	defer a.m.RUnlock()
	return a.List
}

func (a *ApiDynamicWorldRowList) String() string {
	a.m.RLock()
	defer a.m.RUnlock()
	b, _ := json.Marshal(a)
	return string(b)
}

type ApiDynamicWorldRow struct {
	Index         int    `json:"index"`         // Index is the index of the row.
	Name          string `json:"name"`          // Name is the name of the row.
	Platform      string `json:"platform"`      // Platform is the platform of the row.
	SortHeading   string `json:"sortHeading"`   // SortHeading ???
	SortOrder     string `json:"sortOrder"`     // SortOrder ???
	SortOwnership string `json:"sortOwnership"` // SortOwnership ???
}

type ApiEvents struct {
	m                         sync.RWMutex
	DistanceClose             int64 `json:"distanceClose"`
	DistanceFactor            int64 `json:"distanceFactor"`
	DistanceFar               int64 `json:"distanceFar"`
	GroupDistance             int64 `json:"groupDistance"`
	MaximumBunchSize          int64 `json:"maximumBunchSize"`
	NotVisibleFactor          int64 `json:"notVisibleFactor"`
	PlayerOrderBucketSize     int64 `json:"playerOrderBucketSize"`
	PlayerOrderFactor         int64 `json:"playerOrderFactor"`
	SlowUpdateFactorThreshold int64 `json:"slowUpdateFactorThreshold"`
	ViewSegmentLength         int64 `json:"viewSegmentLength"`
}

type ApiEventsResponse struct {
	DistanceClose             int64 `json:"distanceClose"`
	DistanceFactor            int64 `json:"distanceFactor"`
	DistanceFar               int64 `json:"distanceFar"`
	GroupDistance             int64 `json:"groupDistance"`
	MaximumBunchSize          int64 `json:"maximumBunchSize"`
	NotVisibleFactor          int64 `json:"notVisibleFactor"`
	PlayerOrderBucketSize     int64 `json:"playerOrderBucketSize"`
	PlayerOrderFactor         int64 `json:"playerOrderFactor"`
	SlowUpdateFactorThreshold int64 `json:"slowUpdateFactorThreshold"`
	ViewSegmentLength         int64 `json:"viewSegmentLength"`
}

func (a *ApiEvents) SetString(s string) error {
	a.m.Lock()
	defer a.m.Unlock()
	return json.Unmarshal([]byte(s), a)
}

func (a *ApiEvents) Get() ApiEventsResponse {
	return ApiEventsResponse{
		DistanceClose:             a.DistanceClose,
		DistanceFactor:            a.DistanceFactor,
		DistanceFar:               a.DistanceFar,
		GroupDistance:             a.GroupDistance,
		MaximumBunchSize:          a.MaximumBunchSize,
		NotVisibleFactor:          a.NotVisibleFactor,
		PlayerOrderBucketSize:     a.PlayerOrderBucketSize,
		PlayerOrderFactor:         a.PlayerOrderFactor,
		SlowUpdateFactorThreshold: a.SlowUpdateFactorThreshold,
		ViewSegmentLength:         a.ViewSegmentLength,
	}
}

func (a *ApiEvents) String() string {
	a.m.RLock()
	defer a.m.RUnlock()
	b, _ := json.Marshal(a)
	return string(b)
}

type UrlList struct {
	m    sync.RWMutex
	List []string
}

func (u *UrlList) SetString(s string) error {
	u.m.Lock()
	defer u.m.Unlock()
	return json.Unmarshal([]byte(s), &u.List)
}

func (u *UrlList) Get() []string {
	u.m.RLock()
	defer u.m.RUnlock()
	return u.List
}

func (u *UrlList) String() string {
	u.m.RLock()
	defer u.m.RUnlock()
	b, _ := json.Marshal(u)
	return string(b)
}

type WhitelistedAssetUrlList struct {
	m    sync.RWMutex
	List []string
}

func (w *WhitelistedAssetUrlList) SetString(s string) error {
	w.m.Lock()
	defer w.m.Unlock()
	return json.Unmarshal([]byte(s), &w.List)
}

func (w *WhitelistedAssetUrlList) Get() []string {
	w.m.RLock()
	defer w.m.RUnlock()
	return w.List
}

func (w *WhitelistedAssetUrlList) String() string {
	w.m.RLock()
	defer w.m.RUnlock()
	b, _ := json.Marshal(w)
	return string(b)
}

type ApiInfoPushesList struct {
	m    sync.RWMutex
	List []ApiInfoPush
}

func (w *ApiInfoPushesList) SetString(s string) error {
	w.m.Lock()
	defer w.m.Unlock()
	return json.Unmarshal([]byte(s), &w.List)
}

func (w *ApiInfoPushesList) Get() []ApiInfoPush {
	w.m.RLock()
	defer w.m.RUnlock()
	return w.List
}

func (w *ApiInfoPushesList) String() string {
	w.m.RLock()
	defer w.m.RUnlock()
	b, _ := json.Marshal(w)
	return string(b)
}

type ApiInfoPush struct {
	Id            string                 `json:"id"`
	IsEnabled     bool                   `json:"isEnabled"`
	ReleaseStatus string                 `json:"releaseStatus"`
	Priority      int                    `json:"priority"`
	Tags          []string               `json:"tags"`
	Data          map[string]interface{} `json:"data"`
	Hash          string                 `json:"hash"`
	CreatedAt     string                 `json:"createdAt"`
	UpdatedAt     string                 `json:"updatedAt"`
}

// ApiConfigResponse is the response from the /config endpoint. It contains public values from ApiConfig with native types.
type ApiConfigResponse struct {
	Address                                   string                  `json:"address"`                                       // Address is the physical address of the corporate entity.
	Announcements                             []ApiAnnouncement       `json:"announcements"`                                 // Announcements is a list of announcements to be displayed to the user upon world load.
	ApiKey                                    string                  `json:"apiKey"`                                        // ApiKey is the API key used to authenticate requests.
	AppName                                   string                  `json:"appName"`                                       // AppName is the name of the application.
	BuildVersionTag                           string                  `json:"buildVersionTag"`                               // BuildVersionTag is the tag used to identify which API build is currently running.
	CaptchaPercentage                         int64                   `json:"captchaPercentage"`                             // CaptchaPercentage is the percentage of suspicious world joins that will be required to pass a captcha.
	ClientApiKey                              string                  `json:"clientApiKey"`                                  // ClientApiKey is the API key used to authenticate requests from the client. Should be the same as ApiKey.
	ClientBPSCeiling                          int64                   `json:"clientBPSCeiling"`                              // ClientBPSCeiling is direct client configuration.
	ClientDisconnectTimeout                   int64                   `json:"clientDisconnectTimeout"`                       // ClientDisconnectTimeout is direct client configuration
	ClientReservedPlayerBPS                   int64                   `json:"clientReservedPlayerBPS"`                       // ClientReservedPlayerBPS is direct client configuration
	ClientSentCountAllowance                  int64                   `json:"clientSentCountAllowance"`                      // ClientSentCountAllowance is direct client configuration
	ContactEmail                              string                  `json:"contactEmail"`                                  // ContactEmail is the email address to be used for contact requests.
	CopyrightEmail                            string                  `json:"copyrightEmail"`                                // CopyrightEmail is the email address to be used for copyright requests.
	CurrentTOSVersion                         int64                   `json:"currentTOSVersion"`                             // CurrentTOSVersion is the current version of the Terms of Service.
	DefaultAvatar                             string                  `json:"defaultAvatar"`                                 // DefaultAvatar is the default avatar id to be used for new users.
	DeploymentGroup                           string                  `json:"deploymentGroup"`                               // DeploymentGroup is the name of the deployment group. (blue/green)
	DevAppVersionStandalone                   string                  `json:"devAppVersionStandalone"`                       // DevAppVersionStandalone is not used.
	DevDownloadLinkWindows                    string                  `json:"devDownloadLinkWindows"`                        // DevDownloadLinkWindows is not used.
	DevSdkUrl                                 string                  `json:"devSdkUrl"`                                     // DevSdkUrl is not used.
	DevServerVersionStandalone                string                  `json:"devServerVersionStandalone"`                    // DevServerVersionStandalone is not used.
	DisCountdown                              string                  `json:"disCountdown"`                                  // DisCountdown not used.
	DisableAvatarCopying                      bool                    `json:"disableAvatarCopying"`                          // DisableAvatarCopying is a flag indicating whether to allow avatar copying.
	DisableAvatarGating                       bool                    `json:"disableAvatarGating"`                           // DisableAvatarGating is a flag indicating whether gate avatar uploads.
	DisableCaptcha                            bool                    `json:"disableCaptcha"`                                // DisableCaptcha is a flag indicating whether to use captcha for world joins.
	DisableCommunityLabs                      bool                    `json:"disableCommunityLabs"`                          // DisableCommunityLabs is a flag indicating whether to disable the Community Labs feature.
	DisableCommunityLabsPromotion             bool                    `json:"disableCommunityLabsPromotion"`                 // DisableCommunityLabsPromotion is a flag indicating whether to disable promotion *out* of the Community Labs.
	DisableEmail                              bool                    `json:"disableEmail"`                                  // DisableEmail is a flag indicating whether to disable email sending.
	DisableEventStream                        bool                    `json:"disableEventStream"`                            // DisableEventStream is a flag indicating whether to disable the event stream. (Pipeline?)
	DisableFeedbackGating                     bool                    `json:"disableFeedbackGating"`                         // DisableFeedbackGating is a flag indicating whether to gate feedback requests.
	DisableFrontendBuilds                     bool                    `json:"disableFrontendBuilds"`                         // DisableFrontendBuilds is a flag indicating whether to disable frontend builds.
	DisableHello                              bool                    `json:"disableHello"`                                  // DisableHello is a flag indicating whether to disable an unknown feature.
	DisableOculusSubs                         bool                    `json:"disableOculusSubs"`                             // DisableOculusSubs is a flag indicating whether to disable the Oculus subscriptions.
	DisableRegistration                       bool                    `json:"disableRegistration"`                           // DisableRegistration is a flag indicating whether to disable registration.
	DisableSteamNetworking                    bool                    `json:"disableSteamNetworking"`                        // DisableSteamNetworking is a flag indicating whether to disable Steam networking.
	DisableTwoFactorAuth                      bool                    `json:"disableTwoFactorAuth"`                          // DisableTwoFactorAuth is a flag indicating whether to disable two-factor authentication.
	DisableUdon                               bool                    `json:"disableUdon"`                                   // DisableUdon is a flag indicating whether to disable Udon.
	DisableUpgradeAccount                     bool                    `json:"disableUpgradeAccount"`                         // DisableUpgradeAccount is a flag indicating whether to disable the account upgrade feature
	DownloadUrls                              ApiDownloadUrlsResponse `json:"downloadUrls"`                                  // DownloadUrls is a map of SDK download urls.
	DynamicWorldRows                          []ApiDynamicWorldRow    `json:"dynamicWorldRows"`                              // DynamicWorldRows is a list of dynamic world rows.
	Events                                    ApiEventsResponse       `json:"events"`                                        // Events is a struct containing client event configuration data.
	FrontendBuildBranch                       string                  `json:"frontendBuildBranch,omitempty"`                 // FrontendBuildBranch is the branch to use for frontend builds.
	GearDemoRoomId                            string                  `json:"gearDemoRoomId"`                                // GearDemoRoomId is not used.
	HomeWorldId                               string                  `json:"homeWorldId"`                                   // HomeWorldId is the id of the default home world.
	HomepageRedirectTarget                    string                  `json:"homepageRedirectTarget"`                        // HomepageRedirectTarget is the target of the homepage redirect.
	HubWorldId                                string                  `json:"hubWorldId"`                                    // HubWorldId is the id of the default hub world. (not used anymore)
	JobsEmail                                 string                  `json:"jobsEmail"`                                     // JobsEmail is the email address for the jobs email.
	MessageOfTheDay                           string                  `json:"messageOfTheDay"`                               // MessageOfTheDay is the message of the day.
	ModerationEmail                           string                  `json:"moderationEmail"`                               // ModerationEmail is the email address for the moderation email.
	ModerationQueryPeriod                     int64                   `json:"moderationQueryPeriod"`                         // ModerationQueryPeriod is the period in seconds for querying moderation data.
	NotAllowedToSelectAvatarInPrivateWorldMsg string                  `json:"notAllowedToSelectAvatarInPrivateWorldMessage"` // NotAllowedToSelectAvatarInPrivateWorldMsg is the message to display when a user tries to select an avatar in a private world but their rank does not allow them to do so.
	Plugin                                    string                  `json:"plugin"`                                        // Plugin is the name of the Photon plugin.
	ReleaseAppVersionStandalone               string                  `json:"releaseAppVersionStandalone"`                   // ReleaseAppVersionStandalone is the version of the standalone client.
	ReleaseSdkUrl                             string                  `json:"releaseSdkUrl"`                                 // ReleaseSdkUrl is the url of the release SDK.
	ReleaseSdkVersion                         string                  `json:"releaseSdkVersion"`                             // ReleaseSdkVersion is the version of the release SDK.
	ReleaseServerVersionStandalone            string                  `json:"releaseServerVersionStandalone"`                // ReleaseServerVersionStandalone is not used.
	SdkDeveloperFaqUrl                        string                  `json:"sdkDeveloperFaqUrl"`                            // SdkDeveloperFaqUrl is the url of the SDK developer faq.
	SdkDiscordUrl                             string                  `json:"sdkDiscordUrl"`                                 // SdkDiscordUrl is the url of the SDK discord.
	SdkNotAllowedToPublishMsg                 string                  `json:"sdkNotAllowedToPublishMessage"`                 // SdkNotAllowedToPublishMsg is the message to display when a user tries to publish content but their rank does not allow them to do so.
	SdkUnityVersion                           string                  `json:"sdkUnityVersion"`                               // SdkUnityVersion is the version of Unity supported by the SDK.
	ServerName                                string                  `json:"serverName"`                                    // ServerName is the name of the current API instance.
	SupportEmail                              string                  `json:"supportEmail"`                                  // SupportEmail is the email address for the support email.
	TimeoutWorldId                            string                  `json:"timeOutWorldId"`                                // TimeoutWorldId is the id of the timeout world.
	TutorialWorldId                           string                  `json:"tutorialWorldId"`                               // TutorialWorldId is the id of the tutorial world.
	UpdateRateMsMaximum                       int64                   `json:"updateRateMsMaximum"`                           // UpdateRateMsMaximum is direct client configuration.
	UpdateRateMsMinimum                       int64                   `json:"updateRateMsMinimum"`                           // UpdateRateMsMinimum is direct client configuration.
	UpdateRateMsNormal                        int64                   `json:"updateRateMsNormal"`                            // UpdateRateMsNormal is direct client configuration.
	UpdateRateMsUdonManual                    int64                   `json:"updateRateMsUdonManual"`                        // UpdateRateMsUdonManual is direct client configuration.
	UploadAnalysisPercent                     int64                   `json:"uploadAnalysisPercent"`                         // UploadAnalysisPercent is not used.
	UrlList                                   []string                `json:"urlList"`                                       // UrlList is a whitelist of URLs that can be accessed by video players from within the client with "Allow Untrusted URLs" off.
	UseReliableUdpForVoice                    bool                    `json:"useReliableUdpForVoice"`                        // UseReliableUdpForVoice is whether to use reliable UDP for voice.
	UserUpdatePeriod                          int64                   `json:"userUpdatePeriod"`                              // UserUpdatePeriod ???
	UserVerificationDelay                     int64                   `json:"userVerificationDelay"`                         // UserVerificationDelay ???
	UserVerificationRetry                     int64                   `json:"userVerificationRetry"`                         // UserVerificationRetry ???
	UserVerificationTimeout                   int64                   `json:"userVerificationTimeout"`                       // UserVerificationTimeout ???
	ViveWindowsUrl                            string                  `json:"viveWindowsUrl"`                                // ViveWindowsUrl is the url of the Vive Windows Client.
	WhitelistedAssetUrls                      []string                `json:"whiteListedAssetUrls"`                          // WhitelistedAssetUrls is a whitelist of URLs that the client can retrieve assets from.
	WorldUpdatePeriod                         int64                   `json:"worldUpdatePeriod"`                             // WorldUpdatePeriod ???
	YoutubeDLHash                             string                  `json:"youtubedl-hash"`                                // YoutubeDLHash is the hash of the youtube-dl binary.
	YoutubeDLVersion                          string                  `json:"youtubedl-version"`                             // YoutubeDLVersion is the version of youtube-dl.
}

// NewApiConfigResponse returns a new instance of ApiConfigResponse.
// This is used to create the response for the /api/config endpoint.
func NewApiConfigResponse(config *ApiConfig) *ApiConfigResponse {
	return &ApiConfigResponse{
		Address:                       config.Address.Get(),
		Announcements:                 config.Announcements.Get(),
		ApiKey:                        config.ApiKey.Get(),
		AppName:                       config.AppName.Get(),
		BuildVersionTag:               config.BuildVersionTag.Get(),
		CaptchaPercentage:             config.CaptchaPercentage.Get(),
		ClientApiKey:                  config.ClientApiKey.Get(),
		ClientBPSCeiling:              config.ClientBPSCeiling.Get(),
		ClientDisconnectTimeout:       config.ClientDisconnectTimeout.Get(),
		ClientReservedPlayerBPS:       config.ClientReservedPlayerBPS.Get(),
		ClientSentCountAllowance:      config.ClientSentCountAllowance.Get(),
		ContactEmail:                  config.ContactEmail.Get(),
		CopyrightEmail:                config.CopyrightEmail.Get(),
		CurrentTOSVersion:             config.CurrentTOSVersion.Get(),
		DefaultAvatar:                 config.DefaultAvatar.Get(),
		DeploymentGroup:               config.DeploymentGroup.Get(),
		DevAppVersionStandalone:       config.DevAppVersionStandalone.Get(),
		DevDownloadLinkWindows:        config.DevDownloadLinkWindows.Get(),
		DevSdkUrl:                     config.DevSdkUrl.Get(),
		DevServerVersionStandalone:    config.DevServerVersionStandalone.Get(),
		DisCountdown:                  config.DisCountdown.Get(),
		DisableAvatarCopying:          config.DisableAvatarCopying.Get(),
		DisableAvatarGating:           config.DisableAvatarGating.Get(),
		DisableCaptcha:                config.DisableCaptcha.Get(),
		DisableCommunityLabs:          config.DisableCommunityLabs.Get(),
		DisableCommunityLabsPromotion: config.DisableCommunityLabsPromotion.Get(),
		DisableEmail:                  config.DisableEmail.Get(),
		DisableEventStream:            config.DisableEventStream.Get(),
		DisableFeedbackGating:         config.DisableFeedbackGating.Get(),
		DisableOculusSubs:             config.DisableOculusSubs.Get(),
		DisableRegistration:           config.DisableRegistration.Get(),
		DisableSteamNetworking:        config.DisableSteamNetworking.Get(),
		DisableTwoFactorAuth:          config.DisableTwoFactorAuth.Get(),
		DisableUdon:                   config.DisableUdon.Get(),
		DisableUpgradeAccount:         config.DisableUpgradeAccount.Get(),
		DownloadUrls:                  config.DownloadUrls.Get(),
		DynamicWorldRows:              config.DynamicWorldRows.Get(),
		Events:                        config.Events.Get(),
		GearDemoRoomId:                config.GearDemoRoomId.Get(),
		HomeWorldId:                   config.HomeWorldId.Get(),
		HomepageRedirectTarget:        config.HomepageRedirectTarget.Get(),
		HubWorldId:                    config.HubWorldId.Get(),
		JobsEmail:                     config.JobsEmail.Get(),
		MessageOfTheDay:               config.MessageOfTheDay.Get(),
		ModerationEmail:               config.ModerationEmail.Get(),
		ModerationQueryPeriod:         config.ModerationQueryPeriod.Get(),
		NotAllowedToSelectAvatarInPrivateWorldMsg: config.NotAllowedToSelectAvatarInPrivateWorldMsg.Get(),
		Plugin:                         config.Plugin.Get(),
		ReleaseAppVersionStandalone:    config.ReleaseAppVersionStandalone.Get(),
		ReleaseSdkUrl:                  config.ReleaseSdkUrl.Get(),
		ReleaseSdkVersion:              config.ReleaseSdkVersion.Get(),
		ReleaseServerVersionStandalone: config.ReleaseServerVersionStandalone.Get(),
		SdkDeveloperFaqUrl:             config.SdkDeveloperFaqUrl.Get(),
		SdkDiscordUrl:                  config.SdkDiscordUrl.Get(),
		SdkNotAllowedToPublishMsg:      config.SdkNotAllowedToPublishMsg.Get(),
		SdkUnityVersion:                config.SdkUnityVersion.Get(),
		ServerName:                     config.ServerName.Get(),
		SupportEmail:                   config.SupportEmail.Get(),
		TimeoutWorldId:                 config.TimeoutWorldId.Get(),
		TutorialWorldId:                config.TutorialWorldId.Get(),
		UpdateRateMsMaximum:            config.UpdateRateMsMaximum.Get(),
		UpdateRateMsMinimum:            config.UpdateRateMsMinimum.Get(),
		UpdateRateMsNormal:             config.UpdateRateMsNormal.Get(),
		UpdateRateMsUdonManual:         config.UpdateRateMsUdonManual.Get(),
		UploadAnalysisPercent:          config.UploadAnalysisPercent.Get(),
		UrlList:                        config.UrlList.Get(),
		UseReliableUdpForVoice:         config.UseReliableUdpForVoice.Get(),
		UserUpdatePeriod:               config.UserUpdatePeriod.Get(),
		UserVerificationDelay:          config.UserVerificationDelay.Get(),
		UserVerificationRetry:          config.UserVerificationRetry.Get(),
		UserVerificationTimeout:        config.UserVerificationTimeout.Get(),
		YoutubeDLHash:                  config.YoutubeDLHash.Get(),
		YoutubeDLVersion:               config.YoutubeDLVersion.Get(),
		WhitelistedAssetUrls:           config.WhitelistedAssetUrls.Get(),
		WorldUpdatePeriod:              config.WorldUpdatePeriod.Get(),
	}
}
