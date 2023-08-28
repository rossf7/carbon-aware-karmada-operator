package controller

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/jellydator/ttlcache/v3"
	gridprovider "github.com/thegreenwebfoundation/grid-intensity-go/pkg/provider"
)

type CarbonIntensity struct {
	IsValid   bool
	Location  string
	Units     string
	ValidFrom time.Time
	ValidTo   time.Time
	Value     float64
}

type ClusterCarbonIntensity struct {
	CarbonIntensity CarbonIntensity
	ClusterName     string
}

type CarbonIntensityFetcher interface {
	Fetch(ctx context.Context, clusterName, location string) (ClusterCarbonIntensity, error)
	Provider() string
}

type GridIntensityFetcher struct {
	cache        *ttlcache.Cache[string, CarbonIntensity]
	provider     gridprovider.Interface
	providerName string
}

func NewGridIntensityFetcher(providerName string) (*GridIntensityFetcher, error) {
	var provider gridprovider.Interface

	switch providerName {
	case gridprovider.ElectricityMap:
		apiURL, err := getEnvVar("ELECTRICITY_MAP_API_URL")
		if err != nil {
			return nil, err
		}
		token, err := getEnvVar("ELECTRICITY_MAP_API_TOKEN")
		if err != nil {
			return nil, err
		}
		c := gridprovider.ElectricityMapConfig{
			APIURL: apiURL,
			Token:  token,
		}
		provider, err = gridprovider.NewElectricityMap(c)
		if err != nil {
			return nil, err
		}
	case gridprovider.WattTime:
		apiUser, err := getEnvVar("WATT_TIME_API_USER")
		if err != nil {
			return nil, err
		}
		apiPassword, err := getEnvVar("WATT_TIME_API_PASSWORD")
		if err != nil {
			return nil, err
		}
		c := gridprovider.WattTimeConfig{
			APIUser:     apiUser,
			APIPassword: apiPassword,
		}
		provider, err = gridprovider.NewWattTime(c)
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("provider name %s not supported", providerName)
	}

	return &GridIntensityFetcher{
		cache:        ttlcache.New[string, CarbonIntensity](ttlcache.WithDisableTouchOnHit[string, CarbonIntensity]()),
		provider:     provider,
		providerName: providerName,
	}, nil
}

func (g *GridIntensityFetcher) Fetch(ctx context.Context, clusterName, location string) (ClusterCarbonIntensity, error) {
	item := g.cache.Get(location)
	if item != nil && !item.IsExpired() {
		carbonIntensity := item.Value()
		return ClusterCarbonIntensity{ClusterName: clusterName, CarbonIntensity: carbonIntensity}, nil
	}

	carbonIntensity, err := g.fetch(ctx, location)
	if err != nil {
		return ClusterCarbonIntensity{}, err
	}

	ttl := time.Until(carbonIntensity.ValidTo)
	g.cache.Set(location, carbonIntensity, ttl)

	return ClusterCarbonIntensity{
		CarbonIntensity: carbonIntensity,
		ClusterName:     clusterName,
	}, nil
}

func (g *GridIntensityFetcher) Provider() string {
	return g.providerName
}

func (g *GridIntensityFetcher) fetch(ctx context.Context, location string) (CarbonIntensity, error) {
	carbonIntensity, err := g.provider.GetCarbonIntensity(ctx, location)
	if errors.Is(err, gridprovider.ErrReceivedNon200Status) {
		return CarbonIntensity{IsValid: false, Location: location}, nil
	} else if err != nil {
		return CarbonIntensity{}, nil
	}

	return parseCarbonIntensity(location, carbonIntensity)
}

func getEnvVar(varName string) (string, error) {
	val := os.Getenv(varName)
	if val == "" {
		return "", fmt.Errorf("env var %s must be set", varName)
	}

	return val, nil
}

func parseCarbonIntensity(location string, results []gridprovider.CarbonIntensity) (CarbonIntensity, error) {
	var result gridprovider.CarbonIntensity

	if len(results) == 1 {
		result = results[0]
	} else if len(results) == 0 {
		return CarbonIntensity{IsValid: false, Location: location}, nil
	} else {
		for _, r := range results {
			// If the results have both absolute and relative metrics use the
			// absolute value.
			if r.MetricType == gridprovider.AbsoluteMetricType {
				result = r
				break
			}
		}
	}

	return CarbonIntensity{
		IsValid:   true,
		Location:  location,
		Units:     result.Units,
		ValidFrom: result.ValidFrom,
		ValidTo:   result.ValidTo,
		Value:     result.Value,
	}, nil
}
