// Copyright 2019 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"

	"github.com/google/sredocs/charter"
	"github.com/google/sredocs/exporter/bigquery"
	"github.com/google/sredocs/postmortem"
	"github.com/google/sredocs/source/drive"
)

var (
	mode                = flag.String("mode", "parse", "download, parse or upload.")
	parseKind           = flag.String("parse_kind", "auto", "auto, charter or postmortem.")
	parsePath           = flag.String("parse_path", "", "Path with documents to be parsed.")
	parseOutputPath     = flag.String("parse_output_path", "", "Path to save parser output to.")
	cloudCredentials    = flag.String("cloud_credentials", "", "Path to service account credentials in JSON.")
	downloadFolder      = flag.String("download_folder", "", "Folder to download.")
	downloadDestination = flag.String("download_destination", "", "Path to download to.")
	sourceToUpload      = flag.String("upload_path", "", "Path with CSV files to be uploaded to BigQuery.")
	bigqueryProject     = flag.String("upload_project", "", "GCP project with BigQuery enabled.")
	bigqueryDataset     = flag.String("upload_dataset", "", "BigQuery dataset to be created/updated.")
	bigqueryTable       = flag.String("upload_table", "", "BigQuery table to be created/updated.")
	bigqueryTruncate    = flag.Bool("upload_truncate", false, "BigQuery existing table content will be overriden.")
)

func main() {
	flag.Parse()
	if *mode == "" {
		log.Fatalf("mode must be set to download, parse or upload.")
	}
	switch *mode {
	default:
		log.Fatalf("Unknown mode. It must be set to download, parse or upload.")
	case "download":
		if *cloudCredentials == "" || *downloadFolder == "" || *downloadDestination == "" {
			log.Fatalf("cloud_credentials, download_folder and download_destination must be set in download mode.")
		}
		err := drive.Download(*cloudCredentials, *downloadFolder, *downloadDestination)
		if err != nil {
			log.Fatal(err)
		}
	case "upload":
		if *cloudCredentials == "" || *bigqueryProject == "" || *sourceToUpload == "" || *bigqueryDataset == "" || *bigqueryTable == "" {
			log.Fatalf("cloud_credentials, upload_path, upload_project, upload_dataset and upload_table must be set in upload mode.")
		}
		err := bigquery.Upload(*cloudCredentials, *bigqueryProject, *sourceToUpload, *bigqueryDataset, *bigqueryTable, *bigqueryTruncate)
		if err != nil {
			log.Fatal(err)
		}
	case "parse":
		if *parseKind == "" || *parsePath == "" || *parseOutputPath == "" {
			log.Fatalf("parse_kind, parse_path and parse_output_path must be set in parse mode.")
		}

		files, err := ioutil.ReadDir(*parsePath)
		if err != nil {
			log.Fatal(err)
		}

		if len(files) == 0 {
			log.Fatalf("No files found in %d", parsePath)
		}

		for _, f := range files {
			content, err := ioutil.ReadFile(filepath.Join(*parsePath, f.Name()))
			if err != nil {
				log.Fatal(err)
			}

			s := strings.ToLower(f.Name())
			name := fmt.Sprintf("%s.csv", f.Name())

			switch *parseKind {
			case "auto":
				switch {
				case strings.Contains(s, "charter"):
					parseCharter(content, name)
				case strings.Contains(s, "postmortem"):
					parsePostmortem(content, name)
				default:
					continue
				}
			case "charter":
				parseCharter(content, name)
			case "postmortem":
				parsePostmortem(content, name)
			default:
				log.Fatalf("Unsupported -parse_kind, use auto, charter or postmortem")
			}
		}
	}
}

func parseCharter(content []byte, name string) {
	csv, err := charter.Parse(charter.Fields, content)
	if err != nil {
		log.Println(err)
	}
	charter.Save(csv, filepath.Join(*parseOutputPath, name))
}

func parsePostmortem(content []byte, name string) {
	csv, err := postmortem.Parse(postmortem.Fields, content)
	if err != nil {
		log.Println(err)
	}
	postmortem.Save(csv, filepath.Join(*parseOutputPath, name))
}
