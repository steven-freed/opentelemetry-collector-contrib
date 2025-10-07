package redfishreceiver // import "github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redfishreceiver"

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/redfishreceiver/internal/metadata"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/pdata/pcommon"
	"go.opentelemetry.io/collector/pdata/pmetric"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper/scrapererror"
)

// scraperClient is a struct containing the RedfishClient
// and the resources it needs to collect.
type scraperClient struct {
	*redfishClient
	ResourceSet map[Resource]bool
}

type redfishScraper struct {
	clients  []*scraperClient
	cfg      *Config
	settings component.TelemetrySettings
	mb       *metadata.MetricsBuilder
	logger   *zap.Logger
}

func newScraper(conf *Config, settings receiver.Settings) *redfishScraper {
	return &redfishScraper{
		cfg:      conf,
		settings: settings.TelemetrySettings,
		mb:       metadata.NewMetricsBuilder(conf.MetricsBuilderConfig, settings),
		clients:  make([]*scraperClient, 0, len(conf.Servers)),
		logger:   settings.Logger,
	}
}

func (s *redfishScraper) start(ctx context.Context, host component.Host) error {
	s.logger.Info("starting scraper")

	// create redfish clients
	for _, server := range s.cfg.Servers {
		// sets default timeout for redfish client
		timeout := time.Second * 60
		if server.Timeout != "" {
			timeout, _ = time.ParseDuration(server.Timeout)
		}

		// create redfish client from spec
		client, err := NewRedfishClient(
			server.ComputerSystemId,
			server.User,
			server.Pwd,
			server.Host,
			server.Redfish.Version,
			timeout,
			server.Insecure,
		)
		if err != nil {
			return err
		}

		// create resource set from spec
		rcMap := make(map[Resource]bool, len(server.Resources))
		for _, rc := range server.Resources {
			rcMap[rc] = true
		}

		// append new redfish client
		s.clients = append(
			s.clients,
			&scraperClient{client, rcMap},
		)
	}

	return nil
}

func (s *redfishScraper) scrape(ctx context.Context) (pmetric.Metrics, error) {
	s.logger.Info("scraping")

	errs := &scrapererror.ScrapeErrors{}
	now := pcommon.NewTimestampFromTime(time.Now())
	for _, client := range s.clients {
		baseUrl := client.baseURL.String()

		// ComputerSystem
		compSys, err := client.GetComputerSystem()
		if err != nil {
			errs.Add(err)
			continue
		}

		if _, exists := client.ResourceSet[ComputerSystemResource]; exists {
			s.scrapeComputerSystem(now, baseUrl, compSys)
		}

		// Chassis
		for _, link := range compSys.Links.Chassis {

			chassis, err := client.GetChassis(link.Ref)
			if err != nil {
				errs.Add(err)
				continue
			}

			if _, exists := client.ResourceSet[ChassisResource]; exists {
				s.scrapeChassis(now, compSys.HostName, baseUrl, chassis)
			}

			// Fans and Temperatures
			if client.ResourceSet[FansResource] || client.ResourceSet[TemperaturesResource] {
				thermal, err := client.GetThermal(chassis.Thermal.Ref)
				if err != nil {
					errs.Add(err)
					continue
				}
				if client.ResourceSet[FansResource] {
					s.scrapeFans(now, compSys.HostName, baseUrl, chassis.Id, thermal.Fans)
				}
				if client.ResourceSet[TemperaturesResource] {
					s.scrapeTemperatures(now, compSys.HostName, baseUrl, chassis.Id, thermal.Temperatures)
				}
			} // Fans and Temperatures
		} // Chassis

	} // ComputerSystem

	return s.mb.Emit(), errs.Combine()
}
