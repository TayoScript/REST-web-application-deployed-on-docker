package handlers

import (
	"encoding/csv"
	"log"
	"os"
)

// Read the CSV data
func ReadCSV(filename string) ([][]string, error) {
	file, err := os.Open(filename)

	if err != nil {
		log.Println("E: Failed to read data file")
		return nil, err
	}

	csvData := csv.NewReader(file)

	data, err := csvData.ReadAll()

	if err != nil {
		log.Println("E: Failed to parse data file")
		return nil, err
	}

	return data, nil
}
