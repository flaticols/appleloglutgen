# LUT Generator for Apple Log

This utility generates 3D LUTs (Lookup Tables) that convert from Apple Log to Rec.709 with optional creative looks.

## ⚠️ WARNING ⚠️

**THIS IS A PLAYGROUND/EXPERIMENTAL TOOL**

The results produced by this tool may be very unpredictable and are not intended for professional use. The generated LUTs:
- Use simplified approximations of actual color transforms
- Do not accurately replicate Apple's official color science
- May produce unexpected or inconsistent results
- Are NOT suitable for critical color work

This tool is intended for experimentation and educational purposes only.

## Features

- Converts Apple Log to Rec.709 color space
- Customizable LUT size (default 17x17x17)
- Optional creative looks:
  - Teal & Orange
  - Warm Vintage
- Exposure adjustment parameter
- Batch processing via JSON configuration files

## Usage

### Using Make and Docker (Recommended)

The simplest way to run the tool is using the provided Makefile with Docker:

1. Create a `configs` directory with your JSON configuration files
2. Create an `output` directory where the generated LUTs will be saved
3. Run the tool with Make:

```bash
# Build and run in one command
make

# To build the Docker image only
make build

# To run the container only (after building)
make run

# To clean up the Docker image
make clean
```

The Makefile automatically mounts your local config and output directories to the Docker container.

### Manual Go Execution

Alternatively, run the Go code directly:

```bash
go run loglutgen/main.go --configDir=configs --outputDir=output
```

Or build and run:

```bash
go build -o loglutgen loglutgen/main.go
./loglutgen --configDir=configs --outputDir=output
```

## Configuration Parameters

Create JSON files in your config directory with these parameters:

```json
{
  "size": 17,
  "red_tint": 1.05,
  "blue_tint": 0.95,
  "output": "my_custom_lut.cube",
  "look": "tealOrange",
  "exposure_offset": 1.0
}
```

| Parameter | Description | Default |
|-----------|-------------|---------|
| `size` | Grid dimension of the LUT | 17 |
| `red_tint` | Additional red multiplier | 1.05 |
| `blue_tint` | Additional blue multiplier | 0.95 |
| `output` | Output file name | "output.cube" |
| `look` | Creative look ("none", "tealOrange", or "warmVintage") | "none" |
| `exposure_offset` | Factor to adjust exposure | 1.0 |

## Example Configurations

### Default Log Conversion

```json
{
  "output": "apple_log_default.cube",
  "look": "none"
}
```

### Teal & Orange Look

```json
{
  "output": "apple_log_teal_orange.cube",
  "look": "tealOrange",
  "exposure_offset": 1.1
}
```

### Warm Vintage Look

```json
{
  "output": "apple_log_warm_vintage.cube",
  "look": "warmVintage",
  "red_tint": 1.1,
  "blue_tint": 0.9
}
```

## Using the Generated LUTs

The generated `.cube` files can be imported into video editing software that supports 3D LUTs, such as:
- DaVinci Resolve
- Final Cut Pro
- Adobe Premiere Pro
- Adobe After Effects

Remember that these are experimental approximations and may require additional adjustment in your editing software.

## Docker Setup

The Makefile configures Docker to build a multi-architecture image (amd64 and arm64) that can run on both Intel/AMD and Apple Silicon machines. It mounts your local directories:

- `./configs` → `/app/configs` in the container
- `./output` → `/app/output` in the container

This allows you to prepare configurations locally and access the outputs without copying files to/from the container.
