# Overview

This Go program scrapes and analyzes a given Shopify website's sections, calculates their usage count, and determines if they are unused. The program generates an Excel file containing this information.

## Prerequisites

Ensure that you have Go installed on your machine. If you haven't, please follow the official installation guide. [Homebrew installation](https://formulae.brew.sh/formula/go)
Install the required packages using the following commands:

```sh
go get -u github.com/gocolly/colly/v2
go get -u github.com/tealeg/xlsx
```

## How to Run

Clone or download the source code to your local machine.
Open a terminal and navigate to the directory containing the Go program (main.go).
Run the program with the following command:

```sh
go run main.go <website>
```

Replace <website> with the URL of the Shopify website you want to analyze, e.g., https://example.com.

## Command-Line Arguments

`<website>`: The URL of the Shopify website you want to analyze (required). It must include the scheme, either http:// or https://. Note: sometimes you need to add `www.` to the beginning of the URL.

## Additional Requirements

Please note that the sections directory must exist in the same directory as the Go program, containing text files representing the Shopify section names you want to analyze. For example: `sections/product.liquid` tells the program to look for sections with the name `product`. You can copy + paste this folder from the built theme files

## Expected Output

The program will generate an Excel file named shopify_sections.xlsx in the same directory as the Go program. The Excel file will contain a sheet named "Sections" with the following columns:

Section Name: The name of the Shopify section.
Usage Count: The number of times the section is used on the website.
Unused: A boolean value indicating if the section is unused.
After successfully generating the Excel file, the program will output:

```sh
Scraping and ranking complete. Saved to shopify_sections.xlsx.
```
