# Changelog

## X.X.X - YYYY-MM-DD
- Changed: Describe the subject style without example wording the model copied into subjects
- Changed: Split the 4 KB diff budget evenly across staged files so every file keeps its header
- Changed: Regenerate asks for new subjects instead of repeating the same list
- Changed: Send file status (added, modified, deleted, renamed) instead of diff stat
- Changed: Send one line of diff context so more changes fit in the budget
- Changed: Ask for specific, distinct subjects that describe the change's intent
- Changed: Leave source maps, SVGs, dist, vendor and snapshot files out of the diff
- Changed: Cap the staged file list sent to the model at 50 files

## 0.1.2 - 2026-11-28
- Added: KOMIT_REASONING_EFFORT to set the model's reasoning effort, e.g. none for faster replies
- Added: Suggest KOMIT_REASONING_EFFORT=none when the reply hits the token limit
- Changed: Send at most 4 KB of the staged diff to the model
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
