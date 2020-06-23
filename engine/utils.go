package engine

import (
	"fmt"
	specv1 "github.com/baetyl/baetyl-go/spec/v1"
)

const (
	// TODO
	// when baetyl-init applies the beatyl-core deployment yaml， passes the app data
	// path in host through env to baetyl-core, does not hardcode the value here because the host path may be changed during baetyl installation
	appDataHostPath = "/var/lib/baetyl/app-data"
	configKeyObject = "_object_"
)

func checkService(apps map[string]specv1.Application, stats map[string]specv1.AppStats, update map[string]specv1.AppInfo) {
	svcs := make(map[string][]string)
	for n, app := range apps {
		for _, svc := range app.Services {
			svcs[svc.Name] = append(svcs[svc.Name], n)
		}
	}
	var first string
	for sName, aNames := range svcs {
		if len(aNames) <= 1 {
			continue
		}
		// when multiple apps have same service name,
		// it will only launch the first app or not launch any app by deleting update map
		// if there was one app existed which is in stats
		first = ""
		for _, aName := range aNames {
			if first == "" {
				first = aName
			}
			if _, ok := stats[aName]; ok {
				first = aName
			}
		}
		for _, aName := range aNames {
			if aName == first {
				continue
			}
			delete(update, aName)
			delete(apps, aName)
			stat, ok := stats[aName]
			if !ok {
				stat = specv1.AppStats{InstanceStats: map[string]specv1.InstanceStats{}}
			}
			iStat, ok := stat.InstanceStats[sName]
			if !ok {
				iStat = specv1.InstanceStats{ServiceName: sName}
			}
			iStat.Cause += fmt.Sprintf("service [%s] in application [%s] collide with application [%s]", sName, aName, first)
			stat.InstanceStats[sName] = iStat
			stats[aName] = stat
		}
	}
}

func checkPort(apps map[string]specv1.Application, stats map[string]specv1.AppStats, update map[string]specv1.AppInfo) {
	ports := make(map[int32][]string)
	svcs := make(map[string]string)
	for n, app := range apps {
		for _, svc := range app.Services {
			svcs[svc.Name] = n
			for _, p := range svc.Ports {
				ports[p.HostPort] = append(ports[p.HostPort], svc.Name)
			}
		}
	}
	var first string
	for p, sNames := range ports {
		if len(sNames) <= 1 {
			continue
		}
		// when multiple apps have same host port,
		// it will only launch the first app or not launch any app by deleting update map
		// if there was one app existed which is in stats
		first = ""
		for _, sName := range sNames {
			aName := svcs[sName]
			if first == "" {
				first = sName
			}
			if stat, ok := stats[aName]; ok {
				if _, ok := stat.InstanceStats[sName]; ok {
					first = sName
				}
			}
		}
		for _, sName := range sNames {
			aName := svcs[sName]
			if sName == first {
				continue
			}
			delete(update, aName)
			delete(apps, aName)
			stat, ok := stats[aName]
			if !ok {
				stat = specv1.AppStats{}
			}
			if stat.InstanceStats == nil {
				stat.InstanceStats = map[string]specv1.InstanceStats{}
			}
			iStat, ok := stat.InstanceStats[sName]
			if !ok {
				iStat = specv1.InstanceStats{ServiceName: sName}
			}
			iStat.Cause += fmt.Sprintf("port [%d] in service [%s] collide with service [%s]", p, sName, first)
			stat.InstanceStats[sName] = iStat
			stats[aName] = stat
		}
	}
}

func makeKey(kind specv1.Kind, name, ver string) string {
	if name == "" || ver == "" {
		return ""
	}
	return string(kind) + "-" + name + "-" + ver
}

// ensuring apps have same order in report and desire list
func alignApps(reApps, deApps []specv1.AppInfo) []specv1.AppInfo {
	if len(reApps) == 0 || len(deApps) == 0 {
		return reApps
	}
	as := map[string]specv1.AppInfo{}
	for _, a := range reApps {
		as[a.Name] = a
	}
	var res []specv1.AppInfo
	for _, a := range deApps {
		if r, ok := as[a.Name]; ok {
			res = append(res, r)
			delete(as, a.Name)
		}
	}
	for _, a := range as {
		res = append(res, a)
	}
	return res
}

func isRegistrySecret(secret specv1.Secret) bool {
	registry, ok := secret.Labels[specv1.SecretLabel]
	return ok && registry == specv1.SecretRegistry
}
