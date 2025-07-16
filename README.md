# ZettelFlow

ZettelFlow is a powerful, CLI-driven pipeline for transforming raw, unstructured text into a collection of structured, inter-connected notes, ready for any Zettelkasten or personal knowledge management (PKM) system.

It uses a three-stage, LLM-powered workflow to distill your ideas into a clean, organized format.

## Quick Start

To get started, you need a Go environment (version 1.23 or later) and an OpenAI-compatible API key.

### 1. Build the Application

From the project root, run the `make` command to build the `zettelflow` binary:

```sh
make build
```

### 2. Set Your API Key

The application needs an OpenAI API key to function. The first time you run any command that requires the LLM, it will prompt you to enter your key, which will be saved to the configuration file.

Alternatively, you can set it as an environment variable:

```sh
export OPENAI_API_KEY="your-api-key-here"
```

### 3. Run the Pipeline

The core workflow consists of three commands run in sequence.

**Stage 1: Ingest**
Take a local file (`my-raw-notes.txt`) and process it with the initial prompt.

```sh
./bin/zettelflow ingest my-raw-notes.txt
```
This saves a new file in your `ingest` data directory.

**Stage 2: Split**
Take the latest ingested file, split it into chunks based on the configured delimiter (`###`), and create structured note stubs.

```sh
./bin/zettelflow split
```
This creates new `.md` files in your `split` data directory.

**Stage 3: Enrich**
Process all the notes in the `split` directory, using the LLM to intelligently fill in the YAML frontmatter fields.

```sh
./bin/zettelflow enrich
```
This saves the final, completed notes to your `enrich` data directory.

## Commands

*   `./bin/zettelflow ingest [file]` - Processes text from a file or stdin.
*   `./bin/zettelflow split [file]` - Splits an ingested file into note stubs. If no file is provided, it uses the latest one.
*   `./bin/zettelflow enrich [directory]` - Enriches all notes in the `split` directory.
*   `./bin/zettelflow list` - Lists files in a processed stage.
*   `./bin/zettelflow clean` - Removes files from a processed stage.
*   `./bin/zettelflow config path` - Prints the path to your configuration directory.

## Configuration

ZettelFlow is highly configurable. On the first run, it creates a configuration directory at `~/.config/zettelflow`. You can find this path easily by running:

```sh
./bin/zettelflow config path
```

Inside this directory, you will find:
*   `config.yaml`: The main configuration file. You can change paths, models, and other settings here.
*   `prompts/`: Contains the `default_ingest.md` and `default_enrich.md` prompts. You can edit these to change the LLM's behavior.
*   `templates/`: Contains the `note_header.yml` template used by the `split` command.
