# ZettelFlow

ZettelFlow is a powerful, CLI-driven pipeline for transforming raw, unstructured text into a collection of structured, inter-connected notes, ready for any Zettelkasten or personal knowledge management (PKM) system.

It uses a three-stage, LLM-powered workflow to distill your ideas into a clean, organized format.

## The ZettelFlow Pipeline

The application is designed around a simple, three-stage data pipeline. Each stage has a dedicated input and output directory, allowing you to inspect the results at each step.

1.  **`ingest`**: This is the entry point. You provide raw text and the application combines it with a prompt and sends it to an LLM. The LLM's processed, semi-structured text is saved to the `ingest` data directory. The `ingest` command is highly flexible and can accept input in several ways:
    *   **From a single file:** `./bin/zettelflow ingest my_note.txt`
    *   **From a whole directory:** `./bin/zettelflow ingest my_notes_folder/` (This will process every file in the directory).
    *   **From a pipe:** `cat my_note.txt | ./bin/zettelflow ingest`
    *   **From direct input:** Run `./bin/zettelflow ingest` and type or paste directly into the terminal.

2.  **`split`**: This command processes all files currently in the `ingest` directory. It splits each file into multiple chunks based on a delimiter (`###` by default) and formats each chunk into a structured note stub using a template. These stubs, which now have empty YAML frontmatter, are saved to the `split` data directory. The original files from the `ingest` directory are then moved to a `processed` subdirectory to prevent them from being processed again.

3.  **`enrich`**: This is the final stage. The command processes all note stubs in the `split` directory. For each note, it sends the entire content to an LLM with a prompt that asks it to intelligently fill in the YAML frontmatter fields (like `title`, `tags`, etc.). The final, completed notes are saved to the `enrich` data directory.

This entire workflow is designed to be idempotent and inspectable. You can run the commands multiple times, and the use of `processed` subdirectories prevents duplicate work.

## Quick Start

To get started, you need a Go environment (version 1.23 or later) and an OpenAI-compatible API key.

### 1. Build the Application

From the project root, run the `make` command to build the `zettelflow` binary into the `bin/` directory:

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
Process all pending files from the `ingest` stage.

```sh
./bin/zettelflow split
```
This creates new `.md` files in your `split` data directory and moves the processed files to `ingest/processed`.

**Stage 3: Enrich**
Process all the notes in the `split` directory.

```sh
./bin/zettelflow enrich
```
This saves the final, completed notes to your `enrich` data directory.

## Commands

*   `./bin/zettelflow ingest [path]`: Processes text from a file, a directory, or stdin and saves the results to the `ingest` directory.
*   `./bin/zettelflow split`: Splits all pending files from the `ingest` directory into note stubs in the `split` directory.
*   `./bin/zettelflow enrich`: Enriches all notes from the `split` directory and saves them to the `enrich` directory.

### Utility Commands

*   `./bin/zettelflow list <stage>`: Lists the files in a specific stage's data directory. The `<stage>` can be `ingest`, `split`, `enrich`, or `all`.
*   `./bin/zettelflow clean <stage>`: Deletes all files from a specific stage's data directory. The `<stage>` can be `ingest`, `split`, `enrich`, or `all`. Use the `-d` or `--dry-run` flag to see what would be deleted without actually deleting anything.
*   `./bin/zettelflow config path`: Prints the absolute path to your configuration directory, making it easy to find and edit your settings.

## Configuration

ZettelFlow is highly configurable. On the first run, it creates a configuration directory at `~/.config/zettelflow`. You can find this path easily by running:

```sh
./bin/zettelflow config path
```

Inside this directory, you will find:
*   `config.yaml`: The main configuration file. This is where you can change data paths, API settings, and tune the LLM parameters for each stage of the pipeline.
*   `prompts/`: Contains the `default_ingest.md` and `default_enrich.md` prompts. You can edit these to change the LLM's behavior.
*   `templates/`: Contains the `note_header.yml` template used by the `split` command.

### Per-Stage LLM Configuration

You can configure the LLM settings independently for the `ingest` and `enrich` stages to optimize for cost and quality. The `config.yaml` file allows you to set the following for each stage:
*   `model`: The specific model to use (e.g., `gpt-4o`, `gpt-4o-mini`).
*   `temperature`: Controls the creativity of the output (e.g., `0.5`).
*   `max_completion_tokens`: The maximum number of tokens to generate in the response.
