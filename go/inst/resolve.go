package inst

import (
	"errors"
	"fmt"
	"github.com/tctalk-cn/gorchestrator/go/config"
	"regexp"
	"strings"
	"time"

	"github.com/openark/golib/log"
)

var hostnameResolvesLightweightCacheLoadedOnceFromDB bool = false

// ResolveHostname Attempt to resolve a hostname. This may return a database cached hostname or otherwise
// it may resolve the hostname via CNAME
func ResolveHostname(hostname string) (string, error) {
	hostname = strings.TrimSpace(hostname)
	if hostname == "" {
		return hostname, errors.New("Will not resolve empty hostname")
	}
	if strings.Contains(hostname, ",") {
		return hostname, fmt.Errorf("Will not resolve multi-hostname: %+v", hostname)
	}
	if (&InstanceKey{Hostname: hostname}).IsDetached() {
		return hostname, nil
	}

	// First go to lightweight cache
	if resolvedHostname, found := getHostnameResolvesLightweightCache().Get(hostname); found {
		return resolvedHostname.(string), nil
	}

	if !hostnameResolvesLightweightCacheLoadedOnceFromDB {
		// A continuous-discovery will first make sure to load all resolves from DB.
		// However cli does not do so.
		// Anyway, it seems like the cache was not loaded from DB. Before doing real resolves,
		// let's try and get the resolved hostname from database.
		if !HostnameResolveMethodIsNone() {
			go func() {
				if resolvedHostname, err := ReadResolvedHostname(hostname); err == nil && resolvedHostname != "" {
					getHostnameResolvesLightweightCache().Set(hostname, resolvedHostname, 0)
				}
			}()
		}
	}

	// Unfound: resolve!
	log.Debugf("Hostname unresolved yet: %s", hostname)
	resolvedHostname, err := resolveHostname(hostname)
	if config.Config.RejectHostnameResolvePattern != "" {
		// Reject, don't even cache
		if matched, _ := regexp.MatchString(config.Config.RejectHostnameResolvePattern, resolvedHostname); matched {
			log.Warningf("ResolveHostname: %+v resolved to %+v but rejected due to RejectHostnameResolvePattern '%+v'", hostname, resolvedHostname, config.Config.RejectHostnameResolvePattern)
			return hostname, nil
		}
	}

	if err != nil {
		// Problem. What we'll do is cache the hostname for just one minute, so as to avoid flooding requests
		// on one hand, yet make it refresh shortly on the other hand. Anyway do not write to database.
		getHostnameResolvesLightweightCache().Set(hostname, resolvedHostname, time.Minute)
		return hostname, err
	}
	// Good result! Cache it, also to DB
	log.Debugf("Cache hostname resolve %s as %s", hostname, resolvedHostname)
	go UpdateResolvedHostname(hostname, resolvedHostname)
	return resolvedHostname, nil

}
