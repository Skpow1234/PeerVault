package ml

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
)

// FileClassification represents a file classification result
type FileClassification struct {
	FilePath   string                 `json:"file_path"`
	Category   string                 `json:"category"`
	Confidence float64                `json:"confidence"`
	Tags       []string               `json:"tags"`
	Size       int64                  `json:"size"`
	Extension  string                 `json:"extension"`
	MimeType   string                 `json:"mime_type"`
	CreatedAt  time.Time              `json:"created_at"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// OptimizationResult represents an optimization result
type OptimizationResult struct {
	OriginalSize     int64                  `json:"original_size"`
	OptimizedSize    int64                  `json:"optimized_size"`
	CompressionRatio float64                `json:"compression_ratio"`
	OptimizationType string                 `json:"optimization_type"`
	Algorithm        string                 `json:"algorithm"`
	ProcessingTime   time.Duration          `json:"processing_time"`
	Metadata         map[string]interface{} `json:"metadata"`
}

// CachePrediction represents a cache prediction result
type CachePrediction struct {
	Key               string                 `json:"key"`
	AccessProbability float64                `json:"access_probability"`
	RecommendedTTL    time.Duration          `json:"recommended_ttl"`
	Priority          int                    `json:"priority"`
	Reason            string                 `json:"reason"`
	Metadata          map[string]interface{} `json:"metadata"`
}

// MLModel represents a machine learning model
type MLModel struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	Version      string                 `json:"version"`
	Accuracy     float64                `json:"accuracy"`
	TrainingData map[string]interface{} `json:"training_data"`
	Parameters   map[string]interface{} `json:"parameters"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// MLClassificationEngine provides machine learning classification functionality
type MLClassificationEngine struct {
	mu               sync.RWMutex
	models           map[string]*MLModel
	classifications  map[string]*FileClassification
	optimizations    map[string]*OptimizationResult
	cachePredictions map[string]*CachePrediction
}

// NewMLClassificationEngine creates a new ML classification engine
func NewMLClassificationEngine() *MLClassificationEngine {
	return &MLClassificationEngine{
		models:           make(map[string]*MLModel),
		classifications:  make(map[string]*FileClassification),
		optimizations:    make(map[string]*OptimizationResult),
		cachePredictions: make(map[string]*CachePrediction),
	}
}

// ClassifyFile classifies a file based on its content and metadata
func (mce *MLClassificationEngine) ClassifyFile(ctx context.Context, filePath string, content []byte, metadata map[string]interface{}) (*FileClassification, error) {
	// Extract file information
	extension := getFileExtension(filePath)
	mimeType := getMimeType(extension)

	// Simple classification based on file extension and content
	category, confidence, tags := mce.classifyByContent(extension, content, metadata)

	classification := &FileClassification{
		FilePath:   filePath,
		Category:   category,
		Confidence: confidence,
		Tags:       tags,
		Size:       int64(len(content)),
		Extension:  extension,
		MimeType:   mimeType,
		CreatedAt:  time.Now(),
		Metadata:   metadata,
	}

	mce.mu.Lock()
	mce.classifications[filePath] = classification
	mce.mu.Unlock()

	return classification, nil
}

// OptimizeFile optimizes a file using ML-based algorithms
func (mce *MLClassificationEngine) OptimizeFile(ctx context.Context, filePath string, content []byte, optimizationType string) (*OptimizationResult, error) {
	startTime := time.Now()

	var optimizedSize int64
	var algorithm string

	switch optimizationType {
	case "compression":
		optimizedSize, algorithm = mce.optimizeCompression(content)
	case "deduplication":
		optimizedSize, algorithm = mce.optimizeDeduplication(content)
	case "encoding":
		optimizedSize, algorithm = mce.optimizeEncoding(content)
	default:
		optimizedSize = int64(len(content))
		algorithm = "none"
	}

	processingTime := time.Since(startTime)
	compressionRatio := float64(optimizedSize) / float64(len(content))

	result := &OptimizationResult{
		OriginalSize:     int64(len(content)),
		OptimizedSize:    optimizedSize,
		CompressionRatio: compressionRatio,
		OptimizationType: optimizationType,
		Algorithm:        algorithm,
		ProcessingTime:   processingTime,
		Metadata: map[string]interface{}{
			"file_path": filePath,
			"timestamp": time.Now(),
		},
	}

	mce.mu.Lock()
	mce.optimizations[filePath] = result
	mce.mu.Unlock()

	return result, nil
}

// PredictCacheAccess predicts cache access patterns
func (mce *MLClassificationEngine) PredictCacheAccess(ctx context.Context, key string, accessHistory []time.Time, metadata map[string]interface{}) (*CachePrediction, error) {
	// Calculate access probability based on historical patterns
	accessProbability := mce.calculateAccessProbability(accessHistory)

	// Determine recommended TTL based on access patterns
	recommendedTTL := mce.calculateRecommendedTTL(accessHistory)

	// Calculate priority based on access probability and recency
	priority := mce.calculatePriority(accessProbability, accessHistory)

	// Generate reason for the prediction
	reason := mce.generatePredictionReason(accessProbability, recommendedTTL, priority)

	prediction := &CachePrediction{
		Key:               key,
		AccessProbability: accessProbability,
		RecommendedTTL:    recommendedTTL,
		Priority:          priority,
		Reason:            reason,
		Metadata:          metadata,
	}

	mce.mu.Lock()
	mce.cachePredictions[key] = prediction
	mce.mu.Unlock()

	return prediction, nil
}

// TrainModel trains a machine learning model
func (mce *MLClassificationEngine) TrainModel(ctx context.Context, model *MLModel, trainingData []map[string]interface{}) error {
	// Simulate model training
	// In a real implementation, this would use actual ML libraries

	model.TrainingData = map[string]interface{}{
		"sample_count": len(trainingData),
		"features":     mce.extractFeatures(trainingData),
		"labels":       mce.extractLabels(trainingData),
	}

	model.Accuracy = mce.calculateAccuracy(trainingData)
	model.CreatedAt = time.Now()
	model.UpdatedAt = time.Now()

	mce.mu.Lock()
	mce.models[model.ID] = model
	mce.mu.Unlock()

	return nil
}

// GetModel retrieves a model by ID
func (mce *MLClassificationEngine) GetModel(ctx context.Context, modelID string) (*MLModel, error) {
	mce.mu.RLock()
	model, exists := mce.models[modelID]
	mce.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("model not found: %s", modelID)
	}

	return model, nil
}

// ListModels lists all models
func (mce *MLClassificationEngine) ListModels(ctx context.Context) ([]*MLModel, error) {
	mce.mu.RLock()
	models := make([]*MLModel, 0, len(mce.models))
	for _, model := range mce.models {
		models = append(models, model)
	}
	mce.mu.RUnlock()
	return models, nil
}

// GetClassification retrieves a classification by file path
func (mce *MLClassificationEngine) GetClassification(ctx context.Context, filePath string) (*FileClassification, error) {
	mce.mu.RLock()
	classification, exists := mce.classifications[filePath]
	mce.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("classification not found: %s", filePath)
	}

	return classification, nil
}

// ListClassifications lists all classifications
func (mce *MLClassificationEngine) ListClassifications(ctx context.Context) ([]*FileClassification, error) {
	mce.mu.RLock()
	classifications := make([]*FileClassification, 0, len(mce.classifications))
	for _, classification := range mce.classifications {
		classifications = append(classifications, classification)
	}
	mce.mu.RUnlock()
	return classifications, nil
}

// GetOptimization retrieves an optimization result by file path
func (mce *MLClassificationEngine) GetOptimization(ctx context.Context, filePath string) (*OptimizationResult, error) {
	mce.mu.RLock()
	optimization, exists := mce.optimizations[filePath]
	mce.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("optimization not found: %s", filePath)
	}

	return optimization, nil
}

// ListOptimizations lists all optimization results
func (mce *MLClassificationEngine) ListOptimizations(ctx context.Context) ([]*OptimizationResult, error) {
	mce.mu.RLock()
	optimizations := make([]*OptimizationResult, 0, len(mce.optimizations))
	for _, optimization := range mce.optimizations {
		optimizations = append(optimizations, optimization)
	}
	mce.mu.RUnlock()
	return optimizations, nil
}

// GetCachePrediction retrieves a cache prediction by key
func (mce *MLClassificationEngine) GetCachePrediction(ctx context.Context, key string) (*CachePrediction, error) {
	mce.mu.RLock()
	prediction, exists := mce.cachePredictions[key]
	mce.mu.RUnlock()
	if !exists {
		return nil, fmt.Errorf("cache prediction not found: %s", key)
	}

	return prediction, nil
}

// ListCachePredictions lists all cache predictions
func (mce *MLClassificationEngine) ListCachePredictions(ctx context.Context) ([]*CachePrediction, error) {
	mce.mu.RLock()
	predictions := make([]*CachePrediction, 0, len(mce.cachePredictions))
	for _, prediction := range mce.cachePredictions {
		predictions = append(predictions, prediction)
	}
	mce.mu.RUnlock()
	return predictions, nil
}

// classifyByContent classifies content based on extension and content
func (mce *MLClassificationEngine) classifyByContent(extension string, content []byte, metadata map[string]interface{}) (string, float64, []string) {
	var category string
	var confidence float64
	var tags []string

	// Simple classification logic based on file extension
	switch strings.ToLower(extension) {
	case ".txt", ".md", ".rst":
		category = "document"
		confidence = 0.95
		tags = []string{"text", "document", "readable"}
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp":
		category = "image"
		confidence = 0.98
		tags = []string{"image", "visual", "media"}
	case ".mp4", ".avi", ".mov", ".mkv":
		category = "video"
		confidence = 0.97
		tags = []string{"video", "media", "streaming"}
	case ".mp3", ".wav", ".flac", ".aac":
		category = "audio"
		confidence = 0.96
		tags = []string{"audio", "music", "sound"}
	case ".pdf":
		category = "document"
		confidence = 0.99
		tags = []string{"document", "pdf", "readable"}
	case ".zip", ".rar", ".7z", ".tar", ".gz":
		category = "archive"
		confidence = 0.98
		tags = []string{"archive", "compressed", "package"}
	case ".exe", ".dll", ".so", ".dylib":
		category = "executable"
		confidence = 0.99
		tags = []string{"executable", "binary", "program"}
	case ".json", ".xml", ".yaml", ".yml":
		category = "data"
		confidence = 0.94
		tags = []string{"data", "structured", "config"}
	case ".go", ".js", ".py", ".java", ".cpp", ".c":
		category = "code"
		confidence = 0.93
		tags = []string{"code", "programming", "source"}
	default:
		category = "unknown"
		confidence = 0.5
		tags = []string{"unknown", "unclassified"}
	}

	// Adjust confidence based on content analysis
	if len(content) > 0 {
		confidence = math.Min(confidence+0.05, 1.0)
	}

	return category, confidence, tags
}

// optimizeCompression optimizes content using compression
func (mce *MLClassificationEngine) optimizeCompression(content []byte) (int64, string) {
	// Simulate compression optimization
	// In a real implementation, this would use actual compression algorithms
	compressedSize := int64(float64(len(content)) * 0.7) // 30% compression
	return compressedSize, "gzip"
}

// optimizeDeduplication optimizes content using deduplication
func (mce *MLClassificationEngine) optimizeDeduplication(content []byte) (int64, string) {
	// Simulate deduplication optimization
	// In a real implementation, this would use actual deduplication algorithms
	deduplicatedSize := int64(float64(len(content)) * 0.8) // 20% deduplication
	return deduplicatedSize, "content-hash"
}

// optimizeEncoding optimizes content using encoding
func (mce *MLClassificationEngine) optimizeEncoding(content []byte) (int64, string) {
	// Simulate encoding optimization
	// In a real implementation, this would use actual encoding optimization
	optimizedSize := int64(float64(len(content)) * 0.9) // 10% optimization
	return optimizedSize, "base64"
}

// calculateAccessProbability calculates access probability based on history
func (mce *MLClassificationEngine) calculateAccessProbability(accessHistory []time.Time) float64 {
	if len(accessHistory) == 0 {
		return 0.1 // Default low probability
	}

	// Calculate based on recency and frequency
	now := time.Now()
	recentAccesses := 0

	for _, access := range accessHistory {
		if now.Sub(access) < 24*time.Hour {
			recentAccesses++
		}
	}

	// Simple probability calculation
	probability := float64(recentAccesses) / float64(len(accessHistory))
	return math.Min(probability, 1.0)
}

// calculateRecommendedTTL calculates recommended TTL based on access patterns
func (mce *MLClassificationEngine) calculateRecommendedTTL(accessHistory []time.Time) time.Duration {
	if len(accessHistory) == 0 {
		return 1 * time.Hour // Default TTL
	}

	// Calculate average time between accesses
	if len(accessHistory) < 2 {
		return 1 * time.Hour
	}

	sort.Slice(accessHistory, func(i, j int) bool {
		return accessHistory[i].Before(accessHistory[j])
	})

	totalDuration := accessHistory[len(accessHistory)-1].Sub(accessHistory[0])
	avgDuration := totalDuration / time.Duration(len(accessHistory)-1)

	// Set TTL to 2x the average duration, with min/max bounds
	ttl := avgDuration * 2
	if ttl < 5*time.Minute {
		ttl = 5 * time.Minute
	}
	if ttl > 24*time.Hour {
		ttl = 24 * time.Hour
	}

	return ttl
}

// calculatePriority calculates priority based on access probability and recency
func (mce *MLClassificationEngine) calculatePriority(accessProbability float64, accessHistory []time.Time) int {
	priority := int(accessProbability * 100)

	// Boost priority for recent accesses
	if len(accessHistory) > 0 {
		lastAccess := accessHistory[len(accessHistory)-1]
		if time.Since(lastAccess) < 1*time.Hour {
			priority += 20
		}
	}

	// Ensure priority is within bounds
	if priority > 100 {
		priority = 100
	}
	if priority < 1 {
		priority = 1
	}

	return priority
}

// generatePredictionReason generates a reason for the prediction
func (mce *MLClassificationEngine) generatePredictionReason(accessProbability float64, ttl time.Duration, priority int) string {
	if accessProbability > 0.8 {
		return fmt.Sprintf("High access probability (%.2f), recommended TTL: %v, priority: %d", accessProbability, ttl, priority)
	} else if accessProbability > 0.5 {
		return fmt.Sprintf("Medium access probability (%.2f), recommended TTL: %v, priority: %d", accessProbability, ttl, priority)
	} else {
		return fmt.Sprintf("Low access probability (%.2f), recommended TTL: %v, priority: %d", accessProbability, ttl, priority)
	}
}

// extractFeatures extracts features from training data
func (mce *MLClassificationEngine) extractFeatures(trainingData []map[string]interface{}) []string {
	features := make([]string, 0)

	for _, data := range trainingData {
		for key := range data {
			if !contains(features, key) {
				features = append(features, key)
			}
		}
	}

	return features
}

// extractLabels extracts labels from training data
func (mce *MLClassificationEngine) extractLabels(trainingData []map[string]interface{}) []string {
	labels := make([]string, 0)

	for _, data := range trainingData {
		if label, exists := data["label"]; exists {
			if labelStr, ok := label.(string); ok && !contains(labels, labelStr) {
				labels = append(labels, labelStr)
			}
		}
	}

	return labels
}

// calculateAccuracy calculates model accuracy
func (mce *MLClassificationEngine) calculateAccuracy(trainingData []map[string]interface{}) float64 {
	// Simulate accuracy calculation
	// In a real implementation, this would use actual accuracy metrics
	return 0.85 + (float64(len(trainingData)%100) / 1000.0) // Simulate 85-95% accuracy
}

// getFileExtension extracts file extension from path
func getFileExtension(filePath string) string {
	parts := strings.Split(filePath, ".")
	if len(parts) > 1 {
		return "." + parts[len(parts)-1]
	}
	return ""
}

// getMimeType returns MIME type based on extension
func getMimeType(extension string) string {
	mimeTypes := map[string]string{
		".txt":  "text/plain",
		".md":   "text/markdown",
		".json": "application/json",
		".xml":  "application/xml",
		".pdf":  "application/pdf",
		".jpg":  "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
		".zip":  "application/zip",
		".go":   "text/x-go",
		".js":   "application/javascript",
		".py":   "text/x-python",
	}

	if mimeType, exists := mimeTypes[strings.ToLower(extension)]; exists {
		return mimeType
	}

	return "application/octet-stream"
}

// contains checks if a string slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
