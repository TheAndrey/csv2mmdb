package main

import (
	"encoding/csv"
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
	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s <input.csv> <output.mmdb>\n", os.Args[0])
		os.Exit(1)
	}

	inputPath := os.Args[1]
	outputPath := os.Args[2]

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
	file, err := os.Open(inputPath)
	if err != nil {
		log.Fatalf("Unable to open: %v", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)

	// Read header and skip
	if _, err := reader.Read(); err != nil {
		log.Fatalf("CSV read error: %v", err)
	}

	fmt.Printf("Start processing of: %s\n", inputPath)

	line := 1
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
		asName := record[2]

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

		data := mmdbtype.Map{
			"route":   mmdbtype.String(cidrStr),
			"asn":     mmdbtype.Uint32(asn),
			"as_name": mmdbtype.String(asName),
		}

		if err := tree.Insert(ipNet, data); err != nil {
			log.Printf("Error adding %s: %v", cidrStr, err)
		}
	}

	// Write final database to file
	outFile, err := os.Create(outputPath)
	if err != nil {
		log.Fatalf("Unable to create file: %v", err)
	}
	defer outFile.Close()

	_, err = tree.WriteTo(outFile)
	if err != nil {
		log.Fatalf("Error writing file: %v", err)
	}

	fmt.Printf("Database successfully written (Processed %d lines).\n", line-1)
}
