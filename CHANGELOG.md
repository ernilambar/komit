# Changelog

## X.X.X - YYYY-MM-DD
- Changed: Send at most 4 KB of diff to the model by default
- Changed: Ask the model for a short, deterministic reply and regenerate only when requested

## 0.1.1 - 2026-11-27
- Added: Press e to edit a suggestion and r to regenerate
- Added: Show a spinner while generating subjects
- Added: Unversioned release archive for a stable latest download URL
- Fixed: Send the full staged diff when run from a subdirectory
- Fixed: Reject unexpected arguments instead of committing
- Fixed: Drop model preamble and stray text from suggestions

## 0.1.0 - 2026-11-27
- Added: Generate commit subjects from the staged diff with an OpenAI-compatible model
- Added: Subject picker with regenerate and write-your-own options
- Added: `komit status` to check configuration and server connectivity
- Added: Flags for staging all changes, dry run, skipping the push, candidate count, and model override
