module github.com/pingcap/tidb-lightning

go 1.13

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/DATA-DOG/go-sqlmock v1.4.1
	github.com/coreos/go-semver v0.3.0
	github.com/go-sql-driver/mysql v1.5.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/mock v1.4.0
	github.com/joho/sqltocsv v0.0.0-20190824231449-5650f27fd5b6
	github.com/pingcap/check v0.0.0-20191216031241-8a5a85928f12
	github.com/pingcap/errors v0.11.5-0.20190809092503-95897b64e011
	github.com/pingcap/failpoint v0.0.0-20191029060244-12f4ac2fd11d
	github.com/pingcap/kvproto v0.0.0-20200108025604-a4dc183d2af5
	github.com/pingcap/log v0.0.0-20191012051959-b742a5d432e9
	github.com/pingcap/parser v0.0.0-20200109073933-a9496438d77d
	github.com/pingcap/tidb v1.1.0-beta.0.20210831174426-f2b55a925a01
	github.com/pingcap/tidb-tools v4.0.0-beta+incompatible
	github.com/prometheus/client_golang v1.4.0
	github.com/prometheus/client_model v0.2.0
	github.com/satori/go.uuid v1.2.0
	github.com/shurcooL/httpgzip v0.0.0-20190720172056-320755c1c1b0
	go.uber.org/zap v1.13.0
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e
	golang.org/x/text v0.3.2
	google.golang.org/grpc v1.25.1
	modernc.org/mathutil v1.0.0
)

replace github.com/pingcap/pd => github.com/nolouch/pd v1.1.0-beta.0.20210831131047-1b250f3ee060
