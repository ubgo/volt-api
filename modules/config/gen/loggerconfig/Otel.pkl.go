// Code generated from Pkl module `Logger`. DO NOT EDIT.
package loggerconfig

type Otel struct {
	Enable bool `pkl:"enable"`

	IsEndpoingUrl bool `pkl:"isEndpoingUrl"`

	Endpoint string `pkl:"endpoint"`

	ApiKey string `pkl:"apiKey"`

	PrefixBasic bool `pkl:"prefixBasic"`

	Insecure bool `pkl:"insecure"`

	EnableMetrics bool `pkl:"enableMetrics"`

	EnableLogs bool `pkl:"enableLogs"`

	EnableTraces bool `pkl:"enableTraces"`

	ServiceName string `pkl:"serviceName"`

	ServiceVersion string `pkl:"serviceVersion"`

	EnableHttp bool `pkl:"enableHttp"`

	EnableStdout bool `pkl:"enableStdout"`
}
