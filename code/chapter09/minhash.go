package main

import (
	"bufio"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/spaolacci/murmur3"
)

// MinHash implementation

// MinHash represents a MinHash signature generator
type MinHash struct {
	numHashes int
	seeds     []uint32
}

// New creates a new MinHash with the specified number of hash functions
func NewMinHash(numHashes int) *MinHash {
	seeds := make([]uint32, numHashes)
	for i := 0; i < numHashes; i++ {
		seeds[i] = uint32(i + 1) // Simple seed generation
	}

	return &MinHash{
		numHashes: numHashes,
		seeds:     seeds,
	}
}

// Signature generates a MinHash signature for a set of strings
func (mh *MinHash) Signature(set []string) []uint32 {
	signature := make([]uint32, mh.numHashes)

	// Initialize with max uint32 values
	for i := range signature {
		signature[i] = ^uint32(0) // max uint32
	}

	// Update signature for each element in the set
	for _, s := range set {
		for i, seed := range mh.seeds {
			hash := murmur3.Sum32WithSeed([]byte(s), seed)
			if hash < signature[i] {
				signature[i] = hash
			}
		}
	}

	return signature
}

// Similarity calculates the estimated Jaccard similarity between two signatures
func (mh *MinHash) Similarity(sig1, sig2 []uint32) float64 {
	if len(sig1) != mh.numHashes || len(sig2) != mh.numHashes {
		return 0.0
	}

	// Count matching elements
	matches := 0
	for i := 0; i < mh.numHashes; i++ {
		if sig1[i] == sig2[i] {
			matches++
		}
	}

	return float64(matches) / float64(mh.numHashes)
}

// DocumentToSet converts a document to a set of k-shingles
func DocumentToSet(r io.Reader, k int) []string {
	result := make(map[string]struct{}) // Use map as a set

	scanner := bufio.NewScanner(r)
	scanner.Split(bufio.ScanWords)

	// Collect all words
	words := []string{}
	for scanner.Scan() {
		words = append(words, strings.ToLower(scanner.Text()))
	}

	// Generate k-shingles
	if len(words) < k {
		return []string{}
	}

	for i := 0; i <= len(words)-k; i++ {
		shingle := strings.Join(words[i:i+k], " ")
		result[shingle] = struct{}{}
	}

	// Convert map to slice
	uniqueShingles := make([]string, 0, len(result))
	for shingle := range result {
		uniqueShingles = append(uniqueShingles, shingle)
	}

	return uniqueShingles
}

// LSH Implementation
// LSH represents a Locality Sensitive Hashing index
type LSH struct {
	bands        int
	rows         int
	hashTables   []map[string][]int
	minHash      *MinHash
	numDocuments int
}

// New creates a new LSH index
func NewLSH(bands, rows int) *LSH {
	hashTables := make([]map[string][]int, bands)
	for i := range hashTables {
		hashTables[i] = make(map[string][]int)
	}

	return &LSH{
		bands:      bands,
		rows:       rows,
		hashTables: hashTables,
		minHash:    NewMinHash(bands * rows),
	}
}

// AddDocument adds a document to the LSH index
func (lsh *LSH) AddDocument(docID int, shingles []string) {
	// Generate signature
	signature := lsh.minHash.Signature(shingles)

	// Split signature into bands
	for i := 0; i < lsh.bands; i++ {
		start := i * lsh.rows
		end := start + lsh.rows
		if end > len(signature) {
			end = len(signature)
		}

		// Create a band signature
		bandSig := signature[start:end]

		// Convert band to string representation
		bandKey := bandToString(bandSig)

		// Add document to the band's bucket
		lsh.hashTables[i][bandKey] = append(lsh.hashTables[i][bandKey], docID)
	}

	lsh.numDocuments++
}

// FindSimilar finds similar documents to the query
func (lsh *LSH) FindSimilar(shingles []string, threshold float64) map[int]float64 {
	// Generate signature for query
	signature := lsh.minHash.Signature(shingles)

	// Track candidates and actual similarities
	candidates := make(map[int]struct{})
	similarities := make(map[int]float64)

	// Find candidates from each band
	for i := 0; i < lsh.bands; i++ {
		start := i * lsh.rows
		end := start + lsh.rows
		if end > len(signature) {
			end = len(signature)
		}

		// Get band signature
		bandSig := signature[start:end]
		bandKey := bandToString(bandSig)

		// Get documents in the same bucket
		for _, docID := range lsh.hashTables[i][bandKey] {
			candidates[docID] = struct{}{}
		}
	}

	// Compute actual similarities for candidates
	for docID := range candidates {
		// In a real implementation, we would store and retrieve the
		// document signatures rather than recomputing them

		// For this example, we'll just assume we can access signatures
		// otherSignature := getStoredSignature(docID)
		// similarity := lsh.minHash.Similarity(signature, otherSignature)

		// Placeholder for actual similarity computation
		similarity := 0.0

		if similarity >= threshold {
			similarities[docID] = similarity
		}
	}

	return similarities
}

// bandToString converts a band signature to a string representation
func bandToString(band []uint32) string {
	// Simple hash function for the band
	h := uint32(math.MaxUint32)
	for _, v := range band {
		h ^= v
		h *= uint32(math.MaxUint32) / 2
	}
	return fmt.Sprintf("%d", h)
}

// Example usage
// Document represents a text document
type Document struct {
	ID        int
	Path      string
	Shingles  []string
	Signature []uint32
}

// DocumentSet manages a collection of documents
type DocumentSet struct {
	docs    map[int]*Document
	minHash *MinHash
	lsh     *LSH
	nextID  int
}

// NewDocumentSet creates a new document set
func NewDocumentSet(hashFunctions, bands int) *DocumentSet {
	rows := hashFunctions / bands
	return &DocumentSet{
		docs:    make(map[int]*Document),
		minHash: NewMinHash(hashFunctions),
		lsh:     NewLSH(bands, rows),
		nextID:  0,
	}
}

// AddDocument adds a document to the set
func (ds *DocumentSet) AddDocument(path string) (*Document, error) {
	// Read file
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Convert to shingles
	shingles := DocumentToSet(file, 3) // 3-word shingles

	// Create document
	docID := ds.nextID
	ds.nextID++

	doc := &Document{
		ID:        docID,
		Path:      path,
		Shingles:  shingles,
		Signature: ds.minHash.Signature(shingles),
	}

	// Add to collection
	ds.docs[docID] = doc

	// Add to LSH index
	ds.lsh.AddDocument(docID, shingles)

	return doc, nil
}

// FindSimilar finds documents similar to the specified one
func (ds *DocumentSet) FindSimilar(docID int, threshold float64) []*Document {
	doc, exists := ds.docs[docID]
	if !exists {
		return nil
	}

	// Find candidate similar documents
	similarIDs := ds.lsh.FindSimilar(doc.Shingles, threshold)

	// Compute actual similarity for each candidate
	similar := make([]*Document, 0, len(similarIDs))
	for id, _ := range similarIDs { // _similarityEstimate
		if id == docID {
			continue // Skip the document itself
		}

		otherDoc := ds.docs[id]

		// Calculate actual similarity
		similarity := ds.minHash.Similarity(doc.Signature, otherDoc.Signature)

		if similarity >= threshold {
			similar = append(similar, otherDoc)
			// Update with actual similarity
			similarIDs[id] = similarity
		}
	}

	return similar
}

// FindDuplicates finds all groups of similar documents
func (ds *DocumentSet) FindDuplicates(threshold float64) [][]int {
	seen := make(map[int]bool)
	groups := [][]int{}

	for id := range ds.docs {
		if seen[id] {
			continue
		}

		// Find similar documents
		similarDocs := ds.FindSimilar(id, threshold)
		if len(similarDocs) == 0 {
			continue
		}

		// Create a group
		group := []int{id}
		seen[id] = true

		for _, doc := range similarDocs {
			group = append(group, doc.ID)
			seen[doc.ID] = true
		}

		groups = append(groups, group)
	}

	return groups
}

func main() {
	// Create a document set
	docSet := NewDocumentSet(100, 20) // 100 hash functions, 20 bands

	// Sample documents directory
	docsDir := "./sample_docs"

	// Add all text files
	err := filepath.Walk(docsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".txt") {
			fmt.Printf("Adding document: %s\n", path)
			_, err := docSet.AddDocument(path)
			if err != nil {
				fmt.Printf("Error adding document %s: %v\n", path, err)
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("Error walking directory: %v\n", err)
		return
	}

	// Find duplicate groups with similarity threshold of 0.8
	duplicateGroups := docSet.FindDuplicates(0.8)

	// Print results
	fmt.Printf("\nFound %d groups of similar documents:\n", len(duplicateGroups))
	for i, group := range duplicateGroups {
		fmt.Printf("\nGroup %d:\n", i+1)
		for _, docID := range group {
			doc := docSet.docs[docID]
			fmt.Printf("  - %s\n", doc.Path)
		}
	}
}
