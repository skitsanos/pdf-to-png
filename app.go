package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/unidoc/unipdf/v3/model"
	"github.com/unidoc/unipdf/v3/render"
	"os"
	"path/filepath"
)

type pdfInfo struct {
	Source string      `json:"source"`
	Pages  int         `json:"pages"`
	Author string      `json:"author,omitempty"`
	Errors []errorInfo `json:"errors,omitempty"`
}

type errorInfo struct {
	PageNum int    `json:"page_num"`
	Reason  string `json:"reason"`
}

func main() {
	source := flag.String("source", "", "Path to the source PDF document")
	destination := flag.String("destination", ".", "Folder where images will be stored")
	width := flag.Int("width", 400, "Width of the image")
	jsonOutput := flag.Bool("json", false, "Output details in JSON format")
	flag.Parse()

	if *source == "" || *destination == "" {
		fmt.Println("Both '--source' and '--destination' arguments are required.")
		flag.Usage()
		os.Exit(1)
	}

	if _, err := os.Stat(*source); os.IsNotExist(err) {
		fmt.Printf("Error: File %s does not exist\n", *source)
		os.Exit(1)
	}

	if _, err := os.Stat(*destination); os.IsNotExist(err) {
		// folder does not exist, create it
		err := os.Mkdir(*destination, os.ModePerm)
		if err != nil {
			return
		}
	}

	f, err := os.Open(*source)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	}(f)

	// Load the PDF document
	pdfDoc, err := model.NewPdfReader(f)
	if err != nil {
		fmt.Println("Error opening PDF:", err)
		os.Exit(1)
	}

	numPages, err := pdfDoc.GetNumPages()
	if err != nil {
		fmt.Println("Error opening PDF:", err)
		os.Exit(1)
	}

	pdfInfo := pdfInfo{
		Source: *source,
		Pages:  numPages,
	}

	absPath, err := filepath.Abs(*destination)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Println("Destination:", absPath)
	fmt.Println("Pages:", numPages)

	//basename := strings.TrimSuffix(filepath.Base(*source), filepath.Ext(*source))

	device := render.NewImageDevice()
	device.OutputWidth = *width

	// Convert each page of the PDF to an image
	for pageNum := 1; pageNum <= numPages; pageNum++ {
		page, err := pdfDoc.GetPage(pageNum)

		if err != nil {
			fmt.Println("Error opening PDF:", err)
			pdfInfo.Errors = append(pdfInfo.Errors, errorInfo{
				PageNum: pageNum,
				Reason:  fmt.Sprintf("Error getting page %d: %v", pageNum, err),
			})
			continue
		}

		destFile := fmt.Sprintf("%s/page_%d.png", absPath, pageNum)
		fmt.Sprintf("%s", destFile)

		if err = device.RenderToPath(page, destFile); err != nil {
			pdfInfo.Errors = append(pdfInfo.Errors, errorInfo{
				PageNum: pageNum,
				Reason:  fmt.Sprintf("Error creating image from page %d: %v", pageNum, err),
			})
			continue
		}
	}

	if *jsonOutput {
		b, err := json.MarshalIndent(pdfInfo, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling PDF info to JSON:", err)
			return
		}
		fmt.Println(string(b))
	}
}
