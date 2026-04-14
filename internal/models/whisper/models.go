package whisper

import (
	"os"
	"path/filepath"
)

// ModelInfo holds metadata for a whisper model
type ModelInfo struct {
	ID           string // model identifier (e.g., "base.en")
	Name         string // display name (e.g., "Base English")
	Filename     string // file name (e.g., "ggml-base.en.bin")
	Size         string // human readable size
	SizeBytes    int64  // size in bytes for progress tracking
	SHA1         string // official Hugging Face "Disk SHA" for integrity verification
	Multilingual bool   // true if supports multiple languages
}

// available whisper models from huggingface.co/ggerganov/whisper.cpp
var models = []ModelInfo{
	// english-only models (faster, smaller)
	{ID: "tiny.en", Name: "Tiny English", Filename: "ggml-tiny.en.bin", Size: "75MB", SizeBytes: 75_000_000, SHA1: "c78c86eb1a8faa21b369bcd33207cc90d64ae9df", Multilingual: false},
	{ID: "base.en", Name: "Base English", Filename: "ggml-base.en.bin", Size: "142MB", SizeBytes: 142_000_000, SHA1: "137c40403d78fd54d454da0f9bd998f78703390c", Multilingual: false},
	{ID: "small.en", Name: "Small English", Filename: "ggml-small.en.bin", Size: "466MB", SizeBytes: 466_000_000, SHA1: "db8a495a91d927739e50b3fc1cc4c6b8f6c2d022", Multilingual: false},
	{ID: "medium.en", Name: "Medium English", Filename: "ggml-medium.en.bin", Size: "1.5GB", SizeBytes: 1_500_000_000, SHA1: "8c30f0e44ce9560643ebd10bbe50cd20eafd3723", Multilingual: false},

	// multilingual models
	{ID: "tiny", Name: "Tiny", Filename: "ggml-tiny.bin", Size: "75MB", SizeBytes: 75_000_000, SHA1: "bd577a113a864445d4c299885e0cb97d4ba92b5f", Multilingual: true},
	{ID: "base", Name: "Base", Filename: "ggml-base.bin", Size: "142MB", SizeBytes: 142_000_000, SHA1: "465707469ff3a37a2b9b8d8f89f2f99de7299dac", Multilingual: true},
	{ID: "small", Name: "Small", Filename: "ggml-small.bin", Size: "466MB", SizeBytes: 466_000_000, SHA1: "55356645c2b361a969dfd0ef2c5a50d530afd8d5", Multilingual: true},
	{ID: "medium", Name: "Medium", Filename: "ggml-medium.bin", Size: "1.5GB", SizeBytes: 1_500_000_000, SHA1: "fd9727b6e1217c2f614f9b698455c4ffd82463b4", Multilingual: true},
	{ID: "large-v1", Name: "Large V1", Filename: "ggml-large-v1.bin", Size: "2.9GB", SizeBytes: 2_900_000_000, SHA1: "b1caaf735c4cc1429223d5a74f0f4d0b9b59a299", Multilingual: true},
	{ID: "large-v2", Name: "Large V2", Filename: "ggml-large-v2.bin", Size: "2.9GB", SizeBytes: 2_900_000_000, SHA1: "0f4c8e34f21cf1a914c59d8b3ce882345ad349d6", Multilingual: true},
	{ID: "large-v3", Name: "Large V3", Filename: "ggml-large-v3.bin", Size: "3GB", SizeBytes: 3_000_000_000, SHA1: "ad82bf6a9043ceed055076d0fd39f5f186ff8062", Multilingual: true},
	{ID: "large-v3-turbo", Name: "Large V3 Turbo", Filename: "ggml-large-v3-turbo.bin", Size: "1.6GB", SizeBytes: 1_600_000_000, SHA1: "4af2b29d7ec73d781377bfd1758ca957a807e941", Multilingual: true},
}

// modelByID maps model ID to ModelInfo for quick lookup
var modelByID = func() map[string]ModelInfo {
	m := make(map[string]ModelInfo, len(models))
	for _, model := range models {
		m[model.ID] = model
	}
	return m
}()

const (
	// base URL for downloading models from huggingface
	baseDownloadURL = "https://huggingface.co/ggerganov/whisper.cpp/resolve/main"
)

// GetModelsDir returns the directory where whisper models are stored.
// Creates the directory if it doesn't exist.
func GetModelsDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".local", "share", "hyprvoice", "models", "whisper")
	return dir, nil
}

// GetModelPath returns the full path to a model file.
// Returns empty string if model ID is unknown.
func GetModelPath(modelID string) string {
	info, ok := modelByID[modelID]
	if !ok {
		return ""
	}
	dir, err := GetModelsDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, info.Filename)
}

// GetDownloadURL returns the full download URL for a model.
// Returns empty string if model ID is unknown.
func GetDownloadURL(modelID string) string {
	info, ok := modelByID[modelID]
	if !ok {
		return ""
	}
	return baseDownloadURL + "/" + info.Filename
}

// GetModel returns info for a model by ID.
// Returns nil if model ID is unknown.
func GetModel(modelID string) *ModelInfo {
	info, ok := modelByID[modelID]
	if !ok {
		return nil
	}
	return &info
}

// ListModels returns all available whisper models
func ListModels() []ModelInfo {
	result := make([]ModelInfo, len(models))
	copy(result, models)
	return result
}

// ListMultilingualModels returns models that support multiple languages
func ListMultilingualModels() []ModelInfo {
	var result []ModelInfo
	for _, m := range models {
		if m.Multilingual {
			result = append(result, m)
		}
	}
	return result
}

// ListEnglishOnlyModels returns english-only models
func ListEnglishOnlyModels() []ModelInfo {
	var result []ModelInfo
	for _, m := range models {
		if !m.Multilingual {
			result = append(result, m)
		}
	}
	return result
}
