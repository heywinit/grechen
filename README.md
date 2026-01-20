# grechen

personal ai assistant for task and commitment management. cli-first, lightweight, designed to be an externalized memory and accountability engine.

## setup

```bash
go build -o grechen ./cmd/grechen
```

set `GEMINI_API_KEY` in your environment or `.env` file. defaults to gemini for extraction.

data lives in `~/.grechen` by default, or set `GRECHEN_DATA_DIR` to customize.

run `grechen setup` to initialize.

## usage

just talk to it:

```bash
grechen told deep pr will be ready by tomorrow
grechen had lunch. gonna sit for working
grechen gemini integration is almost done
```

commands:

- `grechen <natural language>` - log activities, create commitments, update progress
- `grechen today` - situational awareness, open commitments
- `grechen commitments` - view all commitments
- `grechen goodnight` - daily evaluation, pattern checks, questions
- `grechen review` - stats summary and pattern alerts
- `grechen thats-wrong` - correction flow

## how it works

natural language input gets parsed into structured data (commitments, progress, logs). everything is append-only. daily markdown files in `daily/`, metadata in `meta/`. patterns get detected automatically - late starts, sparse logs, commitment silence, that sort of thing.

goodnight routine compares today to rolling averages and asks targeted questions when things look off.
