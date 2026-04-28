package config

import (
	"bytes"

	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v3"
)

type AlertmanagerConfig struct {
	Global       *GlobalConfig          `yaml:"global,omitempty" json:"global,omitempty"`
	Route        *amconfig.Route        `yaml:"route,omitempty" json:"route,omitempty"`
	InhibitRules []amconfig.InhibitRule `yaml:"inhibit_rules,omitempty" json:"inhibit_rules,omitempty"`
	Receivers    []Receiver             `yaml:"receivers,omitempty" json:"receivers,omitempty"`
	Templates    []string               `yaml:"templates,omitempty" json:"templates,omitempty"`
	// Deprecated. Remove before v1.0 release.
	MuteTimeIntervals []amconfig.MuteTimeInterval `yaml:"mute_time_intervals,omitempty" json:"mute_time_intervals,omitempty"`
	TimeIntervals     []amconfig.TimeInterval     `yaml:"time_intervals,omitempty" json:"time_intervals,omitempty"`
}

type GlobalConfig struct {
	ResolveTimeout           model.Duration `yaml:"resolve_timeout" json:"resolve_timeout"`
	HTTPConfig               map[string]any `yaml:"http_config,omitempty" json:"http_config,omitempty"`
	JiraAPIURL               string         `yaml:"jira_api_url,omitempty" json:"jira_api_url,omitempty"`
	SMTPFrom                 string         `yaml:"smtp_from,omitempty" json:"smtp_from,omitempty"`
	SMTPHello                string         `yaml:"smtp_hello,omitempty" json:"smtp_hello,omitempty"`
	SMTPSmarthost            string         `yaml:"smtp_smarthost,omitempty" json:"smtp_smarthost,omitempty"`
	SMTPAuthUsername         string         `yaml:"smtp_auth_username,omitempty" json:"smtp_auth_username,omitempty"`
	SMTPAuthPassword         string         `yaml:"smtp_auth_password,omitempty" json:"smtp_auth_password,omitempty"`
	SMTPAuthPasswordFile     string         `yaml:"smtp_auth_password_file,omitempty" json:"smtp_auth_password_file,omitempty"`
	SMTPAuthSecret           string         `yaml:"smtp_auth_secret,omitempty" json:"smtp_auth_secret,omitempty"`
	SMTPAuthSecretFile       string         `yaml:"smtp_auth_secret_file,omitempty" json:"smtp_auth_secret_file,omitempty"`
	SMTPAuthIdentity         string         `yaml:"smtp_auth_identity,omitempty" json:"smtp_auth_identity,omitempty"`
	SMTPRequireTLS           bool           `yaml:"smtp_require_tls" json:"smtp_require_tls,omitempty"`
	SMTPTLSConfig            map[string]any `yaml:"smtp_tls_config,omitempty" json:"smtp_tls_config,omitempty"`
	SMTPForceImplicitTLS     *bool          `yaml:"smtp_force_implicit_tls,omitempty" json:"smtp_force_implicit_tls,omitempty"`
	SlackAPIURL              string         `yaml:"slack_api_url,omitempty" json:"slack_api_url,omitempty"`
	SlackAPIURLFile          string         `yaml:"slack_api_url_file,omitempty" json:"slack_api_url_file,omitempty"`
	SlackAppToken            string         `yaml:"slack_app_token,omitempty" json:"slack_app_token,omitempty"`
	SlackAppTokenFile        string         `yaml:"slack_app_token_file,omitempty" json:"slack_app_token_file,omitempty"`
	SlackAppURL              string         `yaml:"slack_app_url,omitempty" json:"slack_app_url,omitempty"`
	PagerdutyURL             string         `yaml:"pagerduty_url,omitempty" json:"pagerduty_url,omitempty"`
	OpsGenieAPIURL           string         `yaml:"opsgenie_api_url,omitempty" json:"opsgenie_api_url,omitempty"`
	OpsGenieAPIKey           string         `yaml:"opsgenie_api_key,omitempty" json:"opsgenie_api_key,omitempty"`
	OpsGenieAPIKeyFile       string         `yaml:"opsgenie_api_key_file,omitempty" json:"opsgenie_api_key_file,omitempty"`
	WeChatAPIURL             string         `yaml:"wechat_api_url,omitempty" json:"wechat_api_url,omitempty"`
	WeChatAPISecret          string         `yaml:"wechat_api_secret,omitempty" json:"wechat_api_secret,omitempty"`
	WeChatAPISecretFile      string         `yaml:"wechat_api_secret_file,omitempty" json:"wechat_api_secret_file,omitempty"`
	WeChatAPICorpID          string         `yaml:"wechat_api_corp_id,omitempty" json:"wechat_api_corp_id,omitempty"`
	VictorOpsAPIURL          string         `yaml:"victorops_api_url,omitempty" json:"victorops_api_url,omitempty"`
	VictorOpsAPIKey          string         `yaml:"victorops_api_key,omitempty" json:"victorops_api_key,omitempty"`
	VictorOpsAPIKeyFile      string         `yaml:"victorops_api_key_file,omitempty" json:"victorops_api_key_file,omitempty"`
	TelegramAPIUrl           string         `yaml:"telegram_api_url,omitempty" json:"telegram_api_url,omitempty"`
	TelegramBotToken         string         `yaml:"telegram_bot_token,omitempty" json:"telegram_bot_token,omitempty"`
	TelegramBotTokenFile     string         `yaml:"telegram_bot_token_file,omitempty" json:"telegram_bot_token_file,omitempty"`
	WebexAPIURL              string         `yaml:"webex_api_url,omitempty" json:"webex_api_url,omitempty"`
	RocketchatAPIURL         string         `yaml:"rocketchat_api_url,omitempty" json:"rocketchat_api_url,omitempty"`
	RocketchatToken          string         `yaml:"rocketchat_token,omitempty" json:"rocketchat_token,omitempty"`
	RocketchatTokenFile      string         `yaml:"rocketchat_token_file,omitempty" json:"rocketchat_token_file,omitempty"`
	RocketchatTokenID        string         `yaml:"rocketchat_token_id,omitempty" json:"rocketchat_token_id,omitempty"`
	RocketchatTokenIDFile    string         `yaml:"rocketchat_token_id_file,omitempty" json:"rocketchat_token_id_file,omitempty"`
	MattermostWebhookURL     string         `yaml:"mattermost_webhook_url,omitempty" json:"mattermost_webhook_url,omitempty"`
	MattermostWebhookURLFile string         `yaml:"mattermost_webhook_url_file,omitempty" json:"mattermost_webhook_url_file,omitempty"`
}

type ReceiverConfig = map[string]any

type Receiver struct {
	Name              string            `yaml:"name" json:"name"`
	WebhookConfigs    []*ReceiverConfig `yaml:"webhook_configs,omitempty" json:"webhook_configs,omitempty"`
	SlackConfigs      []*ReceiverConfig `yaml:"slack_configs,omitempty" json:"slack_configs,omitempty"`
	PagerdutyConfigs  []*ReceiverConfig `yaml:"pagerduty_configs,omitempty" json:"pagerduty_configs,omitempty"`
	EmailConfigs      []*ReceiverConfig `yaml:"email_configs,omitempty" json:"email_configs,omitempty"`
	OpsGenieConfigs   []*ReceiverConfig `yaml:"opsgenie_configs,omitempty" json:"opsgenie_configs,omitempty"`
	WechatConfigs     []*ReceiverConfig `yaml:"wechat_configs,omitempty" json:"wechat_configs,omitempty"`
	PushoverConfigs   []*ReceiverConfig `yaml:"pushover_configs,omitempty" json:"pushover_configs,omitempty"`
	VictorOpsConfigs  []*ReceiverConfig `yaml:"victorops_configs,omitempty" json:"victorops_configs,omitempty"`
	SNSConfigs        []*ReceiverConfig `yaml:"sns_configs,omitempty" json:"sns_configs,omitempty"`
	DiscordConfigs    []*ReceiverConfig `yaml:"discord_configs,omitempty" json:"discord_configs,omitempty"`
	WebexConfigs      []*ReceiverConfig `yaml:"webex_configs,omitempty" json:"webex_configs,omitempty"`
	TelegramConfigs   []*ReceiverConfig `yaml:"telegram_configs,omitempty" json:"telegram_configs,omitempty"`
	MSTeamsConfigs    []*ReceiverConfig `yaml:"msteams_configs,omitempty" json:"msteams_configs,omitempty"`
	JiraConfigs       []*ReceiverConfig `yaml:"jira_configs,omitempty" json:"jira_configs,omitempty"`
	RocketchatConfigs []*ReceiverConfig `yaml:"rocketchat_configs,omitempty" json:"rocketchat_configs,omitempty"`
	MattermostConfigs []*ReceiverConfig `yaml:"mattermost_configs,omitempty" json:"mattermost_configs,omitempty"`
}

func (c *AlertmanagerConfig) String() (string, error) {
	data, err := yaml.Marshal(c)
	return string(data), err
}

func (c *AlertmanagerConfig) ConvertToAMConfigStruct() (*amconfig.Config, error) {
	data, err := yaml.Marshal(c)
	if err != nil {
		return nil, err
	}
	return amconfig.Load(string(data))
}

func (c *AlertmanagerConfig) MarshalIndent(spaces int) (*bytes.Buffer, error) {
	var buff bytes.Buffer
	encoder := yaml.NewEncoder(&buff)
	encoder.SetIndent(spaces)
	err := encoder.Encode(c)
	return &buff, err
}
