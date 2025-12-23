package config

import (
	"flag"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
)

// Constants
const (
	TrueStr = "true" // String representation of boolean true value
)

type Env string

func (e Env) IsProd() bool {
	return e == "prod"
}

var GlobalConfig *Config
var configMutex sync.RWMutex
var lastConfigChangeTime time.Time

// GetLastConfigChangeTime returns the time when the config was last changed
func GetLastConfigChangeTime() time.Time {
	configMutex.RLock()
	defer configMutex.RUnlock()
	return lastConfigChangeTime
}

type Config struct {
	Env           Env               `yaml:"env" mapstructure:"env"`
	App           *AppConfig        `yaml:"app" mapstructure:"app"`
	HTTPServer    *HttpServerConfig `yaml:"http_server" mapstructure:"http_server"`
	MetricsServer *MetricsConfig    `yaml:"metrics_server" mapstructure:"metrics_server"`
	Tracing       *TracingConfig    `yaml:"tracing" mapstructure:"tracing"`
	Log           *LogConfig        `yaml:"log" mapstructure:"log"`
	Redis         *RedisConfig      `yaml:"redis" mapstructure:"redis"`
	Postgre       *PostgreSQLConfig `yaml:"postgres" mapstructure:"postgres"`
	MongoDB       *MongoDBConfig    `yaml:"mongodb" mapstructure:"mongodb"`
	DynamoDB      *DynamoDBConfig   `yaml:"dynamodb" mapstructure:"dynamodb"`
	Kafka         *KafkaConfig      `yaml:"kafka" mapstructure:"kafka"`
	RabbitMQ      *RabbitMQConfig   `yaml:"rabbitmq" mapstructure:"rabbitmq"`
	MigrationDir  string            `yaml:"migration_dir" mapstructure:"migration_dir"`
}

type AppConfig struct {
	Name    string `yaml:"name" mapstructure:"name"`
	Debug   bool   `yaml:"debug" mapstructure:"debug"`
	Version string `yaml:"version" mapstructure:"version"`
}

type HttpServerConfig struct {
	Addr            string `yaml:"addr" mapstructure:"addr"`
	Pprof           bool   `yaml:"pprof" mapstructure:"pprof"`
	DefaultPageSize int    `yaml:"default_page_size" mapstructure:"default_page_size"`
	MaxPageSize     int    `yaml:"max_page_size" mapstructure:"max_page_size"`
	ReadTimeout     string `yaml:"read_timeout" mapstructure:"read_timeout"`
	WriteTimeout    string `yaml:"write_timeout" mapstructure:"write_timeout"`
}

type MetricsConfig struct {
	Addr    string `yaml:"addr" mapstructure:"addr"`
	Enabled bool   `yaml:"enabled" mapstructure:"enabled"`
	Path    string `yaml:"path" mapstructure:"path"`
}

type TracingConfig struct {
	Enabled  bool    `yaml:"enabled" mapstructure:"enabled"`
	Endpoint string  `yaml:"endpoint" mapstructure:"endpoint"`
	Sampler  float64 `yaml:"sampler" mapstructure:"sampler"`
}

type LogConfig struct {
	SavePath         string `yaml:"save_path" mapstructure:"save_path"`
	FileName         string `yaml:"file_name" mapstructure:"file_name"`
	MaxSize          int    `yaml:"max_size" mapstructure:"max_size"`
	MaxAge           int    `yaml:"max_age" mapstructure:"max_age"`
	LocalTime        bool   `yaml:"local_time" mapstructure:"local_time"`
	Compress         bool   `yaml:"compress" mapstructure:"compress"`
	Level            string `yaml:"level" mapstructure:"level"`
	EnableConsole    bool   `yaml:"enable_console" mapstructure:"enable_console"`
	EnableColor      bool   `yaml:"enable_color" mapstructure:"enable_color"`
	EnableCaller     bool   `yaml:"enable_caller" mapstructure:"enable_caller"`
	EnableStacktrace bool   `yaml:"enable_stacktrace" mapstructure:"enable_stacktrace"`
}

type PostgreSQLConfig struct {
	User            string `yaml:"user" mapstructure:"user"`
	Password        string `yaml:"password" mapstructure:"password"`
	Host            string `yaml:"host" mapstructure:"host"`
	Port            int    `yaml:"port" mapstructure:"port"`
	Database        string `yaml:"database" mapstructure:"database"`
	SSLMode         string `yaml:"ssl_mode" mapstructure:"ssl_mode"`
	Options         string `yaml:"options" mapstructure:"options"`
	MaxConnections  int32  `yaml:"max_connections" mapstructure:"max_connections"`
	MinConnections  int32  `yaml:"min_connections" mapstructure:"min_connections"`
	MaxConnLifetime int    `yaml:"max_conn_lifetime" mapstructure:"max_conn_lifetime"`
	IdleTimeout     int    `yaml:"idle_timeout" mapstructure:"idle_timeout"`
	ConnectTimeout  int    `yaml:"connect_timeout" mapstructure:"connect_timeout"`
	TimeZone        string `yaml:"time_zone" mapstructure:"time_zone"`
}

type RedisConfig struct {
	Host         string `yaml:"host" mapstructure:"host"`
	Port         int    `yaml:"port" mapstructure:"port"`
	Password     string `yaml:"password" mapstructure:"password"`
	DB           int    `yaml:"db" mapstructure:"db"`
	PoolSize     int    `yaml:"poolSize" mapstructure:"poolSize"`
	IdleTimeout  int    `yaml:"idleTimeout" mapstructure:"idleTimeout"`
	MinIdleConns int    `yaml:"minIdleConns" mapstructure:"minIdleConns"`
}

type MongoDBConfig struct {
	Host        string `yaml:"host" mapstructure:"host"`
	Port        int    `yaml:"port" mapstructure:"port"`
	Database    string `yaml:"database" mapstructure:"database"`
	User        string `yaml:"user" mapstructure:"user"`
	Password    string `yaml:"password" mapstructure:"password"`
	AuthSource  string `yaml:"auth_source" mapstructure:"auth_source"`
	Options     string `yaml:"options" mapstructure:"options"`
	MinPoolSize int    `yaml:"min_pool_size" mapstructure:"min_pool_size"`
	MaxPoolSize int    `yaml:"max_pool_size" mapstructure:"max_pool_size"`
	IdleTimeout int    `yaml:"idle_timeout" mapstructure:"idle_timeout"`
}

type DynamoDBConfig struct {
	Endpoint        string `yaml:"endpoint" mapstructure:"endpoint"`
	Region          string `yaml:"region" mapstructure:"region"`
	AccessKeyID     string `yaml:"access_key_id" mapstructure:"access_key_id"`
	SecretAccessKey string `yaml:"secret_access_key" mapstructure:"secret_access_key"`
	TablePrefix     string `yaml:"table_prefix" mapstructure:"table_prefix"`
}

type KafkaConfig struct {
	Brokers       []string `yaml:"brokers" mapstructure:"brokers"`
	ConsumerGroup string   `yaml:"consumer_group" mapstructure:"consumer_group"`
	Topics        struct {
		AuditEvents string `yaml:"audit_events" mapstructure:"audit_events"`
	} `yaml:"topics" mapstructure:"topics"`
	Producer struct {
		RequiredAcks int `yaml:"required_acks" mapstructure:"required_acks"`
		MaxRetry     int `yaml:"max_retry" mapstructure:"max_retry"`
	} `yaml:"producer" mapstructure:"producer"`
	Consumer struct {
		AutoCommit     bool `yaml:"auto_commit" mapstructure:"auto_commit"`
		CommitInterval int  `yaml:"commit_interval" mapstructure:"commit_interval"`
	} `yaml:"consumer" mapstructure:"consumer"`
}

type RabbitMQConfig struct {
	Host       string `yaml:"host" mapstructure:"host"`
	Port       int    `yaml:"port" mapstructure:"port"`
	User       string `yaml:"user" mapstructure:"user"`
	Password   string `yaml:"password" mapstructure:"password"`
	VHost      string `yaml:"vhost" mapstructure:"vhost"`
	Exchange   string `yaml:"exchange" mapstructure:"exchange"`
	Queue      string `yaml:"queue" mapstructure:"queue"`
	RoutingKey string `yaml:"routing_key" mapstructure:"routing_key"`
	Prefetch   int    `yaml:"prefetch" mapstructure:"prefetch"`
}

func Load(configPath string, configFile string) (*Config, error) {
	var conf *Config
	vip := viper.New()
	vip.AddConfigPath(configPath)
	vip.SetConfigName(configFile)

	vip.SetConfigType("yaml")
	if err := vip.ReadInConfig(); err != nil {
		return nil, err
	}

	// Enable environment variables to override config
	vip.SetEnvPrefix("APP")
	vip.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	vip.AutomaticEnv()

	err := vip.Unmarshal(&conf)
	if err != nil {
		return nil, err
	}

	// Apply environment variable overrides
	applyEnvOverrides(conf)

	// Setup config file change monitoring
	vip.WatchConfig()
	vip.OnConfigChange(func(e fsnotify.Event) {
		// Reload configuration when file changes
		var newConf Config
		if err := vip.Unmarshal(&newConf); err == nil {
			// Apply environment variable overrides to the new config
			applyEnvOverrides(&newConf)

			// Update global config with new values - with mutex protection
			configMutex.Lock()
			*GlobalConfig = newConf
			lastConfigChangeTime = time.Now()
			configMutex.Unlock()
		}
	})

	return conf, nil
}

// applyEnvOverrides applies environment variable overrides to the configuration
func applyEnvOverrides(conf *Config) {
	// Apply config overrides by category
	applyAppEnvOverrides(conf)
	applyHTTPServerEnvOverrides(conf)
	applyMetricsServerEnvOverrides(conf)
	applyTracingEnvOverrides(conf)
	applyPostgresEnvOverrides(conf)
	applyRedisEnvOverrides(conf)
	applyMongoDBEnvOverrides(conf)
	applyDynamoDBEnvOverrides(conf)
	applyKafkaEnvOverrides(conf)
	applyRabbitMQEnvOverrides(conf)
	applyLogEnvOverrides(conf)

	// Migration directory
	if migrationDir := os.Getenv("APP_MIGRATION_DIR"); migrationDir != "" {
		conf.MigrationDir = migrationDir
	}
}

// applyAppEnvOverrides applies App related environment variables
func applyAppEnvOverrides(conf *Config) {
	// Environment
	if env := os.Getenv("APP_ENV"); env != "" {
		conf.Env = Env(env)
	}

	// App config
	if name := os.Getenv("APP_APP_NAME"); name != "" {
		conf.App.Name = name
	}
	if debug := os.Getenv("APP_APP_DEBUG"); debug != "" {
		conf.App.Debug = debug == TrueStr
	}
	if version := os.Getenv("APP_APP_VERSION"); version != "" {
		conf.App.Version = version
	}
}

// applyHTTPServerEnvOverrides applies HTTP server related environment variables
func applyHTTPServerEnvOverrides(conf *Config) {
	if addr := os.Getenv("APP_HTTP_SERVER_ADDR"); addr != "" {
		conf.HTTPServer.Addr = addr
	}
	if pprof := os.Getenv("APP_HTTP_SERVER_PPROF"); pprof != "" {
		conf.HTTPServer.Pprof = pprof == TrueStr
	}
	if pageSize := os.Getenv("APP_HTTP_SERVER_DEFAULT_PAGE_SIZE"); pageSize != "" {
		if val, err := strconv.Atoi(pageSize); err == nil {
			conf.HTTPServer.DefaultPageSize = val
		}
	}
	if maxPageSize := os.Getenv("APP_HTTP_SERVER_MAX_PAGE_SIZE"); maxPageSize != "" {
		if val, err := strconv.Atoi(maxPageSize); err == nil {
			conf.HTTPServer.MaxPageSize = val
		}
	}
	if readTimeout := os.Getenv("APP_HTTP_SERVER_READ_TIMEOUT"); readTimeout != "" {
		conf.HTTPServer.ReadTimeout = readTimeout
	}
	if writeTimeout := os.Getenv("APP_HTTP_SERVER_WRITE_TIMEOUT"); writeTimeout != "" {
		conf.HTTPServer.WriteTimeout = writeTimeout
	}
}

// applyMetricsServerEnvOverrides applies metrics server related environment variables
func applyMetricsServerEnvOverrides(conf *Config) {
	// Initialize MetricsServer if it doesn't exist
	if conf.MetricsServer == nil {
		conf.MetricsServer = &MetricsConfig{
			Addr:    ":9090",
			Enabled: true,
			Path:    "/metrics",
		}
	}

	if addr := os.Getenv("APP_METRICS_SERVER_ADDR"); addr != "" {
		conf.MetricsServer.Addr = addr
	}
	if enabled := os.Getenv("APP_METRICS_SERVER_ENABLED"); enabled != "" {
		conf.MetricsServer.Enabled = enabled == TrueStr
	}
	if path := os.Getenv("APP_METRICS_SERVER_PATH"); path != "" {
		conf.MetricsServer.Path = path
	}
}

// applyTracingEnvOverrides applies Tracing related environment variables
func applyTracingEnvOverrides(conf *Config) {
	if conf.Tracing == nil {
		return
	}

	if enabled := os.Getenv("APP_TRACING_ENABLED"); enabled != "" {
		conf.Tracing.Enabled = enabled == TrueStr
	}
	if endpoint := os.Getenv("APP_TRACING_ENDPOINT"); endpoint != "" {
		conf.Tracing.Endpoint = endpoint
	}
	if sampler := os.Getenv("APP_TRACING_SAMPLER"); sampler != "" {
		if val, err := strconv.ParseFloat(sampler, 64); err == nil {
			conf.Tracing.Sampler = val
		}
	}
}

// applyPostgresEnvOverrides applies PostgreSQL related environment variables
func applyPostgresEnvOverrides(conf *Config) {
	if host := os.Getenv("APP_POSTGRES_HOST"); host != "" {
		conf.Postgre.Host = host
	}
	if port := os.Getenv("APP_POSTGRES_PORT"); port != "" {
		if val, err := strconv.Atoi(port); err == nil {
			conf.Postgre.Port = val
		}
	}
	if username := os.Getenv("APP_POSTGRES_USERNAME"); username != "" {
		conf.Postgre.User = username
	}
	if password := os.Getenv("APP_POSTGRES_PASSWORD"); password != "" {
		conf.Postgre.Password = password
	}
	if database := os.Getenv("APP_POSTGRES_DB_NAME"); database != "" {
		conf.Postgre.Database = database
	}
	if sslMode := os.Getenv("APP_POSTGRES_SSL_MODE"); sslMode != "" {
		conf.Postgre.SSLMode = sslMode
	}
	if options := os.Getenv("APP_POSTGRES_OPTIONS"); options != "" {
		conf.Postgre.Options = options
	}
	if maxConnections := os.Getenv("APP_POSTGRES_MAX_CONNECTIONS"); maxConnections != "" {
		if val, err := strconv.Atoi(maxConnections); err == nil {
			conf.Postgre.MaxConnections = int32(val)
		}
	}
	if minConnections := os.Getenv("APP_POSTGRES_MIN_CONNECTIONS"); minConnections != "" {
		if val, err := strconv.Atoi(minConnections); err == nil {
			conf.Postgre.MinConnections = int32(val)
		}
	}
	if maxConnLifetime := os.Getenv("APP_POSTGRES_MAX_CONN_LIFETIME"); maxConnLifetime != "" {
		if val, err := strconv.Atoi(maxConnLifetime); err == nil {
			conf.Postgre.MaxConnLifetime = val
		}
	}
	if idleTimeout := os.Getenv("APP_POSTGRES_IDLE_TIMEOUT"); idleTimeout != "" {
		if val, err := strconv.Atoi(idleTimeout); err == nil {
			conf.Postgre.IdleTimeout = val
		}
	}
	if connectTimeout := os.Getenv("APP_POSTGRES_CONNECT_TIMEOUT"); connectTimeout != "" {
		if val, err := strconv.Atoi(connectTimeout); err == nil {
			conf.Postgre.ConnectTimeout = val
		}
	}
	if timeZone := os.Getenv("APP_POSTGRES_TIME_ZONE"); timeZone != "" {
		conf.Postgre.TimeZone = timeZone
	}
}

// applyRedisEnvOverrides applies Redis related environment variables
func applyRedisEnvOverrides(conf *Config) {
	if host := os.Getenv("APP_REDIS_HOST"); host != "" {
		conf.Redis.Host = host
	}
	if port := os.Getenv("APP_REDIS_PORT"); port != "" {
		if val, err := strconv.Atoi(port); err == nil {
			conf.Redis.Port = val
		}
	}
	if password := os.Getenv("APP_REDIS_PASSWORD"); password != "" {
		conf.Redis.Password = password
	}
	if db := os.Getenv("APP_REDIS_DB"); db != "" {
		if val, err := strconv.Atoi(db); err == nil {
			conf.Redis.DB = val
		}
	}
	if poolSize := os.Getenv("APP_REDIS_POOL_SIZE"); poolSize != "" {
		if val, err := strconv.Atoi(poolSize); err == nil {
			conf.Redis.PoolSize = val
		}
	}
	if idleTimeout := os.Getenv("APP_REDIS_IDLE_TIMEOUT"); idleTimeout != "" {
		if val, err := strconv.Atoi(idleTimeout); err == nil {
			conf.Redis.IdleTimeout = val
		}
	}
	if minIdleConns := os.Getenv("APP_REDIS_MIN_IDLE_CONNS"); minIdleConns != "" {
		if val, err := strconv.Atoi(minIdleConns); err == nil {
			conf.Redis.MinIdleConns = val
		}
	}
}

// applyMongoDBEnvOverrides applies MongoDB related environment variables
func applyMongoDBEnvOverrides(conf *Config) {
	if conf.MongoDB == nil {
		return
	}

	if host := os.Getenv("APP_MONGODB_HOST"); host != "" {
		conf.MongoDB.Host = host
	}
	if port := os.Getenv("APP_MONGODB_PORT"); port != "" {
		if val, err := strconv.Atoi(port); err == nil {
			conf.MongoDB.Port = val
		}
	}
	if database := os.Getenv("APP_MONGODB_DATABASE"); database != "" {
		conf.MongoDB.Database = database
	}
	if user := os.Getenv("APP_MONGODB_USER"); user != "" {
		conf.MongoDB.User = user
	}
	if password := os.Getenv("APP_MONGODB_PASSWORD"); password != "" {
		conf.MongoDB.Password = password
	}
	if authSource := os.Getenv("APP_MONGODB_AUTH_SOURCE"); authSource != "" {
		conf.MongoDB.AuthSource = authSource
	}
	if options := os.Getenv("APP_MONGODB_OPTIONS"); options != "" {
		conf.MongoDB.Options = options
	}
	if minPoolSize := os.Getenv("APP_MONGODB_MIN_POOL_SIZE"); minPoolSize != "" {
		if val, err := strconv.Atoi(minPoolSize); err == nil {
			conf.MongoDB.MinPoolSize = val
		}
	}
	if maxPoolSize := os.Getenv("APP_MONGODB_MAX_POOL_SIZE"); maxPoolSize != "" {
		if val, err := strconv.Atoi(maxPoolSize); err == nil {
			conf.MongoDB.MaxPoolSize = val
		}
	}
	if idleTimeout := os.Getenv("APP_MONGODB_IDLE_TIMEOUT"); idleTimeout != "" {
		if val, err := strconv.Atoi(idleTimeout); err == nil {
			conf.MongoDB.IdleTimeout = val
		}
	}
}

// applyLogEnvOverrides applies Log related environment variables
func applyLogEnvOverrides(conf *Config) {
	if savePath := os.Getenv("APP_LOG_SAVE_PATH"); savePath != "" {
		conf.Log.SavePath = savePath
	}
	if fileName := os.Getenv("APP_LOG_FILE_NAME"); fileName != "" {
		conf.Log.FileName = fileName
	}
	if maxSize := os.Getenv("APP_LOG_MAX_SIZE"); maxSize != "" {
		if val, err := strconv.Atoi(maxSize); err == nil {
			conf.Log.MaxSize = val
		}
	}
	if maxAge := os.Getenv("APP_LOG_MAX_AGE"); maxAge != "" {
		if val, err := strconv.Atoi(maxAge); err == nil {
			conf.Log.MaxAge = val
		}
	}
	if localTime := os.Getenv("APP_LOG_LOCAL_TIME"); localTime != "" {
		conf.Log.LocalTime = localTime == TrueStr
	}
	if compress := os.Getenv("APP_LOG_COMPRESS"); compress != "" {
		conf.Log.Compress = compress == TrueStr
	}
	if level := os.Getenv("APP_LOG_LEVEL"); level != "" {
		conf.Log.Level = level
	}
	if enableConsole := os.Getenv("APP_LOG_ENABLE_CONSOLE"); enableConsole != "" {
		conf.Log.EnableConsole = enableConsole == TrueStr
	}
	if enableColor := os.Getenv("APP_LOG_ENABLE_COLOR"); enableColor != "" {
		conf.Log.EnableColor = enableColor == TrueStr
	}
	if enableCaller := os.Getenv("APP_LOG_ENABLE_CALLER"); enableCaller != "" {
		conf.Log.EnableCaller = enableCaller == TrueStr
	}
	if enableStacktrace := os.Getenv("APP_LOG_ENABLE_STACKTRACE"); enableStacktrace != "" {
		conf.Log.EnableStacktrace = enableStacktrace == TrueStr
	}
}

// applyDynamoDBEnvOverrides applies DynamoDB related environment variables
func applyDynamoDBEnvOverrides(conf *Config) {
	if conf.DynamoDB == nil {
		return
	}

	if endpoint := os.Getenv("APP_DYNAMODB_ENDPOINT"); endpoint != "" {
		conf.DynamoDB.Endpoint = endpoint
	}
	if region := os.Getenv("APP_DYNAMODB_REGION"); region != "" {
		conf.DynamoDB.Region = region
	}
	if accessKeyID := os.Getenv("APP_DYNAMODB_ACCESS_KEY_ID"); accessKeyID != "" {
		conf.DynamoDB.AccessKeyID = accessKeyID
	}
	if secretAccessKey := os.Getenv("APP_DYNAMODB_SECRET_ACCESS_KEY"); secretAccessKey != "" {
		conf.DynamoDB.SecretAccessKey = secretAccessKey
	}
	if tablePrefix := os.Getenv("APP_DYNAMODB_TABLE_PREFIX"); tablePrefix != "" {
		conf.DynamoDB.TablePrefix = tablePrefix
	}
}

// applyKafkaEnvOverrides applies Kafka related environment variables
func applyKafkaEnvOverrides(conf *Config) {
	if conf.Kafka == nil {
		return
	}

	if brokers := os.Getenv("APP_KAFKA_BROKERS"); brokers != "" {
		conf.Kafka.Brokers = strings.Split(brokers, ",")
	}
	if consumerGroup := os.Getenv("APP_KAFKA_CONSUMER_GROUP"); consumerGroup != "" {
		conf.Kafka.ConsumerGroup = consumerGroup
	}
	if auditTopic := os.Getenv("APP_KAFKA_TOPICS_AUDIT_EVENTS"); auditTopic != "" {
		conf.Kafka.Topics.AuditEvents = auditTopic
	}
	if requiredAcks := os.Getenv("APP_KAFKA_PRODUCER_REQUIRED_ACKS"); requiredAcks != "" {
		if val, err := strconv.Atoi(requiredAcks); err == nil {
			conf.Kafka.Producer.RequiredAcks = val
		}
	}
	if maxRetry := os.Getenv("APP_KAFKA_PRODUCER_MAX_RETRY"); maxRetry != "" {
		if val, err := strconv.Atoi(maxRetry); err == nil {
			conf.Kafka.Producer.MaxRetry = val
		}
	}
}

// applyRabbitMQEnvOverrides applies RabbitMQ related environment variables
func applyRabbitMQEnvOverrides(conf *Config) {
	if conf.RabbitMQ == nil {
		return
	}

	if host := os.Getenv("APP_RABBITMQ_HOST"); host != "" {
		conf.RabbitMQ.Host = host
	}
	if port := os.Getenv("APP_RABBITMQ_PORT"); port != "" {
		if val, err := strconv.Atoi(port); err == nil {
			conf.RabbitMQ.Port = val
		}
	}
	if user := os.Getenv("APP_RABBITMQ_USER"); user != "" {
		conf.RabbitMQ.User = user
	}
	if password := os.Getenv("APP_RABBITMQ_PASSWORD"); password != "" {
		conf.RabbitMQ.Password = password
	}
	if vhost := os.Getenv("APP_RABBITMQ_VHOST"); vhost != "" {
		conf.RabbitMQ.VHost = vhost
	}
	if exchange := os.Getenv("APP_RABBITMQ_EXCHANGE"); exchange != "" {
		conf.RabbitMQ.Exchange = exchange
	}
	if queue := os.Getenv("APP_RABBITMQ_QUEUE"); queue != "" {
		conf.RabbitMQ.Queue = queue
	}
	if routingKey := os.Getenv("APP_RABBITMQ_ROUTING_KEY"); routingKey != "" {
		conf.RabbitMQ.RoutingKey = routingKey
	}
	if prefetch := os.Getenv("APP_RABBITMQ_PREFETCH"); prefetch != "" {
		if val, err := strconv.Atoi(prefetch); err == nil {
			conf.RabbitMQ.Prefetch = val
		}
	}
}

func Init(path, file string) {
	configPath := flag.String("config-path", path, "path to configuration path")
	configFile := flag.String("config-file", file, "name of configuration file (without extension)")
	flag.Parse()

	conf, err := Load(*configPath, *configFile)
	if err != nil {
		panic("Load config fail : " + err.Error())
	}
	GlobalConfig = conf
}

// GetDuration converts a duration string to time.Duration
func GetDuration(durationStr string) time.Duration {
	return cast.ToDuration(durationStr)
}
