EXECUTABLE := hitbloq_curve_changer
BUILD_DIR := build
WINDOWS := $(BUILD_DIR)/$(EXECUTABLE)_windows_amd64.exe
LINUX := $(BUILD_DIR)/$(EXECUTABLE)_linux_amd64
DARWIN := $(BUILD_DIR)/$(EXECUTABLE)_darwin_amd64
VERSION := $(shell git describe --tags --always --long --dirty)

.PHONY: build clean help

build: windows linux darwin 
	@echo "Version: $(VERSION)"
	@echo "Build complete."

windows: $(WINDOWS) 
$(WINDOWS):
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 go build -o $(WINDOWS)

linux: $(LINUX) 
$(LINUX):
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 go build -o $(LINUX)

darwin: $(DARWIN) 
$(DARWIN):
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 go build -o $(DARWIN)

clean:
	rm -rf $(BUILD_DIR)
	@echo "Clean complete."
