package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Skpow1234/Peervault/internal/ml"
)

func main() {
	var (
		command = flag.String("command", "help", "Command to execute (classify, optimize, predict, train, help)")
		file    = flag.String("file", "", "File to process")
		model   = flag.String("model", "", "Model ID")
		help    = flag.Bool("help", false, "Show help")
	)
	flag.Parse()

	if *help || *command == "help" {
		showHelp()
		return
	}

	// Create ML classification engine
	mlEngine := ml.NewMLClassificationEngine()
	ctx := context.Background()

	switch *command {
	case "classify":
		handleClassifyCommand(ctx, mlEngine, *file)
	case "optimize":
		handleOptimizeCommand(ctx, mlEngine, *file)
	case "predict":
		handlePredictCommand(ctx, mlEngine, *file)
	case "train":
		handleTrainCommand(ctx, mlEngine, *model)
	default:
		log.Fatalf("Unknown command: %s", *command)
	}
}

func handleClassifyCommand(ctx context.Context, mlEngine *ml.MLClassificationEngine, filePath string) {
	if filePath == "" {
		log.Fatal("File path is required for classify command")
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Classify file
	classification, err := mlEngine.ClassifyFile(ctx, filePath, data, map[string]interface{}{
		"source":    "command_line",
		"timestamp": time.Now(),
	})
	if err != nil {
		log.Fatalf("Failed to classify file: %v", err)
	}

	fmt.Printf("File Classification Results:\n")
	fmt.Printf("File Path: %s\n", classification.FilePath)
	fmt.Printf("Category: %s\n", classification.Category)
	fmt.Printf("Confidence: %.2f%%\n", classification.Confidence*100)
	fmt.Printf("Size: %d bytes\n", classification.Size)
	fmt.Printf("Extension: %s\n", classification.Extension)
	fmt.Printf("MIME Type: %s\n", classification.MimeType)
	fmt.Printf("Tags: %v\n", classification.Tags)
	fmt.Printf("Created At: %s\n", classification.CreatedAt.Format(time.RFC3339))

	// List all classifications
	classifications, err := mlEngine.ListClassifications(ctx)
	if err != nil {
		log.Fatalf("Failed to list classifications: %v", err)
	}

	fmt.Printf("\nAll Classifications:\n")
	for _, classification := range classifications {
		fmt.Printf("  %s: %s (%.2f%%)\n", classification.FilePath, classification.Category, classification.Confidence*100)
	}
}

func handleOptimizeCommand(ctx context.Context, mlEngine *ml.MLClassificationEngine, filePath string) {
	if filePath == "" {
		log.Fatal("File path is required for optimize command")
	}

	// Read file
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	// Test different optimization types
	optimizationTypes := []string{"compression", "deduplication", "encoding"}

	for _, optType := range optimizationTypes {
		fmt.Printf("\nTesting %s optimization:\n", optType)

		optimization, err := mlEngine.OptimizeFile(ctx, filePath, data, optType)
		if err != nil {
			log.Printf("Failed to optimize file with %s: %v", optType, err)
			continue
		}

		fmt.Printf("  Original Size: %d bytes\n", optimization.OriginalSize)
		fmt.Printf("  Optimized Size: %d bytes\n", optimization.OptimizedSize)
		fmt.Printf("  Compression Ratio: %.2f%%\n", optimization.CompressionRatio*100)
		fmt.Printf("  Algorithm: %s\n", optimization.Algorithm)
		fmt.Printf("  Processing Time: %v\n", optimization.ProcessingTime)
	}

	// List all optimizations
	optimizations, err := mlEngine.ListOptimizations(ctx)
	if err != nil {
		log.Fatalf("Failed to list optimizations: %v", err)
	}

	fmt.Printf("\nAll Optimizations:\n")
	for _, optimization := range optimizations {
		fmt.Printf("  %s: %s (%.2f%% reduction)\n",
			optimization.Metadata["file_path"],
			optimization.OptimizationType,
			(1-optimization.CompressionRatio)*100)
	}
}

func handlePredictCommand(ctx context.Context, mlEngine *ml.MLClassificationEngine, filePath string) {
	if filePath == "" {
		log.Fatal("File path is required for predict command")
	}

	// Simulate access history for the file
	accessHistory := []time.Time{
		time.Now().Add(-24 * time.Hour),
		time.Now().Add(-12 * time.Hour),
		time.Now().Add(-6 * time.Hour),
		time.Now().Add(-1 * time.Hour),
		time.Now().Add(-30 * time.Minute),
	}

	// Predict cache access
	prediction, err := mlEngine.PredictCacheAccess(ctx, filePath, accessHistory, map[string]interface{}{
		"file_type":     "document",
		"size":          1024,
		"last_modified": time.Now().Add(-2 * time.Hour),
	})
	if err != nil {
		log.Fatalf("Failed to predict cache access: %v", err)
	}

	fmt.Printf("Cache Access Prediction Results:\n")
	fmt.Printf("Key: %s\n", prediction.Key)
	fmt.Printf("Access Probability: %.2f%%\n", prediction.AccessProbability*100)
	fmt.Printf("Recommended TTL: %v\n", prediction.RecommendedTTL)
	fmt.Printf("Priority: %d\n", prediction.Priority)
	fmt.Printf("Reason: %s\n", prediction.Reason)

	// List all predictions
	predictions, err := mlEngine.ListCachePredictions(ctx)
	if err != nil {
		log.Fatalf("Failed to list predictions: %v", err)
	}

	fmt.Printf("\nAll Cache Predictions:\n")
	for _, prediction := range predictions {
		fmt.Printf("  %s: %.2f%% probability, TTL: %v, Priority: %d\n",
			prediction.Key,
			prediction.AccessProbability*100,
			prediction.RecommendedTTL,
			prediction.Priority)
	}
}

func handleTrainCommand(ctx context.Context, mlEngine *ml.MLClassificationEngine, modelID string) {
	if modelID == "" {
		modelID = "default_model"
	}

	// Create a sample model
	model := &ml.MLModel{
		ID:      modelID,
		Name:    "File Classification Model",
		Type:    "classification",
		Version: "1.0.0",
		Parameters: map[string]interface{}{
			"algorithm":    "random_forest",
			"max_depth":    10,
			"n_estimators": 100,
		},
	}

	// Create sample training data
	trainingData := []map[string]interface{}{
		{"extension": ".txt", "size": 1024, "label": "document"},
		{"extension": ".jpg", "size": 2048, "label": "image"},
		{"extension": ".mp4", "size": 1024000, "label": "video"},
		{"extension": ".mp3", "size": 512000, "label": "audio"},
		{"extension": ".pdf", "size": 5120, "label": "document"},
		{"extension": ".zip", "size": 10240, "label": "archive"},
		{"extension": ".go", "size": 2048, "label": "code"},
		{"extension": ".json", "size": 512, "label": "data"},
	}

	// Train model
	err := mlEngine.TrainModel(ctx, model, trainingData)
	if err != nil {
		log.Fatalf("Failed to train model: %v", err)
	}

	fmt.Printf("Model Trained Successfully!\n")
	fmt.Printf("Model ID: %s\n", model.ID)
	fmt.Printf("Name: %s\n", model.Name)
	fmt.Printf("Type: %s\n", model.Type)
	fmt.Printf("Version: %s\n", model.Version)
	fmt.Printf("Accuracy: %.2f%%\n", model.Accuracy*100)
	fmt.Printf("Training Samples: %d\n", model.TrainingData["sample_count"])
	fmt.Printf("Features: %v\n", model.TrainingData["features"])
	fmt.Printf("Labels: %v\n", model.TrainingData["labels"])
	fmt.Printf("Created At: %s\n", model.CreatedAt.Format(time.RFC3339))

	// List all models
	models, err := mlEngine.ListModels(ctx)
	if err != nil {
		log.Fatalf("Failed to list models: %v", err)
	}

	fmt.Printf("\nAll Models:\n")
	for _, model := range models {
		fmt.Printf("  %s: %s (%.2f%% accuracy)\n", model.ID, model.Name, model.Accuracy*100)
	}
}

func showHelp() {
	fmt.Printf("PeerVault Machine Learning Tool\n\n")
	fmt.Printf("Usage: peervault-ml -command <command> [options]\n\n")
	fmt.Printf("Commands:\n")
	fmt.Printf("  classify   Classify a file using ML algorithms\n")
	fmt.Printf("  optimize   Optimize a file using ML-based algorithms\n")
	fmt.Printf("  predict    Predict cache access patterns for a file\n")
	fmt.Printf("  train      Train a machine learning model\n")
	fmt.Printf("  help       Show this help message\n\n")
	fmt.Printf("Options:\n")
	fmt.Printf("  -file <path>      File path (for classify, optimize, predict commands)\n")
	fmt.Printf("  -model <id>       Model ID (for train command)\n")
	fmt.Printf("  -help             Show this help message\n\n")
	fmt.Printf("Examples:\n")
	fmt.Printf("  peervault-ml -command classify -file example.txt\n")
	fmt.Printf("  peervault-ml -command optimize -file example.jpg\n")
	fmt.Printf("  peervault-ml -command predict -file example.pdf\n")
	fmt.Printf("  peervault-ml -command train -model my_model\n")
}
