package main

import (
	"container/list"
	"fmt"
	"sort"
	"strings"
)

func parseIni(input []string) map[string]map[string]string {
	m := make(map[string]map[string]string)
	var k string
	for _, str := range input {
		if len(str) == 0 || str[0] == ';' {
			continue
		}
		if str[0] == '[' {
			k = str[1 : len(str)-1]
			m[k] = make(map[string]string)
		} else {
			if len(k) == 0 {
				panic("input is invalid")
			}
			i := strings.Index(str, "=")
			if i == -1 {
				panic("input is valid")
			}
			m[k][str[:i-1]] = str[i+1:]
		}
	}
	return m
}

func PrintIni(ini map[string]map[string]string) {
	groups := make([]string, 0, len(ini))
	for group := range ini {
		groups = append(groups, group)
	}
	sort.Strings(groups)
	for i, group := range groups {
		fmt.Printf("[%s]\n", group)
		keys := make([]string, 0, len(ini[group]))
		for key := range ini[group] {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			fmt.Printf("%s=%s\n", key, ini[group][key])
		}
		if i+1 < len(groups) {
			fmt.Println()
		}
	}
}

/*func PrintIni(ini map[string]map[string]string) {
	groups := make([]string, 0, len(ini))
	for group := range ini {
		groups = append(groups, group)
	}
	sort.Strings(groups)

	for i, group := range groups {
		m := ini[k]
		fmt.Printf("%s: map[", k)
		for mk, mm := range m {
			fmt.Printf("%s: %s")
		}

	}
}*/

func main() {
	iniData := []string{
		"; Cut down copy of Mozilla application.ini file",
		"",
		"[App]",
		"Vendor=Mozilla",
		"Name=Iceweasel",
		"Profile=mozilla/firefox",
		"Version=3.5.16",
		"[Gecko]",
		"MinVersion=1.9.1",
		"MaxVersion=1.9.1.*",
		"[XRE]",
		"EnableProfileMigrator=0",
		"EnableExtensionManager=1",
	}
	ini := parseIni(iniData)
	PrintIni(ini)
}
