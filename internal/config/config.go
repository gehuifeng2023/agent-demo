package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultConfigPath         = "configs/local.yaml"
	DefaultServerAddr         = ":8080"
	DefaultLLMMode            = "mock"
	DefaultLLMTimeoutSeconds  = 60
	DefaultRAGTopK            = 3
	DefaultUploadDir          = "knowledge_attachment/days"
	DefaultUploadMaxSizeMB    = 20
	DefaultKnowledgeRootDir   = "knowledge_attachment/default/"
	DefaultSessionMaxMessages = 30
	DefaultSessionRecentLimit = 8
	DefaultToolEnabled        = true
	DefaultHTTPTimeoutSeconds = 10
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	LLM       LLMConfig       `yaml:"llm"`
	RAG       RAGConfig       `yaml:"rag"`
	Upload    UploadConfig    `yaml:"upload"`
	Knowledge KnowledgeConfig `yaml:"knowledge"`
	Session   SessionConfig   `yaml:"session"`
	Tool      ToolConfig      `yaml:"tool"`
	Workflow  WorkflowConfig  `yaml:"workflow"`
}

type ServerConfig struct {
	Addr string `yaml:"addr"`
}
type LLMConfig struct {
	Mode           string `yaml:"mode"`
	BaseURL        string `yaml:"base_url"`
	APIKey         string `yaml:"api_key"`
	Model          string `yaml:"model"`
	TimeoutSeconds int    `yaml:"timeout_seconds"`
}
type RAGConfig struct {
	DocsDir string `yaml:"docs_dir"`
	TopK    int    `yaml:"top_k"`
}
type UploadConfig struct {
	Dir       string `yaml:"dir"`
	MaxSizeMB int64  `yaml:"max_size_mb"`
}
type KnowledgeConfig struct {
	RootDir string `yaml:"root_dir"`
}
type SessionConfig struct {
	MaxMessages int `yaml:"max_messages"`
	RecentLimit int `yaml:"recent_limit"`
}
type ToolConfig struct {
	Enabled            *bool    `yaml:"enabled"`
	RootDir            string   `yaml:"root_dir"`
	HTTPAllowedHosts   []string `yaml:"http_allowed_hosts"`
	HTTPTimeoutSeconds int      `yaml:"http_timeout_seconds"`
}
type WorkflowConfig struct {
	Definitions []WorkflowDefinitionConfig `yaml:"definitions"`
}
type WorkflowDefinitionConfig struct {
	ID    string               `yaml:"id"`
	Nodes []WorkflowNodeConfig `yaml:"nodes"`
}
type WorkflowNodeConfig struct {
	Name      string `yaml:"name"`
	Tool      string `yaml:"tool"`
	Input     string `yaml:"input"`
	OutputKey string `yaml:"output_key"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	cfg.ApplyDefaults()
	return &cfg, nil
}

func (c *Config) ApplyDefaults() {
	if c.Server.Addr == "" {
		c.Server.Addr = DefaultServerAddr
	}
	if c.LLM.Mode == "" {
		c.LLM.Mode = DefaultLLMMode
	}
	if c.LLM.TimeoutSeconds <= 0 {
		c.LLM.TimeoutSeconds = DefaultLLMTimeoutSeconds
	}
	if c.RAG.TopK <= 0 {
		c.RAG.TopK = DefaultRAGTopK
	}
	if c.Upload.Dir == "" {
		c.Upload.Dir = DefaultUploadDir
	}
	if c.Upload.MaxSizeMB <= 0 {
		c.Upload.MaxSizeMB = DefaultUploadMaxSizeMB
	}
	if c.Knowledge.RootDir == "" {
		c.Knowledge.RootDir = DefaultKnowledgeRootDir
	}
	if c.Session.MaxMessages <= 0 {
		c.Session.MaxMessages = DefaultSessionMaxMessages
	}
	if c.Session.RecentLimit <= 0 {
		c.Session.RecentLimit = DefaultSessionRecentLimit
	}
	if c.Tool.HTTPTimeoutSeconds <= 0 {
		c.Tool.HTTPTimeoutSeconds = DefaultHTTPTimeoutSeconds
	}
}

func (c Config) UploadMaxBytes() int64 {
	return c.Upload.MaxSizeMB << 20
}

func (c Config) LLMTimeout() time.Duration {
	return time.Duration(c.LLM.TimeoutSeconds) * time.Second
}

func (c Config) ToolEnabled() bool {
	if c.Tool.Enabled == nil {
		return DefaultToolEnabled
	}
	return *c.Tool.Enabled
}

func (c Config) ToolRootDir() string {
	if c.Tool.RootDir != "" {
		return c.Tool.RootDir
	}
	return c.Knowledge.RootDir
}

func (c Config) HTTPToolTimeout() time.Duration {
	return time.Duration(c.Tool.HTTPTimeoutSeconds) * time.Second
}
