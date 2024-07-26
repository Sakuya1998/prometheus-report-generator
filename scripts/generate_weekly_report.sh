#!/bin/bash
cd "$(dirname "$0")/.." # Move to the root directory
./prometheus-report-generator -report_type weekly

