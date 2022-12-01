package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/net/html"
)

var inputFileName = flag.String("if", "Students.html", "Specifies path to HTML file to parse")
var outputFileName = flag.String("of", "Students.csv", "Specifies path to CSV output file")

func main() {
	flag.Parse()

	log.Print("MAXIMA/Robocode parser started...")

	htmlFilePath, err := filepath.Abs(*inputFileName)
	if err != nil {
		log.Fatal("Failed to make absolute filepath: ", err)
	}

	log.Printf("Trying to open '%s'...", htmlFilePath)
	htmlFile, err := os.Open(htmlFilePath)
	if err != nil {
		log.Fatal("Failed to open HTML file of interest: ", err)
	}

	log.Print("Done. Parsing...")
	root, err := html.Parse(htmlFile)
	if err != nil {
		log.Fatal("Failed to parse HTML file: ", err)
	}

	rows := make([][]string, 0, 64)

	headerRow := []string{"Student", "#1.1", "#1.2"}
	for i := 2; i < 44; i++ {
		stringNum := fmt.Sprintf("#%d", i)
		headerRow = append(headerRow, stringNum)
	}
	rows = append(rows, headerRow)

	var htmlNodeMarshaller func(*html.Node)
	htmlNodeMarshaller = func(node *html.Node) {
		/* NOTE: on Robocode page useful data is arranged in a table.
		 * We should read this table and just print it as a CSV.
		 */
		if (node.Type == html.ElementNode) && node.Data == "tr" {
			rowContents := make([]string, 0, 64)

			for td := node.FirstChild; td != nil; td = td.NextSibling {
				for _, attr := range td.Attr {
					// log.Printf("Key = %s, Val = %s", attr.Key, attr.Val)
					if attr.Key == "class" {
						if attr.Val == "sorting_1" {
							/* NOTE: student info */
						studentFor:
							for a := td.FirstChild; a != nil; a = a.NextSibling {
								for _, attr := range a.Attr {
									if attr.Key == "href" {
										rowContents = append(rowContents, a.FirstChild.Data)
										break studentFor
									}
								}
							}
						} else if attr.Val == " small" {
							/* NOTE: student's scores */
						scoreFor:
							for div := td.FirstChild; div != nil; div = div.NextSibling {
								for a := div.FirstChild; a != nil; a = a.NextSibling {
									for span := a.FirstChild; span != nil; span = span.NextSibling {
										for _, attr := range span.Attr {
											if attr.Key == "title" {
												if attr.Val == "Completed, evaluation is completed" {
													rowContents = append(rowContents, span.FirstChild.Data)
												} else {
													rowContents = append(rowContents, "")
												}
												break scoreFor
											}
										}
									}
								}
							}
						}
					}
				}
			}
			if len(rowContents) != 0 {
				rows = append(rows, rowContents)
			}
		} else {
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				htmlNodeMarshaller(child)
			}
		}
	}
	htmlNodeMarshaller(root)

	log.Print("Done. Generating CSV...")
	var outFile *os.File

	csvFilePath, err := filepath.Abs(*outputFileName)
	if err != nil {
		log.Print("Fail to get absolute path of output file. Falling back to stdout...")
		outFile = os.Stdout
	} else {
		outFile, err = os.Create(csvFilePath)
		if err != nil {
			log.Print("Fail to create output file. Falling back to stdout...")
			outFile = os.Stdout
		}
	}

	csvWriter := csv.NewWriter(outFile)
	csvWriter.WriteAll(rows)

	log.Printf("Done. Result is saved in '%s'", outFile.Name())
	log.Print("MAXIMA/Robocode parser finished. Goodbye :)")
}
