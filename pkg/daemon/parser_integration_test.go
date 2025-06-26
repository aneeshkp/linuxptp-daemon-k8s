package daemon

import (
	"testing"

	"github.com/k8snetworkplumbingwg/linuxptp-daemon/pkg/config"
	"github.com/k8snetworkplumbingwg/linuxptp-daemon/pkg/parser"
	"github.com/stretchr/testify/assert"
	ptpv1 "github.com/k8snetworkplumbingwg/ptp-operator/api/v1"
)

func TestParserIntegration(t *testing.T) {
	tests := []struct {
		name           string
		processName    string
		logLine        string
		expectedParser bool
	}{
		{
			name:           "PTP4L process with parser",
			processName:    ptp4lProcessName,
			logLine:        "ptp4l[5196819.100]: [ptp4l.0.config] master offset -4 s2 freq -26835 path delay 525",
			expectedParser: true,
		},
		{
			name:           "PHC2SYS process with parser",
			processName:    phc2sysProcessName,
			logLine:        "phc2sys[3560354.300]: [ptp4l.0.config] CLOCK_REALTIME rms 4 max 4 freq -76829 +/- 0 delay 1085 +/- 0",
			expectedParser: true,
		},
		{
			name:           "TS2PHC process without parser",
			processName:    ts2phcProcessName,
			logLine:        "ts2phc[82674.465]: [ts2phc.0.cfg] ens2f1 master offset 0 s2 freq -0",
			expectedParser: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test process
			process := &ptpProcess{
				name:   tt.processName,
				parser: nil,
			}
			
			// Set the parser for the process
			process.setParserForProcess()
			
			// Check if parser was set correctly
			if tt.expectedParser {
				assert.NotNil(t, process.parser, "Parser should be set for %s", tt.processName)
				
				// Test that the parser can extract from the log line
				metrics, ptpEvent, err := process.parser.Extract(tt.logLine)
				assert.NoError(t, err, "Parser should not return error for valid log line")
				
				// At least one of metrics or event should be extracted
				assert.True(t, metrics != nil || ptpEvent != nil, "Parser should extract either metrics or event")
			} else {
				assert.Nil(t, process.parser, "Parser should not be set for %s", tt.processName)
			}
		})
	}
}

func TestProcessPTPMetricsWithParser(t *testing.T) {
	// Test that processPTPMetrics works correctly with parser
	process := &ptpProcess{
		name:   ptp4lProcessName,
		parser: nil,
	}
	
	// Set up test data
	process.setParserForProcess()
	process.messageTag = "[ptp4l.0.config]"
	process.ifaces = config.IFaces{
		{Name: "master", PhcId: 0},
	}
	
	// Test log line
	logLine := "ptp4l[5196819.100]: [ptp4l.0.config] master offset -4 s2 freq -26835 path delay 525"
	
	// This should not panic and should process the log line
	process.processPTPMetrics(logLine)
	
	// Verify that metrics were collected
	assert.True(t, process.hasCollectedMetrics, "Metrics should be collected")
}

func TestSetParserForProcess(t *testing.T) {
	tests := []struct {
		name           string
		processName    string
		expectedParser string
	}{
		{
			name:           "PTP4L process",
			processName:    ptp4lProcessName,
			expectedParser: "PTP4L",
		},
		{
			name:           "PHC2SYS process",
			processName:    phc2sysProcessName,
			expectedParser: "PHC2SYS",
		},
		{
			name:           "Unknown process",
			processName:    "unknown",
			expectedParser: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			process := &ptpProcess{
				name:   tt.processName,
				parser: nil,
			}
			
			process.setParserForProcess()
			
			if tt.expectedParser != "" {
				assert.NotNil(t, process.parser, "Parser should be set for %s", tt.processName)
				assert.Equal(t, tt.expectedParser, process.parser.ProcessName(), "Parser process name should match")
			} else {
				assert.Nil(t, process.parser, "Parser should not be set for unknown process")
			}
		})
	}
} 