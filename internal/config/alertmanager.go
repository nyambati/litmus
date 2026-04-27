package config

import (
	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
)

type AlertmanagerConfig struct {
	Global       *GlobalConfig        `yaml:"global,omitempty"`
	Route        *config.Route        `yaml:"route,omitempty"`
	InhibitRules []config.InhibitRule `yaml:"inhibit_rules,omitempty"`
	Receivers    []Receiver           `yaml:"receivers,omitempty"`
	Templates    []string             `yaml:"templates,omitempty"`
}

type GlobalConfig struct {
	ResolveTimeout           *model.Duration `yaml:"resolve_timeout,omitempty"`
	SMTPFrom                 string          `yaml:"smtp_from,omitempty"`
	SMTPSmarthost            string          `yaml:"smtp_smarthost,omitempty"`
	SMTPAuthUsername         string          `yaml:"smtp_auth_username,omitempty"`
	SMTPAuthPassword         string          `yaml:"smtp_auth_password,omitempty"`
	SMTPAuthPasswordFile     string          `yaml:"smtp_auth_password_file,omitempty"`
	SMTPAuthSecret           string          `yaml:"smtp_auth_secret,omitempty"`
	SMTPAuthSecretFile       string          `yaml:"smtp_auth_secret_file,omitempty"`
	SMTPRequireTLS           *bool           `yaml:"smtp_require_tls,omitempty"`
	SlackAPIURL              string          `yaml:"slack_api_url,omitempty"`
	SlackAPIURLFile          string          `yaml:"slack_api_url_file,omitempty"`
	SlackAppToken            string          `yaml:"slack_app_token,omitempty"`
	SlackAppTokenFile        string          `yaml:"slack_app_token_file,omitempty"`
	VictorOpsAPIURL          string          `yaml:"victorops_api_url,omitempty"`
	VictorOpsAPIKey          string          `yaml:"victorops_api_key,omitempty"`
	VictorOpsAPIKeyFile      string          `yaml:"victorops_api_key_file,omitempty"`
	PagerdutyURL             string          `yaml:"pagerduty_url,omitempty"`
	OpsGenieAPIURL           string          `yaml:"opsgenie_api_url,omitempty"`
	OpsGenieAPIKey           string          `yaml:"opsgenie_api_key,omitempty"`
	OpsGenieAPIKeyFile       string          `yaml:"opsgenie_api_key_file,omitempty"`
	WechatAPIURL             string          `yaml:"wechat_api_url,omitempty"`
	WechatAPISecret          string          `yaml:"wechat_api_secret,omitempty"`
	WechatAPICorpID          string          `yaml:"wechat_api_corp_id,omitempty"`
	WechatAPIAgentID         string          `yaml:"wechat_api_agent_id,omitempty"`
	TelegramAPIUrl           string          `yaml:"telegram_api_url,omitempty"`
	TelegramBotToken         string          `yaml:"telegram_bot_token,omitempty"`
	TelegramBotTokenFile     string          `yaml:"telegram_bot_token_file,omitempty"`
	WebexAPIURL              string          `yaml:"webex_api_url,omitempty"`
	DiscordURL               string          `yaml:"discord_url,omitempty"`
	JiraAPIURL               string          `yaml:"jira_api_url,omitempty"`
	RocketchatAPIURL         string          `yaml:"rocketchat_api_url,omitempty"`
	RocketchatToken          string          `yaml:"rocketchat_token,omitempty"`
	RocketchatTokenFile      string          `yaml:"rocketchat_token_file,omitempty"`
	RocketchatTokenID        string          `yaml:"rocketchat_token_id,omitempty"`
	RocketchatTokenIDFile    string          `yaml:"rocketchat_token_id_file,omitempty"`
	MattermostWebhookURL     string          `yaml:"mattermost_webhook_url,omitempty"`
	MattermostWebhookURLFile string          `yaml:"mattermost_webhook_url_file,omitempty"`
}

type Receiver struct {
	Name              string              `yaml:"name"`
	WebhookConfigs    []*WebhookConfig    `yaml:"webhook_configs,omitempty"`
	SlackConfigs      []*SlackConfig      `yaml:"slack_configs,omitempty"`
	PagerdutyConfigs  []*PagerdutyConfig  `yaml:"pagerduty_configs,omitempty"`
	EmailConfigs      []*EmailConfig      `yaml:"email_configs,omitempty"`
	SlackUnitConfigs  []*SlackUnitConfig  `yaml:"slack_unit_configs,omitempty"`
	OpsGenieConfigs   []*OpsGenieConfig   `yaml:"opsgenie_configs,omitempty"`
	WechatConfigs     []*WechatConfig     `yaml:"wechat_configs,omitempty"`
	PushoverConfigs   []*PushoverConfig   `yaml:"pushover_configs,omitempty"`
	VictorOpsConfigs  []*VictorOpsConfig  `yaml:"victorops_configs,omitempty"`
	SNSConfigs        []*SNSConfig        `yaml:"sns_configs,omitempty"`
	DiscordConfigs    []*DiscordConfig    `yaml:"discord_configs,omitempty"`
	WebexConfigs      []*WebexConfig      `yaml:"webex_configs,omitempty"`
	TelegramConfigs   []*TelegramConfig   `yaml:"telegram_configs,omitempty"`
	MSTeamsConfigs    []*MSTeamsConfig    `yaml:"msteams_configs,omitempty"`
	JiraConfigs       []*JiraConfig       `yaml:"jira_configs,omitempty"`
	RocketchatConfigs []*RocketchatConfig `yaml:"rocketchat_configs,omitempty"`
	MattermostConfigs []*MattermostConfig `yaml:"mattermost_configs,omitempty"`
}

type WebhookConfig struct {
	URL          string            `yaml:"url,omitempty"`
	MaxAlerts    uint64            `yaml:"max_alerts,omitempty"`
	HTTPConfig   *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
}

type SlackConfig struct {
	APIURL       string            `yaml:"api_url,omitempty"`
	Channel      string            `yaml:"channel,omitempty"`
	Username     string            `yaml:"username,omitempty"`
	Color        string            `yaml:"color,omitempty"`
	IconEmoji    string            `yaml:"icon_emoji,omitempty"`
	IconURL      string            `yaml:"icon_url,omitempty"`
	LinkNames    *bool             `yaml:"link_names,omitempty"`
	ShortFields  *bool             `yaml:"short_fields,omitempty"`
	UnfurlLinks  *bool             `yaml:"unfurl_links,omitempty"`
	UnfurlMedia  *bool             `yaml:"unfurl_media,omitempty"`
	Footer       string            `yaml:"footer,omitempty"`
	Text         string            `yaml:"text,omitempty"`
	Blocks       []any             `yaml:"blocks,omitempty"`
	Actions      []any             `yaml:"actions,omitempty"`
	HTTPConfig   *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
}

type SlackUnitConfig struct {
	APIURL       string            `yaml:"api_url,omitempty"`
	Channel      string            `yaml:"channel,omitempty"`
	Username     string            `yaml:"username,omitempty"`
	Color        string            `yaml:"color,omitempty"`
	IconEmoji    string            `yaml:"icon_emoji,omitempty"`
	IconURL      string            `yaml:"icon_url,omitempty"`
	LinkNames    *bool             `yaml:"link_names,omitempty"`
	ShortFields  *bool             `yaml:"short_fields,omitempty"`
	UnfurlLinks  *bool             `yaml:"unfurl_links,omitempty"`
	UnfurlMedia  *bool             `yaml:"unfurl_media,omitempty"`
	Footer       string            `yaml:"footer,omitempty"`
	Text         string            `yaml:"text,omitempty"`
	Blocks       []any             `yaml:"blocks,omitempty"`
	Actions      []any             `yaml:"actions,omitempty"`
	HTTPConfig   *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
}

type PagerdutyConfig struct {
	URL           string            `yaml:"url,omitempty"`
	RoutingKey    string            `yaml:"routing_key,omitempty"`
	ServiceKey    string            `yaml:"service_key,omitempty"`
	ResolverKey   string            `yaml:"resolver_key,omitempty"`
	Severity      string            `yaml:"severity,omitempty"`
	Class         string            `yaml:"class,omitempty"`
	Component     string            `yaml:"component,omitempty"`
	Group         string            `yaml:"group,omitempty"`
	Source        string            `yaml:"source,omitempty"`
	CustomDetails map[string]string `yaml:"custom_details,omitempty"`
	HTTPConfig    *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved  *bool             `yaml:"send_resolved,omitempty"`
}

type EmailConfig struct {
	To           string            `yaml:"to,omitempty"`
	From         string            `yaml:"from,omitempty"`
	Smarthost    string            `yaml:"smarthost,omitempty"`
	AuthUsername string            `yaml:"auth_username,omitempty"`
	AuthPassword string            `yaml:"auth_password,omitempty"`
	AuthSecret   string            `yaml:"auth_secret,omitempty"`
	AuthIdentity string            `yaml:"auth_identity,omitempty"`
	Headers      map[string]string `yaml:"headers,omitempty"`
	HTML         string            `yaml:"html,omitempty"`
	Text         string            `yaml:"text,omitempty"`
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
}

type OpsGenieConfig struct {
	APIURL       string            `yaml:"api_url,omitempty"`
	APIKey       string            `yaml:"api_key,omitempty"`
	APIKeyFile   string            `yaml:"api_key_file,omitempty"`
	RoutingKey   string            `yaml:"routing_key,omitempty"`
	Message      string            `yaml:"message,omitempty"`
	Source       string            `yaml:"source,omitempty"`
	Tags         []string          `yaml:"tags,omitempty"`
	Note         string            `yaml:"note,omitempty"`
	Priority     string            `yaml:"priority,omitempty"`
	HTTPConfig   *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
}

type WechatConfig struct {
	APIURL       string            `yaml:"api_url,omitempty"`
	APISecret    string            `yaml:"api_secret,omitempty"`
	CorpID       string            `yaml:"corp_id,omitempty"`
	AgentID      string            `yaml:"agent_id,omitempty"`
	ToUser       string            `yaml:"to_user,omitempty"`
	ToParty      string            `yaml:"to_party,omitempty"`
	ToTag        string            `yaml:"to_tag,omitempty"`
	Message      string            `yaml:"message,omitempty"`
	HTTPConfig   *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
}

type PushoverConfig struct {
	URL          string            `yaml:"url,omitempty"`
	UserKey      string            `yaml:"user_key,omitempty"`
	Token        string            `yaml:"token,omitempty"`
	Title        string            `yaml:"title,omitempty"`
	Message      string            `yaml:"message,omitempty"`
	Device       string            `yaml:"device,omitempty"`
	Priority     string            `yaml:"priority,omitempty"`
	Retry        string            `yaml:"retry,omitempty"`
	Expire       string            `yaml:"expire,omitempty"`
	Sound        string            `yaml:"sound,omitempty"`
	HTTPConfig   *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
}

type VictorOpsConfig struct {
	URL               string            `yaml:"url,omitempty"`
	APIKey            string            `yaml:"api_key,omitempty"`
	RoutingKey        string            `yaml:"routing_key,omitempty"`
	MessageType       string            `yaml:"message_type,omitempty"`
	EntityID          string            `yaml:"entity_id,omitempty"`
	EntityDisplayName string            `yaml:"entity_display_name,omitempty"`
	HTTPConfig        *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved      *bool             `yaml:"send_resolved,omitempty"`
}

type SNSConfig struct {
	URL          string            `yaml:"url,omitempty"`
	AWSConfig    *AWSConfig        `yaml:"aws_config,omitempty"`
	TopicARN     string            `yaml:"topic_arn,omitempty"`
	PhoneNumber  string            `yaml:"phone_number,omitempty"`
	TargetARN    string            `yaml:"target_arn,omitempty"`
	Subject      string            `yaml:"subject,omitempty"`
	HTTPConfig   *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
}

type DiscordConfig struct {
	WebhookURL   string            `yaml:"webhook_url,omitempty"`
	Title        string            `yaml:"title,omitempty"`
	Message      string            `yaml:"message,omitempty"`
	HTTPConfig   *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
}

type WebexConfig struct {
	APIURL       string            `yaml:"api_url,omitempty"`
	RoomID       string            `yaml:"room_id"`
	Message      string            `yaml:"message,omitempty"`
	HTTPConfig   *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
}

type TelegramConfig struct {
	APIUrl               string            `yaml:"api_url,omitempty"`
	BotToken             string            `yaml:"bot_token,omitempty"`
	BotTokenFile         string            `yaml:"bot_token_file,omitempty"`
	ChatID               int64             `yaml:"chat_id,omitempty"`
	Message              string            `yaml:"message,omitempty"`
	DisableNotifications bool              `yaml:"disable_notifications,omitempty"`
	DisableWebPreview    bool              `yaml:"disable_web_page_preview,omitempty"`
	ParseMode            string            `yaml:"parse_mode,omitempty"`
	HTTPConfig           *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved         *bool             `yaml:"send_resolved,omitempty"`
}

type MSTeamsConfig struct {
	WebhookURL   string            `yaml:"webhook_url,omitempty"`
	Title        string            `yaml:"title,omitempty"`
	Summary      string            `yaml:"summary,omitempty"`
	Text         string            `yaml:"text,omitempty"`
	HTTPConfig   *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
}

type JiraConfig struct {
	HTTPConfig   *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
	APIURL       string            `yaml:"api_url,omitempty"`
	User         string            `yaml:"user,omitempty"`
	Password     string            `yaml:"password,omitempty"`
	PasswordFile string            `yaml:"password_file,omitempty"`
	Project      string            `yaml:"project,omitempty"`
	IssueType    string            `yaml:"issue_type,omitempty"`
	Priority     string            `yaml:"priority,omitempty"`
	Assignee     *JiraUser         `yaml:"assignee,omitempty"`
	Reporter     *JiraUser         `yaml:"reporter,omitempty"`
	Labels       []string          `yaml:"labels,omitempty"`
	Summary      string            `yaml:"summary,omitempty"`
	Description  string            `yaml:"description,omitempty"`
	Components   []string          `yaml:"components,omitempty"`
	Unknown      map[string]any    `yaml:"unknown_fields,omitempty"`
	Fields       map[string]any    `yaml:"fields,omitempty"`
	CustomFields []JiraField       `yaml:"custom_fields,omitempty"`
	Type         string            `yaml:"type,omitempty"`
}

type JiraUser struct {
	Name string `yaml:"name,omitempty"`
	ID   string `yaml:"id,omitempty"`
}

type JiraField struct {
	Name  string `yaml:"name,omitempty"`
	Value string `yaml:"value,omitempty"`
}

type RocketchatConfig struct {
	HTTPConfig   *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved *bool             `yaml:"send_resolved,omitempty"`
	APIURL       string            `yaml:"api_url,omitempty"`
	Token        string            `yaml:"token,omitempty"`
	TokenFile    string            `yaml:"token_file,omitempty"`
	TokenID      string            `yaml:"token_id,omitempty"`
	TokenIDFile  string            `yaml:"token_id_file,omitempty"`
	Channel      string            `yaml:"channel,omitempty"`
	Username     string            `yaml:"username,omitempty"`
	Title        string            `yaml:"title,omitempty"`
	Text         string            `yaml:"text,omitempty"`
	IconEmoji    string            `yaml:"icon_emoji,omitempty"`
	IconURL      string            `yaml:"icon_url,omitempty"`
}

type MattermostConfig struct {
	HTTPConfig     *HTTPClientConfig `yaml:"http_config,omitempty"`
	SendResolved   *bool             `yaml:"send_resolved,omitempty"`
	WebhookURL     string            `yaml:"webhook_url,omitempty"`
	WebhookURLFile string            `yaml:"webhook_url_file,omitempty"`
	Channel        string            `yaml:"channel,omitempty"`
	Username       string            `yaml:"username,omitempty"`
	IconURL        string            `yaml:"icon_url,omitempty"`
	IconEmoji      string            `yaml:"icon_emoji,omitempty"`
	Title          string            `yaml:"title,omitempty"`
	Text           string            `yaml:"text,omitempty"`
}

type HTTPClientConfig struct {
	BearerToken     string        `yaml:"bearer_token,omitempty"`
	BearerTokenFile string        `yaml:"bearer_token_file,omitempty"`
	Username        string        `yaml:"username,omitempty"`
	Password        string        `yaml:"password,omitempty"`
	AuthType        string        `yaml:"auth_type,omitempty"`
	OAuth2          *OAuth2Config `yaml:"oauth2,omitempty"`
	TLSConfig       *TLSConfig    `yaml:"tls_config,omitempty"`
	FollowRedirects *bool         `yaml:"follow_redirects,omitempty"`
	EnableHTTP2     *bool         `yaml:"enable_http2,omitempty"`
}

type OAuth2Config struct {
	ClientID       string            `yaml:"client_id,omitempty"`
	ClientSecret   string            `yaml:"client_secret,omitempty"`
	TokenURL       string            `yaml:"token_url,omitempty"`
	Scopes         []string          `yaml:"scopes,omitempty"`
	EndpointParams map[string]string `yaml:"endpoint_params,omitempty"`
}

type TLSConfig struct {
	CAFile             string `yaml:"ca_file,omitempty"`
	CertFile           string `yaml:"cert_file,omitempty"`
	KeyFile            string `yaml:"key_file,omitempty"`
	ServerName         string `yaml:"server_name,omitempty"`
	InsecureSkipVerify *bool  `yaml:"insecure_skip_verify,omitempty"`
}

type AWSConfig struct {
	Region string `yaml:"region,omitempty"`
}
