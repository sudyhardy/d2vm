// Copyright 2022 Linka Cloud  All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package run

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Handle flags with multiple occurrences
type MultipleFlag []string

func (f *MultipleFlag) String() string {
	return "A multiple flag is a type of flag that can be repeated any number of times"
}

func (f *MultipleFlag) Set(value string) error {
	*f = append(*f, value)
	return nil
}

func (f *MultipleFlag) Type() string {
	return "multiple-flag"
}

func GetStringValue(envKey string, flagVal string, defaultVal string) string {
	var res string

	// If defined, take the env variable
	if _, ok := os.LookupEnv(envKey); ok {
		res = os.Getenv(envKey)
	}

	// If a flag is specified, this value takes precedence
	// Ignore cases where the flag carries the default value
	if flagVal != "" && flagVal != defaultVal {
		res = flagVal
	}

	// if we still don't have a value, use the default
	if res == "" {
		res = defaultVal
	}
	return res
}

func GetIntValue(envKey string, flagVal int, defaultVal int) int {
	var res int

	// If defined, take the env variable
	if _, ok := os.LookupEnv(envKey); ok {
		var err error
		res, err = strconv.Atoi(os.Getenv(envKey))
		if err != nil {
			res = 0
		}
	}

	// If a flag is specified, this value takes precedence
	// Ignore cases where the flag carries the default value
	if flagVal > 0 {
		res = flagVal
	}

	// if we still don't have a value, use the default
	if res == 0 {
		res = defaultVal
	}
	return res
}

func GetBoolValue(envKey string, flagVal bool) bool {
	var res bool

	// If defined, take the env variable
	if _, ok := os.LookupEnv(envKey); ok {
		switch os.Getenv(envKey) {
		case "":
			res = false
		case "0":
			res = false
		case "false":
			res = false
		case "FALSE":
			res = false
		case "1":
			res = true
		default:
			// catches "true", "TRUE" or anything else
			res = true

		}
	}

	// If a flag is specified, this value takes precedence
	if res != flagVal {
		res = flagVal
	}

	return res
}

func StringToIntArray(l string, sep string) ([]int, error) {
	var err error
	if l == "" {
		return []int{}, err
	}
	s := strings.Split(l, sep)
	i := make([]int, len(s))
	for idx := range s {
		if i[idx], err = strconv.Atoi(s[idx]); err != nil {
			return nil, err
		}
	}
	return i, nil
}

// Convert a multi-line string into an array of strings
func SplitLines(in string) []string {
	res := []string{}

	s := bufio.NewScanner(strings.NewReader(in))
	for s.Scan() {
		res = append(res, s.Text())
	}

	return res
}

// This function parses the "size" parameter of a disk specification
// and returns the size in MB. The "size" parameter defaults to GB, but
// the unit can be explicitly set with either a G (for GB) or M (for
// MB). It returns the disk size in MB.
func GetDiskSizeMB(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	sz := len(s)
	if strings.HasSuffix(s, "M") {
		return strconv.Atoi(s[:sz-1])
	}
	if strings.HasSuffix(s, "G") {
		s = s[:sz-1]
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, err
	}
	return 1024 * i, nil
}

func ConvertMBtoGB(i int) int {
	if i < 1024 {
		return 1
	}

	if i%1024 == 0 {
		return i / 1024
	}

	return (i + (1024 - i%1024)) / 1024
}

// DiskConfig is the config for a disk
type DiskConfig struct {
	Path   string
	Size   int
	Format string
}

// Disks is the type for a list of DiskConfig
type Disks []DiskConfig

func (l *Disks) String() string {
	return fmt.Sprint(*l)
}

// Set is used by flag to configure value from CLI
func (l *Disks) Set(value string) error {
	d := DiskConfig{}
	s := strings.Split(value, ",")
	for _, p := range s {
		c := strings.SplitN(p, "=", 2)
		switch len(c) {
		case 1:
			// assume it is a filename even if no file=x
			d.Path = c[0]
		case 2:
			switch c[0] {
			case "file":
				d.Path = c[1]
			case "size":
				size, err := GetDiskSizeMB(c[1])
				if err != nil {
					return err
				}
				d.Size = size
			case "format":
				d.Format = c[1]
			default:
				return fmt.Errorf("Unknown disk config: %s", c[0])
			}
		}
	}
	*l = append(*l, d)
	return nil
}

func (l *Disks) Type() string {
	return "disk"
}

// PublishedPort is used by some backends to expose a VMs port on the host
type PublishedPort struct {
	Guest    uint16
	Host     uint16
	Protocol string
}

// NewPublishedPort parses a string of the form <host>:<guest>[/<tcp|udp>] and returns a PublishedPort structure
func NewPublishedPort(publish string) (PublishedPort, error) {
	p := PublishedPort{}
	slice := strings.Split(publish, ":")

	if len(slice) < 2 {
		return p, fmt.Errorf("Unable to parse the ports to be published, should be in format <host>:<guest> or <host>:<guest>/<tcp|udp>")
	}

	hostPort, err := strconv.ParseUint(slice[0], 10, 16)
	if err != nil {
		return p, fmt.Errorf("The provided hostPort can't be converted to uint16")
	}

	right := strings.Split(slice[1], "/")

	protocol := "tcp"
	if len(right) == 2 {
		protocol = strings.TrimSpace(strings.ToLower(right[1]))
	}
	if protocol != "tcp" && protocol != "udp" {
		return p, fmt.Errorf("Provided protocol is not valid, valid options are: udp and tcp")
	}

	guestPort, err := strconv.ParseUint(right[0], 10, 16)
	if err != nil {
		return p, fmt.Errorf("The provided guestPort can't be converted to uint16")
	}

	if hostPort < 1 || hostPort > 65535 {
		return p, fmt.Errorf("Invalid hostPort: %d", hostPort)
	}
	if guestPort < 1 || guestPort > 65535 {
		return p, fmt.Errorf("Invalid guestPort: %d", guestPort)
	}

	p.Guest = uint16(guestPort)
	p.Host = uint16(hostPort)
	p.Protocol = protocol
	return p, nil
}