package nerve

import (
	"encoding/json"
	"github.com/n0rad/go-erlog/data"
	"github.com/n0rad/go-erlog/errs"
)

type Report struct {
	Available            bool
	Host                 string
	Port                 int
	HaproxyServerOptions string `json:"haproxy_server_options"`
	Labels               map[string]string
}

func NewReport(content []byte) (*Report, error) {
	var r Report
	err := json.Unmarshal(content, &r)
	return &r, err
}

func (r *Report) toJson() ([]byte, error) {
	return json.Marshal(r)
}

type Reporter interface {
	Report(report Report) error
	Init(s *Service) error
	Destroy()
	getCommon() *ReporterCommon
	GetFields() data.Fields
}

type ReporterCommon struct {
	Type string

	fields data.Fields
}

func (r *ReporterCommon) GetFields() data.Fields {
	return r.fields
}

func (r *ReporterCommon) Init(s *Service) error {
	r.fields = s.fields.WithField("type", r.Type)
	return nil
}

func (r *ReporterCommon) Destroy() {}

func (r *ReporterCommon) getCommon() *ReporterCommon {
	return r
}

func ReporterFromJson(data []byte, s *Service) (Reporter, error) {
	t := &ReporterCommon{}
	if err := json.Unmarshal([]byte(data), t); err != nil {
		return nil, errs.WithE(err, "Failed to unmarshall reporter type")
	}

	fields := s.fields.WithField("type", t.Type)
	var typedReporter Reporter
	switch t.Type {
	case "file":
		typedReporter = NewReporterFile()
	case "console":
		typedReporter = NewReporterConsole()
	case "zookeeper":
		typedReporter = NewReporterZookeeper()
	default:
		return nil, errs.WithF(fields, "Unsupported reporter type")
	}

	if err := typedReporter.getCommon().Init(s); err != nil {
		return nil, errs.WithEF(err, fields, "Failed to init common reporter")
	}

	if err := json.Unmarshal([]byte(data), &typedReporter); err != nil {
		return nil, errs.WithEF(err, fields, "Failed to unmarshall reporter")
	}

	if err := typedReporter.Init(s); err != nil {
		return nil, errs.WithEF(err, fields, "Failed to init reporter")
	}
	return typedReporter, nil
}

func toReport(status error, s *Service) Report {
	return Report{
		Available:            status == nil,
		Host:                 s.Host,
		Port:                 s.Port,
		HaproxyServerOptions: s.HaproxyServerOptions,
		Labels:               s.Labels,
	}
}
