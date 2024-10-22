package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type TestResponse struct {
	Message   string `json:"message"`
	Success   bool   `json:"success"`
	TestId    string `json:"testId"`
	ReportUrl string `json:"reportUrl,omitempty"`
	Status    string `json:"status"`
}

type TestConfig struct {
	TestPlan   string `json:"testPlan"`
	Threads    int    `json:"threads,string"`
	RampUp     int    `json:"rampUp,string"`
	Duration   int    `json:"duration,string"`
	TargetHost string `json:"targetHost"`
}

func main() {
	http.HandleFunc("/", serveHomepage)
	http.HandleFunc("/test-plans", getTestPlans)
	http.HandleFunc("/run-test", runJMeterTest)
	http.Handle("/reports/", http.StripPrefix("/reports/", http.FileServer(http.Dir("jmeter/reports"))))
	http.HandleFunc("/generate-report", generateReport)
	http.HandleFunc("/list-reports", listReports)

	log.Println("Starting server on :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatalf("Server failed: %s", err)
	}
}

func serveHomepage(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmpl.Execute(w, nil)
}

func getTestPlans(w http.ResponseWriter, r *http.Request) {
	files, err := os.ReadDir("jmeter")
	if err != nil {
		http.Error(w, "Failed to read test plans", http.StatusInternalServerError)
		return
	}

	var testPlans []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".jmx" {
			testPlans = append(testPlans, file.Name())
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(testPlans)
}

func runJMeterTest(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var config TestConfig
	err := json.NewDecoder(r.Body).Decode(&config)
	if err != nil {
		json.NewEncoder(w).Encode(TestResponse{
			Success: false,
			Message: "Invalid request body: " + err.Error(),
			Status:  "error",
		})
		return
	}

	// Send initial response indicating the test has started
	json.NewEncoder(w).Encode(TestResponse{
		Success: true,
		Message: "Test started. Please wait...",
		Status:  "pending",
	})

	// Flush the response writer to send the initial response immediately
	if f, ok := w.(http.Flusher); ok {
		f.Flush()
	}

	// Generate a unique identifier for this test run
	timestamp := time.Now().Format("20060102_150405")
	testId := fmt.Sprintf("%s_%s", strings.TrimSuffix(config.TestPlan, ".jmx"), timestamp)

	// Prepare JMeter command
	jmeterPath := "jmeter" // Assumes 'jmeter' is in PATH. Adjust if necessary.
	testPlanPath := filepath.Join("jmeter", config.TestPlan)
	resultPath := filepath.Join("jmeter", "results", testId+".jtl")
	reportPath := filepath.Join("jmeter", "reports", testId)

	cmd := exec.Command(jmeterPath,
		"-n",
		"-t", testPlanPath,
		"-l", resultPath,
		"-e",
		"-o", reportPath,
		"-Jthreads="+fmt.Sprintf("%d", config.Threads),
		"-Jrampup="+fmt.Sprintf("%d", config.RampUp),
		"-Jduration="+fmt.Sprintf("%d", config.Duration),
		"-Jtarget="+config.TargetHost,
	)

	// Run the JMeter test
	output, err := cmd.CombinedOutput()
	if err != nil {
		json.NewEncoder(w).Encode(TestResponse{
			Success: false,
			Message: "Error running JMeter test: " + err.Error() + "\n" + string(output),
			Status:  "error",
		})
		return
	}

	// Generate report URL
	reportUrl := fmt.Sprintf("/reports/%s/index.html", testId)

	// Send final response with the report URL
	json.NewEncoder(w).Encode(TestResponse{
		Success:   true,
		Message:   fmt.Sprintf("Test completed. Config: %+v", config),
		TestId:    testId,
		ReportUrl: reportUrl,
		Status:    "completed",
	})
}

func generateReport(w http.ResponseWriter, r *http.Request) {
	log.Println("Received generate report request")

	var request struct {
		TestId string `json:"testId"`
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Error reading request body: %v", err)
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}
	log.Printf("Received request body: %s", string(body))

	err = json.Unmarshal(body, &request)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if request.TestId == "" {
		log.Println("TestId is empty")
		http.Error(w, "TestId is required", http.StatusBadRequest)
		return
	}

	reportPath := filepath.Join("jmeter", "reports", request.TestId)
	log.Printf("Generating report at: %s", reportPath)

	// ... (code to generate the report) ...

	reportUrl := fmt.Sprintf("/reports/%s/index.html", request.TestId)

	response := TestResponse{
		Success:   true,
		Message:   "Report generated successfully",
		TestId:    request.TestId,
		ReportUrl: reportUrl,
		Status:    "completed",
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.Printf("Error encoding JSON response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

func listReports(w http.ResponseWriter, r *http.Request) {
	reportsDir := "jmeter/reports"
	reports, err := filepath.Glob(filepath.Join(reportsDir, "*"))
	if err != nil {
		http.Error(w, "Failed to list reports", http.StatusInternalServerError)
		return
	}

	var reportList []string
	for _, report := range reports {
		reportName := filepath.Base(report)
		if strings.HasPrefix(reportName, "jmeter_test_plan_") {
			reportList = append(reportList, reportName)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reportList)
}
