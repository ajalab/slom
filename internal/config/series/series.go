package series

import (
	"io"
	"time"

	"gopkg.in/yaml.v3"
)

type SeriesConfigParser struct {
	Start    time.Time
	End      time.Time
	Interval time.Duration
}

func (p *SeriesConfigParser) Parse(r io.Reader) (*SeriesSetConfig, error) {
	var config SeriesSetConfig

	decoder := yaml.NewDecoder(r)
	if err := decoder.Decode(&config); err != nil {
		return nil, err
	}
	if !p.Start.IsZero() {
		config.Start = p.Start
	}
	if !p.End.IsZero() {
		config.End = p.End
	}
	if p.Interval != 0 {
		config.Interval = p.Interval
	}

	return &config, nil
}
