# Test Data Directory

This directory contains test fixtures for integration tests.

## Images

Sample images are auto-generated on first test run:

- `sample.jpg` - 800x600 JPEG with gradient pattern
- `sample.png` - 640x480 PNG with gradient pattern
- `large.jpg` - 3000x2000 JPEG for testing resize functionality

These images are generated programmatically by `internal/testutil/images.go` to ensure consistent test data without bloating the repository.

## Usage

Test fixtures are automatically created when running integration tests. You don't need to manually create these files.

To regenerate fixtures, delete the `images/` directory and run:

```bash
go test ./... -v
```
