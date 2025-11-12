package main

import (
	"fmt"
	"math"
	"os"
	"strings"
	"text/template" // Changed from "html/template" to "text/template"
	"time"
)

type ChartData struct {
	CandleData string
	ZData      string
	VWZData    string
}

func generateHTMLChart(candles CandleSticks, zScores []float64, vwzScores []float64) {
	var candleData []string
	var zData []string
	var vwzData []string

	for i, c := range candles {
		ms := c.Time.UnixNano() / int64(time.Millisecond)
		candlePoint := fmt.Sprintf("{x: %d, o: %.4f, h: %.4f, l: %.4f, c: %.4f}", ms, c.Open, c.High, c.Low, c.Close)
		candleData = append(candleData, candlePoint)

		if math.IsNaN(zScores[i]) {
			zData = append(zData, fmt.Sprintf("{x: %d, y: null}", ms))
		} else {
			zData = append(zData, fmt.Sprintf("{x: %d, y: %.4f}", ms, zScores[i]))
		}
		if math.IsNaN(vwzScores[i]) {
			vwzData = append(vwzData, fmt.Sprintf("{x: %d, y: null}", ms))
		} else {
			vwzData = append(vwzData, fmt.Sprintf("{x: %d, y: %.4f}", ms, vwzScores[i]))
		}
	}

	candleDataJS := "[" + strings.Join(candleData, ",") + "]"
	zDataJS := "[" + strings.Join(zData, ",") + "]"
	vwzDataJS := "[" + strings.Join(vwzData, ",") + "]"

		tmpl, err := template.ParseFiles("chart.html.template")
		if err != nil {
			fmt.Println("Error parsing template:", err)
			return
		}
	
		data := ChartData{
			CandleData: candleDataJS,
			ZData:      zDataJS,
			VWZData:    vwzDataJS,
		}
	
		file, err := os.Create("chart.html")
		if err != nil {
			fmt.Println("Error creating chart.html:", err)
			return
		}
		defer file.Close()
	
		err = tmpl.Execute(file, data)
		if err != nil {
			fmt.Println("Error executing template:", err)
			return
		}
		fmt.Println("Generated chart.html with zoom & sync (scroll or drag to zoom)")
}
