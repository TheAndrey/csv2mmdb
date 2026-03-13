package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"

	"github.com/maxmind/mmdbwriter"
	"github.com/maxmind/mmdbwriter/mmdbtype"
)

func main() {
	namesPath := flag.String("names", "", "CSV file with ASN and AS name (required)")
	routesPath := flag.String("routes", "", "CSV file with CIDR and ASN (required)")
	outputPath := flag.String("output", "", "Output MMDB file path (required)")
	flag.Parse()

	if *routesPath == "" || *namesPath == "" || *outputPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: -names <asn+name.csv> -routes <cidr+asn.csv> -output <output.mmdb>")
		os.Exit(1)
	}

	asNames, err := loadASNames(*namesPath)
	if err != nil {
		log.Fatalf("Failed to load names file: %v", err)
	}
	fmt.Printf("Loaded %d AS names\n", len(asNames))

	/* Init database */
	tree, err := mmdbwriter.New(mmdbwriter.Options{
		DatabaseType: "Custom-ASN-Database",
		RecordSize:   28,
		Description: map[string]string{
			"en": "Custom ASN and CIDR database",
		},
	})
	if err != nil {
		log.Fatalf("Tree initialization error: %v", err)
	}

	// Read CSV
	file, err := os.Open(*routesPath)
	if err != nil {
		log.Fatalf("Unable to open: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header and skip
	if _, err := reader.Read(); err != nil {
		log.Fatalf("CSV read error: %v", err)
	}

	fmt.Printf("Start processing of: %s\n", *routesPath)

	line := 1
	recordCount := 0
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		line++
		if err != nil {
			log.Fatal(err)
		}

		cidrStr := record[0]
		asnStr := record[1]

		// Parse CIDR (supports IPv4/IPv6)
		_, ipNet, err := net.ParseCIDR(cidrStr)
		if err != nil {
			log.Printf("Skip invalid CIDR '%s' on line %d: %v", cidrStr, line, err)
			continue
		}

		// Convert ASN to number
		asn, err := strconv.ParseUint(asnStr, 10, 32)
		if err != nil {
			log.Printf("Invalid ASN number '%s' on line %d: %v", asnStr, line, err)
			continue
		}

		asName, ok := asNames[uint32(asn)]
		if !ok {
			log.Printf("AS name not found for %d, skipping", asn)
			continue
		}

		data := mmdbtype.Map{
			"route":   mmdbtype.String(cidrStr),
			"asn":     mmdbtype.Uint32(uint32(asn)),
			"as_name": asName,
		}

		if err := tree.Insert(ipNet, data); err == nil {
			recordCount++
		} else {
			log.Printf("Error adding %s: %v", cidrStr, err)
		}
	}

	// Write final database to file
	outFile, err := os.Create(*outputPath)
	if err != nil {
		log.Fatalf("Unable to create file: %v", err)
	}
	defer outFile.Close()

	_, err = tree.WriteTo(outFile)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}

	fmt.Printf("Database successfully written (processed %d lines, %d records created).\n", line-1, recordCount)
}

// Generates AS name map
func loadASNames(path string) (map[uint32]mmdbtype.String, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	if _, err := reader.Read(); err != nil { // skip header
		return nil, err
	}

	names := make(map[uint32]mmdbtype.String)

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		asn, err := strconv.ParseUint(record[0], 10, 32)
		if err != nil {
			log.Printf("Invalid ASN number '%s'", record[0])
			continue
		}
		names[uint32(asn)] = mmdbtype.String(record[1])
	}

	return names, nil
}
