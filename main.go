package main

import (
	"bufio"
	"compress/gzip"
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func main() {
	dst := flag.String("destination", "vpnlist.json", "the destination file path")
	src := flag.String("source", "https://raw.githubusercontent.com/X4BNet/lists_vpn/main/ipv4.txt", "the source URL")
	flag.Parse()

	err := run(*dst, *src)
	if err != nil {
		log.Fatalln(err)
	}
}

func run(dst, src string) error {
	ips, err := readIPs(src)
	if err != nil {
		return err
	}

	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()

	return writeFile(f, dst, ips)
}

func readIPs(src string) ([]string, error) {
	res, err := http.Get(src)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode/100 != 2 {
		return nil, fmt.Errorf("unexpected status code: %d %s", res.StatusCode, res.Status)
	}

	var ips []string
	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		ip := strings.TrimSpace(scanner.Text())
		if ip == "" {
			continue
		}

		ips = append(ips, ip)
	}
	sort.Strings(ips)
	return ips, nil
}

func writeFile(w io.Writer, name string, ips []string) error {
	switch ext := filepath.Ext(name); ext {
	case ".gz":
		zw, err := gzip.NewWriterLevel(w, gzip.BestCompression)
		if err != nil {
			return err
		}
		err = writeFile(zw, name[:len(name)-len(ext)], ips)
		if err != nil {
			_ = zw.Close()
			return err
		}
		return zw.Close()
	case ".csv":
		cw := csv.NewWriter(w)
		err := cw.Write([]string{"id"})
		if err != nil {
			return err
		}
		for _, ip := range ips {
			err = cw.Write([]string{ip})
			if err != nil {
				return err
			}
		}
		cw.Flush()
		return cw.Error()
	case ".json":
		var records []any
		for _, ip := range ips {
			records = append(records, map[string]any{"id": ip})
		}
		e := json.NewEncoder(w)
		e.SetIndent("", "  ")
		return e.Encode(records)
	default:
		return fmt.Errorf("unsupported file extension: %s", ext)
	}
}
