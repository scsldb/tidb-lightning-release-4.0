// Copyright 2019 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package config_test

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strconv"
	"testing"

	. "github.com/pingcap/check"
	"github.com/pingcap/parser/mysql"

	"github.com/pingcap/tidb-lightning/lightning/config"
)

func Test(t *testing.T) {
	TestingT(t)
}

var _ = Suite(&configTestSuite{})

type configTestSuite struct{}

func startMockServer(c *C, statusCode int, content string) (*httptest.Server, string, int) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
		fmt.Fprint(w, content)
	}))

	url, err := url.Parse(ts.URL)
	c.Assert(err, IsNil)
	host, portString, err := net.SplitHostPort(url.Host)
	c.Assert(err, IsNil)
	port, err := strconv.Atoi(portString)
	c.Assert(err, IsNil)

	return ts, host, port
}

func assignMinimalLegalValue(cfg *config.Config) {
	cfg.TiDB.Host = "123.45.67.89"
	cfg.TiDB.Port = 4567
	cfg.TiDB.StatusPort = 8901
	cfg.TiDB.PdAddr = "234.56.78.90:12345"
}

func (s *configTestSuite) TestAdjustPdAddrAndPort(c *C) {
	ts, host, port := startMockServer(c, http.StatusOK,
		`{"port":4444,"advertise-address":"","path":"123.45.67.89:1234,56.78.90.12:3456"}`,
	)
	defer ts.Close()

	cfg := config.NewConfig()
	cfg.TiDB.Host = host
	cfg.TiDB.StatusPort = port

	err := cfg.Adjust()
	c.Assert(err, IsNil)
	c.Assert(cfg.TiDB.Port, Equals, 4444)
	c.Assert(cfg.TiDB.PdAddr, Equals, "123.45.67.89:1234")
}

func (s *configTestSuite) TestAdjustPdAddrAndPortViaAdvertiseAddr(c *C) {
	ts, host, port := startMockServer(c, http.StatusOK,
		`{"port":6666,"advertise-address":"121.212.121.212:5555","path":"34.34.34.34:3434"}`,
	)
	defer ts.Close()

	cfg := config.NewConfig()
	cfg.TiDB.Host = host
	cfg.TiDB.StatusPort = port

	err := cfg.Adjust()
	c.Assert(err, IsNil)
	c.Assert(cfg.TiDB.Port, Equals, 6666)
	c.Assert(cfg.TiDB.PdAddr, Equals, "34.34.34.34:3434")
}

func (s *configTestSuite) TestAdjustPageNotFound(c *C) {
	ts, host, port := startMockServer(c, http.StatusNotFound, "{}")
	defer ts.Close()

	cfg := config.NewConfig()
	cfg.TiDB.Host = host
	cfg.TiDB.StatusPort = port

	err := cfg.Adjust()
	c.Assert(err, ErrorMatches, "cannot fetch settings from TiDB.*")
}

func (s *configTestSuite) TestAdjustConnectRefused(c *C) {
	ts, host, port := startMockServer(c, http.StatusOK, "{}")

	cfg := config.NewConfig()
	cfg.TiDB.Host = host
	cfg.TiDB.StatusPort = port

	ts.Close() // immediately close to ensure connection refused.

	err := cfg.Adjust()
	c.Assert(err, ErrorMatches, "cannot fetch settings from TiDB.*")
}

func (s *configTestSuite) TestAdjustInvalidBackend(c *C) {
	cfg := config.NewConfig()
	cfg.TikvImporter.Backend = "no_such_backend"
	err := cfg.Adjust()
	c.Assert(err, ErrorMatches, "invalid config: unsupported `tikv-importer\\.backend` \\(no_such_backend\\)")
}

func (s *configTestSuite) TestDecodeError(c *C) {
	ts, host, port := startMockServer(c, http.StatusOK, "invalid-string")
	defer ts.Close()

	cfg := config.NewConfig()
	cfg.TiDB.Host = host
	cfg.TiDB.StatusPort = port

	err := cfg.Adjust()
	c.Assert(err, ErrorMatches, "cannot fetch settings from TiDB.*")
}

func (s *configTestSuite) TestInvalidSetting(c *C) {
	ts, host, port := startMockServer(c, http.StatusOK, `{"port": 0}`)
	defer ts.Close()

	cfg := config.NewConfig()
	cfg.TiDB.Host = host
	cfg.TiDB.StatusPort = port

	err := cfg.Adjust()
	c.Assert(err, ErrorMatches, "invalid `tidb.port` setting")
}

func (s *configTestSuite) TestInvalidPDAddr(c *C) {
	ts, host, port := startMockServer(c, http.StatusOK, `{"port": 1234, "path": ",,"}`)
	defer ts.Close()

	cfg := config.NewConfig()
	cfg.TiDB.Host = host
	cfg.TiDB.StatusPort = port

	err := cfg.Adjust()
	c.Assert(err, ErrorMatches, "invalid `tidb.pd-addr` setting")
}

func (s *configTestSuite) TestAdjustWillNotContactServerIfEverythingIsDefined(c *C) {
	cfg := config.NewConfig()
	assignMinimalLegalValue(cfg)

	err := cfg.Adjust()
	c.Assert(err, IsNil)
	c.Assert(cfg.TiDB.Port, Equals, 4567)
	c.Assert(cfg.TiDB.PdAddr, Equals, "234.56.78.90:12345")
}

func (s *configTestSuite) TestAdjustWillBatchImportRatioInvalid(c *C) {
	cfg := config.NewConfig()
	assignMinimalLegalValue(cfg)
	cfg.Mydumper.BatchImportRatio = -1
	err := cfg.Adjust()
	c.Assert(err, IsNil)
	c.Assert(cfg.Mydumper.BatchImportRatio, Equals, 0.75)
}

func (s *configTestSuite) TestAdjustSecuritySection(c *C) {
	testCases := []struct {
		input       string
		expectedCA  string
		expectedTLS string
	}{
		{
			input:       ``,
			expectedCA:  "",
			expectedTLS: "false",
		},
		{
			input: `
				[security]
			`,
			expectedCA:  "",
			expectedTLS: "false",
		},
		{
			input: `
				[security]
				ca-path = "/path/to/ca.pem"
			`,
			expectedCA:  "/path/to/ca.pem",
			expectedTLS: "cluster",
		},
		{
			input: `
				[security]
				ca-path = "/path/to/ca.pem"
				[tidb.security]
			`,
			expectedCA:  "",
			expectedTLS: "false",
		},
		{
			input: `
				[security]
				ca-path = "/path/to/ca.pem"
				[tidb.security]
				ca-path = "/path/to/ca2.pem"
			`,
			expectedCA:  "/path/to/ca2.pem",
			expectedTLS: "cluster",
		},
		{
			input: `
				[security]
				[tidb.security]
				ca-path = "/path/to/ca2.pem"
			`,
			expectedCA:  "/path/to/ca2.pem",
			expectedTLS: "cluster",
		},
		{
			input: `
				[security]
				[tidb]
				tls = "skip-verify"
				[tidb.security]
			`,
			expectedCA:  "",
			expectedTLS: "skip-verify",
		},
	}

	for _, tc := range testCases {
		comment := Commentf("input = %s", tc.input)

		cfg := config.NewConfig()
		assignMinimalLegalValue(cfg)
		err := cfg.LoadFromTOML([]byte(tc.input))
		c.Assert(err, IsNil, comment)

		err = cfg.Adjust()
		c.Assert(err, IsNil, comment)
		c.Assert(cfg.TiDB.Security.CAPath, Equals, tc.expectedCA, comment)
		c.Assert(cfg.TiDB.TLS, Equals, tc.expectedTLS, comment)
	}
}

func (s *configTestSuite) TestInvalidCSV(c *C) {
	testCases := []struct {
		input string
		err   string
	}{
		{
			input: `
				[mydumper.csv]
				separator = ''
			`,
			err: "invalid config: `mydumper.csv.separator` must be exactly one byte long",
		},
		{
			input: `
				[mydumper.csv]
				separator = 'hello'
			`,
			err: "invalid config: `mydumper.csv.separator` must be exactly one byte long",
		},
		{
			input: `
				[mydumper.csv]
				separator = '\'
				backslash-escape = false
			`,
			err: "",
		},
		{
			input: `
				[mydumper.csv]
				separator = '???'
			`,
			err: "invalid config: `mydumper.csv.separator` must be exactly one byte long",
		},
		{
			input: `
				[mydumper.csv]
				delimiter = ''
			`,
			err: "",
		},
		{
			input: `
				[mydumper.csv]
				delimiter = 'hello'
			`,
			err: "invalid config: `mydumper.csv.delimiter` must be one byte long or empty",
		},
		{
			input: `
				[mydumper.csv]
				delimiter = '\'
				backslash-escape = false
			`,
			err: "",
		},
		{
			input: `
				[mydumper.csv]
				delimiter = '???'
			`,
			err: "invalid config: `mydumper.csv.delimiter` must be one byte long or empty",
		},
		{
			input: `
				[mydumper.csv]
				separator = '|'
				delimiter = '|'
			`,
			err: "invalid config: cannot use the same character for both CSV delimiter and separator",
		},
		{
			input: `
				[mydumper.csv]
				separator = '\'
				backslash-escape = true
			`,
			err: "invalid config: cannot use '\\' as CSV separator when `mydumper.csv.backslash-escape` is true",
		},
		{
			input: `
				[mydumper.csv]
				delimiter = '\'
				backslash-escape = true
			`,
			err: "invalid config: cannot use '\\' as CSV delimiter when `mydumper.csv.backslash-escape` is true",
		},
		{
			input: `
				[tidb]
				sql-mode = "invalid-sql-mode"
			`,
			err: "invalid config: `mydumper.tidb.sql_mode` must be a valid SQL_MODE: ERROR 1231 (42000): Variable 'sql_mode' can't be set to the value of 'invalid-sql-mode'",
		},
		{
			input: `
				[[routes]]
				schema-pattern = ""
				table-pattern = "shard_table_*"
			`,
			err: "schema pattern of table route rule should not be empty",
		},
		{
			input: `
				[[routes]]
				schema-pattern = "schema_*"
				table-pattern = ""
			`,
			err: "target schema of table route rule should not be empty",
		},
	}

	for _, tc := range testCases {
		comment := Commentf("input = %s", tc.input)

		cfg := config.NewConfig()
		cfg.TiDB.Port = 4000
		cfg.TiDB.PdAddr = "test.invalid:2379"
		err := cfg.LoadFromTOML([]byte(tc.input))
		c.Assert(err, IsNil)

		err = cfg.Adjust()
		if tc.err != "" {
			c.Assert(err, ErrorMatches, regexp.QuoteMeta(tc.err), comment)
		} else {
			c.Assert(err, IsNil, comment)
		}
	}
}

func (s *configTestSuite) TestInvalidTOML(c *C) {
	cfg := &config.Config{}
	err := cfg.LoadFromTOML([]byte(`
		invalid[mydumper.csv]
		delimiter = '\'
		backslash-escape = true
	`))
	c.Assert(err, ErrorMatches, regexp.QuoteMeta("Near line 0 (last key parsed ''): bare keys cannot contain '['"))
}

func (s *configTestSuite) TestTOMLUnusedKeys(c *C) {
	cfg := &config.Config{}
	err := cfg.LoadFromTOML([]byte(`
		[lightning]
		typo = 123
	`))
	c.Assert(err, ErrorMatches, regexp.QuoteMeta("config file contained unknown configuration options: lightning.typo"))
}

func (s *configTestSuite) TestDurationUnmarshal(c *C) {
	duration := config.Duration{}
	err := duration.UnmarshalText([]byte("13m20s"))
	c.Assert(err, IsNil)
	c.Assert(duration.Duration.Seconds(), Equals, 13*60+20.0)
	err = duration.UnmarshalText([]byte("13x20s"))
	c.Assert(err, ErrorMatches, "time: unknown unit x in duration 13x20s")
}

func (s *configTestSuite) TestDurationMarshalJSON(c *C) {
	duration := config.Duration{}
	err := duration.UnmarshalText([]byte("13m20s"))
	c.Assert(err, IsNil)
	c.Assert(duration.Duration.Seconds(), Equals, 13*60+20.0)
	result, err := duration.MarshalJSON()
	c.Assert(err, IsNil)
	c.Assert(string(result), Equals, `"13m20s"`)
}

func (s *configTestSuite) TestLoadConfig(c *C) {
	cfg, err := config.LoadGlobalConfig([]string{"-tidb-port", "sss"}, nil)
	c.Assert(err, ErrorMatches, `invalid value "sss" for flag -tidb-port: parse error`)
	c.Assert(cfg, IsNil)

	cfg, err = config.LoadGlobalConfig([]string{"-V"}, nil)
	c.Assert(err, Equals, flag.ErrHelp)
	c.Assert(cfg, IsNil)

	cfg, err = config.LoadGlobalConfig([]string{"-config", "not-exists"}, nil)
	c.Assert(err, ErrorMatches, ".*(no such file or directory|The system cannot find the file specified).*")
	c.Assert(cfg, IsNil)

	cfg, err = config.LoadGlobalConfig([]string{"--server-mode"}, nil)
	c.Assert(err, ErrorMatches, "If server-mode is enabled, the status-addr must be a valid listen address")
	c.Assert(cfg, IsNil)

	cfg, err = config.LoadGlobalConfig([]string{
		"-L", "debug",
		"-log-file", "/path/to/file.log",
		"-tidb-host", "172.16.30.11",
		"-tidb-port", "4001",
		"-tidb-user", "guest",
		"-tidb-password", "12345",
		"-pd-urls", "172.16.30.11:2379,172.16.30.12:2379",
		"-d", "/path/to/import",
		"-importer", "172.16.30.11:23008",
		"-checksum=false",
	}, nil)
	c.Assert(err, IsNil)
	c.Assert(cfg.App.Config.Level, Equals, "debug")
	c.Assert(cfg.App.Config.File, Equals, "/path/to/file.log")
	c.Assert(cfg.TiDB.Host, Equals, "172.16.30.11")
	c.Assert(cfg.TiDB.Port, Equals, 4001)
	c.Assert(cfg.TiDB.User, Equals, "guest")
	c.Assert(cfg.TiDB.Psw, Equals, "12345")
	c.Assert(cfg.TiDB.PdAddr, Equals, "172.16.30.11:2379,172.16.30.12:2379")
	c.Assert(cfg.Mydumper.SourceDir, Equals, "/path/to/import")
	c.Assert(cfg.TikvImporter.Addr, Equals, "172.16.30.11:23008")
	c.Assert(cfg.PostRestore.Checksum, IsFalse)
	c.Assert(cfg.PostRestore.Analyze, IsTrue)

	taskCfg := config.NewConfig()
	err = taskCfg.LoadFromGlobal(cfg)
	c.Assert(err, IsNil)
	c.Assert(taskCfg.PostRestore.Checksum, IsFalse)
	c.Assert(taskCfg.PostRestore.Analyze, IsTrue)

	taskCfg.Checkpoint.DSN = ""
	taskCfg.Checkpoint.Driver = config.CheckpointDriverMySQL
	err = taskCfg.Adjust()
	c.Assert(err, IsNil)
	c.Assert(taskCfg.Checkpoint.DSN, Equals, "guest:12345@tcp(172.16.30.11:4001)/?charset=utf8mb4&sql_mode='"+mysql.DefaultSQLMode+"'&maxAllowedPacket=67108864&tls=false")

	result := taskCfg.String()
	c.Assert(result, Matches, `.*"pd-addr":"172.16.30.11:2379,172.16.30.12:2379".*`)
}

func (s *configTestSuite) TestDefaultImporterBackendValue(c *C) {
	cfg := config.NewConfig()
	assignMinimalLegalValue(cfg)
	cfg.TikvImporter.Backend = "importer"
	err := cfg.Adjust()
	c.Assert(err, IsNil)
	c.Assert(cfg.App.IndexConcurrency, Equals, 2)
	c.Assert(cfg.App.TableConcurrency, Equals, 6)
}

func (s *configTestSuite) TestDefaultTidbBackendValue(c *C) {
	cfg := config.NewConfig()
	assignMinimalLegalValue(cfg)
	cfg.TikvImporter.Backend = "tidb"
	cfg.App.RegionConcurrency = 123
	err := cfg.Adjust()
	c.Assert(err, IsNil)
	c.Assert(cfg.App.IndexConcurrency, Equals, 123)
	c.Assert(cfg.App.TableConcurrency, Equals, 123)
}

func (s *configTestSuite) TestDefaultCouldBeOverwritten(c *C) {
	cfg := config.NewConfig()
	assignMinimalLegalValue(cfg)
	cfg.TikvImporter.Backend = "importer"
	cfg.App.IndexConcurrency = 20
	cfg.App.TableConcurrency = 60
	err := cfg.Adjust()
	c.Assert(err, IsNil)
	c.Assert(cfg.App.IndexConcurrency, Equals, 20)
	c.Assert(cfg.App.TableConcurrency, Equals, 60)
}

func (s *configTestSuite) TestLoadFromInvalidConfig(c *C) {
	taskCfg := config.NewConfig()
	err := taskCfg.LoadFromGlobal(&config.GlobalConfig{
		ConfigFileContent: []byte("invalid toml"),
	})
	c.Assert(err, ErrorMatches, "Near line 1.*")
}
