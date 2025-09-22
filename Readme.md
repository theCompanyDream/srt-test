# Caption Validator - Command Line Interface

A command-line tool for validating WebVTT (.vtt) and SRT (.srt) caption files against time coverage and language requirements.

## Installation

### Building from Source
```bash
# Clone and build
git clone https://github.com/theCompanyDream/srt-test.git
cd caption-validator
go build -o caption-validator .
```

## Command Syntax

```bash
caption-validator [OPTIONS]
```

## Required Arguments

| Flag | Description | Example |
|------|-------------|---------|
| `--file` | Path to caption file (.vtt or .srt) | `--file=subtitles.vtt` |
| `--t_end` | End time for validation range | `--t_end=5m30s` |
| `--endpoint` | Language detection API endpoint | `--endpoint=https://api.example.com/detect` |

## Optional Arguments

| Flag | Default | Description | Example |
|------|---------|-------------|---------|
| `--t_start` | `0s` | Start time for validation range | `--t_start=1m` |
| `--coverage` | `0.8` | Required coverage percentage (0.0-1.0) | `--coverage=0.9` |

## Time Format Examples

The `--t_start` and `--t_end` flags accept Go duration format:

| Format | Example | Description |
|--------|---------|-------------|
| `30s` | 30 seconds | Simple seconds |
| `5m` | 5 minutes | Simple minutes |
| `1h` | 1 hour | Simple hours |
| `1m30s` | 1 minute 30 seconds | Combined units |
| `2h45m` | 2 hours 45 minutes | Hours and minutes |
| `1h30m45s` | 1 hour 30 minutes 45 seconds | All units |

## Usage Examples

### Basic Usage
```bash
# Validate a WebVTT file for 5 minutes with 80% coverage requirement
caption-validator \
  --file=movie.vtt \
  --t_end=5m \
  --endpoint=http://localhost:8080/detect
```

### Custom Time Range
```bash
# Validate from 1:30 to 10:00 with 90% coverage requirement
caption-validator \
  --file=episode.srt \
  --t_start=1m30s \
  --t_end=10m \
  --coverage=0.9 \
  --endpoint=https://api.langdetect.com/analyze
```

### Movie Validation
```bash
# Validate full movie (2 hours) with default 80% coverage
caption-validator \
  --file=movie.srt \
  --t_end=2h \
  --endpoint=https://lang-api.company.com/detect
```

### Clip Validation
```bash
# Validate a 30-second clip starting at 2 minutes
caption-validator \
  --file=clip.vtt \
  --t_start=2m \
  --t_end=2m30s \
  --coverage=1.0 \
  --endpoint=http://192.168.1.100:3000/language
```