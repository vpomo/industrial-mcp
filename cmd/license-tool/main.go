package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vpomo/industrial-mcp/pkg/license"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "create":
		createLicense(os.Args[2:])
	case "verify":
		verifyLicense(os.Args[2:])
	case "export-hwid":
		exportHWID()
	default:
		printUsage()
		os.Exit(1)
	}
}

func createLicense(args []string) {
	flagSet := flag.NewFlagSet("create", flag.ExitOnError)
	hwidFlag := flagSet.String("hardware-hash", "", "Hardware ID hash")
	expiresFlag := flagSet.String("expires", "", "Expiration date (YYYY-MM-DD)")
	featuresFlag := flagSet.String("features", "basic", "Comma-separated features")
	outputFlag := flagSet.String("output", "license.dat", "Output file path")
	privateKeyFlag := flagSet.String("private-key", "pkg/license/keys/private.pem", "Path to private key")

	flagSet.Parse(args)

	privateKey, err := os.ReadFile(*privateKeyFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading private key: %v\n", err)
		os.Exit(1)
	}

	generator, err := license.NewLicenseGenerator(privateKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating generator: %v\n", err)
		os.Exit(1)
	}

	expiresAt, err := time.Parse("2006-01-02", *expiresFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid date format. Use YYYY-MM-DD\n", err)
		os.Exit(1)
	}

	features := parseFeatures(*featuresFlag)

	lf, err := generator.Create(*hwidFlag, expiresAt, features, "awwantil Licensing")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating license: %v\n", err)
		os.Exit(1)
	}

	if err := lf.Save(*outputFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error saving license: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("License created successfully:\n")
	fmt.Printf("  Hardware Hash: %s\n", lf.HardwareHash)
	fmt.Printf("  Expires: %s\n", lf.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Features: %v\n", lf.Features)
	fmt.Printf("  Output: %s\n", *outputFlag)
}

func verifyLicense(args []string) {
	flagSet := flag.NewFlagSet("verify", flag.ExitOnError)
	fileFlag := flagSet.String("file", "license.dat", "License file path")
	publicKeyFlag := flagSet.String("public-key", "pkg/license/keys/public.pem", "Path to public key")

	flagSet.Parse(args)

	lf := &license.LicenseFile{}
	if err := lf.Load(*fileFlag); err != nil {
		fmt.Fprintf(os.Stderr, "Error loading license: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("License file: %s\n", *fileFlag)
	fmt.Printf("  Version: %d\n", lf.Version)
	fmt.Printf("  Hardware Hash: %s\n", lf.HardwareHash)
	fmt.Printf("  Issued: %s\n", lf.IssuedAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Expires: %s\n", lf.ExpiresAt.Format("2006-01-02 15:04:05"))
	fmt.Printf("  Features: %v\n", lf.Features)
	fmt.Printf("  Issuer: %s\n", lf.Issuer)

	if lf.IsExpired() {
		fmt.Printf("  Status: EXPIRED\n")
	} else {
		fmt.Printf("  Status: VALID\n")
		fmt.Printf("  Days remaining: %d\n", int(time.Until(lf.ExpiresAt).Hours()/24))
	}

	if *publicKeyFlag != "" {
		publicKey, err := os.ReadFile(*publicKeyFlag)
		if err == nil {
			crypto, err := license.NewRSACryptoFromPEM(publicKey)
			if err == nil {
				payload := lf.Payload()
				if crypto.Verify(payload, lf.Signature) {
					fmt.Printf("  Signature: VALID\n")
				} else {
					fmt.Printf("  Signature: INVALID\n")
				}
			}
		}
	}
}

func exportHWID() {
	hw, err := license.GetHardwareInfo()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting hardware info: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Hardware ID: %s\n", hw.Hash())
	fmt.Printf("  CPUID: %s\n", hw.CPUID)
	fmt.Printf("  MAC: %s\n", hw.MACAddr)
	fmt.Printf("  VolumeID: %s\n", hw.VolumeID)
	fmt.Printf("  Motherboard: %s\n", hw.Motherboard)
}

func parseFeatures(featuresStr string) []string {
	if featuresStr == "" {
		return []string{"basic"}
	}
	parts := strings.Split(featuresStr, ",")
	var features []string
	for _, p := range parts {
		f := strings.TrimSpace(p)
		if f != "" {
			features = append(features, f)
		}
	}
	if len(features) == 0 {
		features = []string{"basic"}
	}
	return features
}

func printUsage() {
	fmt.Println("License Tool")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  license-tool create [options]")
	fmt.Println("  license-tool verify [options]")
	fmt.Println("  license-tool export-hwid")
	fmt.Println("")
	fmt.Println("Commands:")
	fmt.Println("  create         Create a new license file")
	fmt.Println("  verify         Verify an existing license file")
	fmt.Println("  export-hwid    Export hardware ID of current machine")
	fmt.Println("")
	fmt.Println("Create options:")
	fmt.Println("  --hardware-hash string   Hardware ID hash")
	fmt.Println("  --expires string         Expiration date (YYYY-MM-DD)")
	fmt.Println("  --features string        Comma-separated features (default: basic)")
	fmt.Println("  --output string          Output file path (default: license.dat)")
	fmt.Println("  --private-key string     Path to private key")
	fmt.Println("")
	fmt.Println("Verify options:")
	fmt.Println("  --file string           License file path (default: license.dat)")
	fmt.Println("  --public-key string     Path to public key")
}
