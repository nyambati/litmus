package config

import (
	"bytes"

	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
	"gopkg.in/yaml.v3"
)

type AlertmanagerConfig struct {
	Global       *GlobalConfig          `yaml:"global,omitempty"`
	Route        *amconfig.Route        `yaml:"route,omitempty"`
	InhibitRules []amconfig.InhibitRule `yaml:"inhibit_rules,omitempty"`
	Receivers    []Receiver             `yaml:"receivers,omitempty"`
	Templates    []string               `yaml:"templates,omitempty"`
	// Deprecated. Remove before v1.0 release.
	MuteTimeIntervals []amconfig.MuteTimeInterval `yaml:"mute_time_intervals,omitempty" json:"mute_time_intervals,omitempty"`
	TimeIntervals     []amconfig.TimeInterval     `yaml:"time_intervals,omitempty" json:"time_intervals,omitempty"`
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
	Name              string            `yaml:"name"`
	WebhookConfigs    []*ReceiverConfig `yaml:"webhook_configs,omitempty"`
	SlackConfigs      []*ReceiverConfig `yaml:"slack_configs,omitempty"`
	PagerdutyConfigs  []*ReceiverConfig `yaml:"pagerduty_configs,omitempty"`
	EmailConfigs      []*ReceiverConfig `yaml:"email_configs,omitempty"`
	OpsGenieConfigs   []*ReceiverConfig `yaml:"opsgenie_configs,omitempty"`
	WechatConfigs     []*ReceiverConfig `yaml:"wechat_configs,omitempty"`
	PushoverConfigs   []*ReceiverConfig `yaml:"pushover_configs,omitempty"`
	VictorOpsConfigs  []*ReceiverConfig `yaml:"victorops_configs,omitempty"`
	SNSConfigs        []*ReceiverConfig `yaml:"sns_configs,omitempty"`
	DiscordConfigs    []*ReceiverConfig `yaml:"discord_configs,omitempty"`
	WebexConfigs      []*ReceiverConfig `yaml:"webex_configs,omitempty"`
	TelegramConfigs   []*ReceiverConfig `yaml:"telegram_configs,omitempty"`
	MSTeamsConfigs    []*ReceiverConfig `yaml:"msteams_configs,omitempty"`
	JiraConfigs       []*ReceiverConfig `yaml:"jira_configs,omitempty"`
	RocketchatConfigs []*ReceiverConfig `yaml:"rocketchat_configs,omitempty"`
	MattermostConfigs []*ReceiverConfig `yaml:"mattermost_configs,omitempty"`
}

type ReceiverConfig = map[string]any

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
