package main

import (
	"bufio"
	"fmt"
	"hash/fnv"
	"os"
	"strconv"
	"strings"
	"time"

	"ourpackage/bloomfilter"
	"ourpackage/cms"
	"ourpackage/hyperloglog"
	"ourpackage/lsh"
	"ourpackage/minhash"
)

// LogEntry represents a parsed log entry
type LogEntry struct {
	Timestamp time.Time
	IP        string
	UserID    string
	SessionID string
	Path      string
	Status    int
	Message   string
}

// LogAnalyzer uses probabilistic data structures to analyze logs
type LogAnalyzer struct {
	deduper        *bloomfilter.BloomFilter
	pathCounter    *cms.CountMinSketch
	userCounter    *hyperloglog.HyperLogLog
	sessionCounter *hyperloglog.HyperLogLog
	errorMinhash   *minhash.MinHash
	errorLSH       *lsh.LSH
	errorMessages  map[int]LogEntry
	nextErrorID    int
}

// NewLogAnalyzer creates a new log analyzer with initialized data structures
func NewLogAnalyzer() *LogAnalyzer {
	// Initialize with reasonable defaults for a medium-sized log analysis
	return &LogAnalyzer{
		deduper:        bloomfilter.New(1000000, 0.01), // 1M entries, 1% error rate
		pathCounter:    cms.New(10000, 5),              // Track up to 10K paths with 5 hash functions
		userCounter:    hyperloglog.New(14),            // 2^14 registers
		sessionCounter: hyperloglog.New(14),            // 2^14 registers
		errorMinhash:   minhash.New(100),               // 100 hash functions for error similarity
		errorLSH:       lsh.New(100, 20, 5),            // bands=20, rows=5 for LSH
		errorMessages:  make(map[int]LogEntry),
		nextErrorID:    0,
	}
}

// Hash generates a hash value for string input
func hash(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

// ProcessLogEntry processes a single log entry through all data structures
func (la *LogAnalyzer) ProcessLogEntry(entry LogEntry) {
	// Create a unique key for deduplication
	entryKey := fmt.Sprintf("%s-%s-%s-%s-%d",
		entry.Timestamp.Format(time.RFC3339),
		entry.IP,
		entry.UserID,
		entry.Path,
		entry.Status)

	// Check if we've seen this exact entry before
	if la.deduper.Test([]byte(entryKey)) {
		return // Skip duplicate entries
	}

	// Add to Bloom filter to mark as seen
	la.deduper.Add([]byte(entryKey))

	// Increment path counter in Count-Min Sketch
	la.pathCounter.Add([]byte(entry.Path), 1)

	// Add user and session to HyperLogLog for cardinality estimation
	la.userCounter.Add([]byte(entry.UserID))
	la.sessionCounter.Add([]byte(entry.SessionID))

	// For error messages (status >= 400), process for similarity analysis
	if entry.Status >= 400 {
		// Generate MinHash signature for error message
		la.errorMinhash.Reset()
		la.errorMinhash.Update([]byte(entry.Message))
		signature := la.errorMinhash.Signature()

		// Store error in our collection
		la.errorMessages[la.nextErrorID] = entry

		// Add to LSH for similarity queries
		la.errorLSH.Insert(la.nextErrorID, signature)

		la.nextErrorID++
	}
}

// ParseLogLine converts a raw log line into a structured LogEntry
func ParseLogLine(line string) (LogEntry, error) {
	// This is a simplified parser for demonstration
	// In a real system, you'd use a more robust parser

	// Example format: [2023-04-15T10:20:30Z] 192.168.1.1 user123 session456 /api/items 200 "Request successful"
	parts := strings.SplitN(line, " ", 7)
	if len(parts) != 7 {
		return LogEntry{}, fmt.Errorf("invalid log format")
	}

	// Parse timestamp
	ts, err := time.Parse("2006-01-02T15:04:05Z", strings.Trim(parts[0], "[]"))
	if err != nil {
		return LogEntry{}, fmt.Errorf("invalid timestamp: %v", err)
	}

	// Parse status code
	status, err := strconv.Atoi(parts[5])
	if err != nil {
		return LogEntry{}, fmt.Errorf("invalid status code: %v", err)
	}

	// Extract message
	message := strings.Trim(parts[6], "\"")

	return LogEntry{
		Timestamp: ts,
		IP:        parts[1],
		UserID:    parts[2],
		SessionID: parts[3],
		Path:      parts[4],
		Status:    status,
		Message:   message,
	}, nil
}

// GetTopPaths returns the estimated most frequent paths
func (la *LogAnalyzer) GetTopPaths(paths []string, n int) []string {
	type PathCount struct {
		Path  string
		Count uint64
	}

	// Get count estimates for all paths
	pathCounts := make([]PathCount, 0, len(paths))
	for _, path := range paths {
		count := la.pathCounter.Estimate([]byte(path))
		pathCounts = append(pathCounts, PathCount{Path: path, Count: count})
	}

	// Sort paths by count (descending)
	for i := 0; i < len(pathCounts); i++ {
		for j := i + 1; j < len(pathCounts); j++ {
			if pathCounts[i].Count < pathCounts[j].Count {
				pathCounts[i], pathCounts[j] = pathCounts[j], pathCounts[i]
			}
		}
	}

	// Return top N paths
	result := make([]string, 0, n)
	for i := 0; i < n && i < len(pathCounts); i++ {
		result = append(result, pathCounts[i].Path)
	}

	return result
}

// GetUniqueUserCount returns the estimated number of unique users
func (la *LogAnalyzer) GetUniqueUserCount() uint64 {
	return la.userCounter.Estimate()
}

// GetUniqueSessionCount returns the estimated number of unique sessions
func (la *LogAnalyzer) GetUniqueSessionCount() uint64 {
	return la.sessionCounter.Estimate()
}

// FindSimilarErrors finds errors similar to the given one
func (la *LogAnalyzer) FindSimilarErrors(errorMsg string, threshold float64) []LogEntry {
	// Create MinHash signature for the query error
	la.errorMinhash.Reset()
	la.errorMinhash.Update([]byte(errorMsg))
	querySignature := la.errorMinhash.Signature()

	// Get candidate matches from LSH
	candidateIDs := la.errorLSH.Query(querySignature)

	// Refine candidates by calculating actual Jaccard similarity
	similarErrors := make([]LogEntry, 0)
	for _, id := range candidateIDs {
		entry := la.errorMessages[id]

		// Calculate actual similarity
		la.errorMinhash.Reset()
		la.errorMinhash.Update([]byte(entry.Message))
		candidateSignature := la.errorMinhash.Signature()

		similarity := minhash.JaccardSimilarity(querySignature, candidateSignature)
		if similarity >= threshold {
			similarErrors = append(similarErrors, entry)
		}
	}

	return similarErrors
}

// GenerateReport creates a summary report of the log analysis
func (la *LogAnalyzer) GenerateReport(knownPaths []string) string {
	var report strings.Builder

	report.WriteString("=== Log Analysis Report ===\n\n")

	// Unique user and session counts
	report.WriteString(fmt.Sprintf("Estimated unique users: %d\n", la.GetUniqueUserCount()))
	report.WriteString(fmt.Sprintf("Estimated unique sessions: %d\n\n", la.GetUniqueSessionCount()))

	// Top paths
	report.WriteString("Top 5 paths:\n")
	topPaths := la.GetTopPaths(knownPaths, 5)
	for i, path := range topPaths {
		count := la.pathCounter.Estimate([]byte(path))
		report.WriteString(fmt.Sprintf("%d. %s (approx %d hits)\n", i+1, path, count))
	}
	report.WriteString("\n")

	// Error statistics
	report.WriteString(fmt.Sprintf("Total unique error types: %d\n\n", len(la.errorMessages)))

	return report.String()
}

func main() {
	// Create a new log analyzer
	analyzer := NewLogAnalyzer()

	// List of known paths in our system (for reporting)
	knownPaths := []string{
		"/api/users",
		"/api/products",
		"/api/orders",
		"/api/login",
		"/api/logout",
		"/api/search",
		"/api/payment",
		"/api/profile",
		"/api/cart",
		"/api/checkout",
	}

	// Process log file
	file, err := os.Open("access.log")
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		return
	}
	defer file.Close()

	// Track statistics
	linesProcessed := 0
	errorLogs := 0

	// Read and process each line
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		entry, err := ParseLogLine(line)
		if err != nil {
			fmt.Printf("Error parsing line: %v\n", err)
			continue
		}

		analyzer.ProcessLogEntry(entry)
		linesProcessed++

		if entry.Status >= 400 {
			errorLogs++
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading log file: %v\n", err)
		return
	}

	// Generate and print report
	fmt.Printf("Processed %d log lines (%d errors)\n\n", linesProcessed, errorLogs)
	fmt.Println(analyzer.GenerateReport(knownPaths))

	// Demonstrate finding similar errors
	if errorLogs > 0 {
		fmt.Println("=== Similar Error Analysis ===")
		sampleError := "Database connection timeout: failed to connect after 30 seconds"
		fmt.Printf("Finding errors similar to: \"%s\"\n", sampleError)

		similarErrors := analyzer.FindSimilarErrors(sampleError, 0.7) // 70% similarity threshold
		fmt.Printf("Found %d similar errors\n", len(similarErrors))

		// Print first few similar errors
		for i, err := range similarErrors {
			if i >= 3 {
				break
			}
			fmt.Printf("  - [%s] %s\n", err.Timestamp.Format(time.RFC3339), err.Message)
		}
	}
}
