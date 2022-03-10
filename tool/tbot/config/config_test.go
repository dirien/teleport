/*
Copyright 2022 Gravitational, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package config

import (
	"strings"
	"testing"
	"time"

	"github.com/gravitational/teleport/tool/tbot/identity"
	"github.com/stretchr/testify/require"
)

func TestConfigDefaults(t *testing.T) {
	cfg, err := NewDefaultConfig("auth.example.com")
	require.NoError(t, err)

	require.Equal(t, DefaultCertificateTTL, cfg.CertificateTTL)
	require.Equal(t, DefaultRenewInterval, cfg.RenewInterval)

	storageDest, err := cfg.Storage.GetDestination()
	require.NoError(t, err)

	storageImpl, ok := storageDest.(*DestinationDirectory)
	require.True(t, ok)

	require.Equal(t, defaultStoragePath, storageImpl.Path)

	// Onboarding config unset
	require.Nil(t, cfg.Onboarding)

	// Default config has no destinations (without CLI)
	require.Empty(t, cfg.Destinations)
}

func TestConfigCLIOnlySample(t *testing.T) {
	// Test the sample config generated by `tctl bots add ...`
	cf := CLIConf{
		DestinationDir: "/tmp/foo",
		Token:          "foo",
		CAPins:         []string{"abc123"},
		AuthServer:     "auth.example.com",
	}
	cfg, err := FromCLIConf(&cf)
	require.NoError(t, err)

	require.Equal(t, cf.AuthServer, cfg.AuthServer)

	require.NotNil(t, cfg.Onboarding)
	require.Equal(t, cf.Token, cfg.Onboarding.Token)
	require.Equal(t, cf.CAPins, cfg.Onboarding.CAPins)

	// Storage is still default
	storageDest, err := cfg.Storage.GetDestination()
	require.NoError(t, err)
	storageImpl, ok := storageDest.(*DestinationDirectory)
	require.True(t, ok)
	require.Equal(t, defaultStoragePath, storageImpl.Path)

	// A single default destination should exist
	require.Len(t, cfg.Destinations, 1)
	dest := cfg.Destinations[0]
	require.ElementsMatch(t, []identity.ArtifactKind{identity.KindSSH}, dest.Kinds)

	require.Len(t, dest.Configs, 1)
	template := dest.Configs[0]
	require.NotNil(t, template.SSHClient)

	destImpl, err := dest.GetDestination()
	require.NoError(t, err)
	destImplReal, ok := destImpl.(*DestinationDirectory)
	require.True(t, ok)

	require.Equal(t, cf.DestinationDir, destImplReal.Path)
}

func TestConfigFile(t *testing.T) {
	cfg, err := ReadConfig(strings.NewReader(exampleConfigFile))
	require.NoError(t, err)

	require.Equal(t, "auth.example.com", cfg.AuthServer)
	require.Equal(t, time.Minute*5, cfg.RenewInterval)

	require.NotNil(t, cfg.Onboarding)
	require.Equal(t, "foo", cfg.Onboarding.Token)
	require.ElementsMatch(t, []string{"sha256:abc123"}, cfg.Onboarding.CAPins)

	storage, err := cfg.Storage.GetDestination()
	require.NoError(t, err)

	_, ok := storage.(*DestinationMemory)
	require.True(t, ok)

	require.Len(t, cfg.Destinations, 1)
	destination := cfg.Destinations[0]

	require.ElementsMatch(t, []identity.ArtifactKind{identity.KindSSH, identity.KindTLS}, destination.Kinds)

	require.Len(t, destination.Configs, 1)
	template := destination.Configs[0]
	templateImpl, err := template.GetConfigTemplate()
	require.NoError(t, err)
	sshTemplate, ok := templateImpl.(*TemplateSSHClient)
	require.True(t, ok)
	require.Equal(t, uint16(1234), sshTemplate.ProxyPort)

	destImpl, err := destination.GetDestination()
	require.NoError(t, err)
	destImplReal, ok := destImpl.(*DestinationDirectory)
	require.True(t, ok)
	require.Equal(t, "/tmp/foo", destImplReal.Path)
}

const exampleConfigFile = `
auth_server: auth.example.com
renew_interval: 5m
onboarding:
  token: foo
  ca_pins:
    - sha256:abc123
storage:
  memory: {}
destinations:
  - directory:
      path: /tmp/foo
    kinds: [ssh, tls]
    configs:
      - ssh_client:
          proxy_port: 1234
`
