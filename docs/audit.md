# Audit Summary

This document summarizes the main security, reliability, and usability findings from a code audit of the `hyprvoice` package.

## Scope

Reviewed areas:

- daemon lifecycle and IPC
- config loading, validation, and persistence
- recording, transcription, LLM post-processing, and injection pipeline
- local model download and storage
- onboarding and configuration UX

## Key Findings

### 1. Transcript content is logged

Several code paths log raw transcription text and LLM-processed text. This creates a privacy risk because dictated content may be stored in `systemd --user` journal logs or other log collectors.

Affected areas include:

- pipeline final transcription and processed output
- batch transcription adapters
- local `whisper.cpp` transcription
- LLM post-processing adapters

Impact:

- private dictated content can persist outside the target application
- logs may expose sensitive business, personal, or credential-related speech

Recommended action:

- remove transcript payloads from default logs
- add a debug mode if payload logging is ever needed
- keep only metadata such as provider, duration, byte counts, and error class

### 2. API keys are stored in plaintext with weak file-permission guarantees

Provider API keys are serialized directly into the config file. The config directory is created with permissive defaults, and file creation relies on process `umask` rather than enforcing strict file permissions.

Impact:

- secrets may be readable by other local users on misconfigured systems
- secret handling depends too heavily on environment defaults

Recommended action:

- write config files with `0600`
- create config directories with `0700`
- save atomically via temp file + rename
- prefer environment variables or a keyring-backed store for secrets

### 3. Recording timeout applies to the whole pipeline

The current timeout starts before recording and remains in force for transcription, LLM cleanup, and text injection. Long dictation can leave insufficient time for later stages.

Impact:

- recordings may succeed but fail during transcription or injection
- failures may be hard to diagnose because the timeout spans multiple phases

Recommended action:

- split timeouts by stage
- keep a recording timeout separate from post-processing / injection deadlines

### 4. Pipeline monitor goroutines can accumulate

Each new pipeline run starts monitoring goroutines for notifications and errors. Those goroutines terminate only when the daemon exits, not when an individual pipeline ends.

Impact:

- repeated use can leak goroutines
- long-running daemon sessions may accumulate unnecessary background work

Recommended action:

- close pipeline-owned channels when a pipeline ends, or
- bind monitor goroutines to a per-pipeline context instead of the daemon context

### 5. IPC read path is unbounded

The daemon protocol expects a single-character command, but the handler reads a full newline-terminated string without a size limit.

Impact:

- a local client can force unnecessary allocation and noisy error handling
- this is a local robustness issue rather than a remote exploit surface

Recommended action:

- cap command length
- read exactly the expected protocol size where possible

### 6. Local model downloads are not integrity-verified

Whisper model downloads rely on HTTPS and expected file size, but do not verify checksums or signatures.

Impact:

- corrupted or tampered model files are not detected explicitly
- local transcription depends on unverified downloaded artifacts

Recommended action:

- publish and verify SHA256 checksums for bundled model metadata
- fail closed on checksum mismatch

### 7. Config manager returns a shallow copy

`GetConfig()` returns a copied struct, but map and slice fields are still shared references.

Impact:

- callers can mutate shared config state unintentionally
- this weakens thread-safety assumptions around config access

Recommended action:

- deep-copy maps and slices before returning config values

### 8. Config saves are not atomic

The config writer truncates and rewrites the target file directly while the daemon is also watching that file for reloads.

Impact:

- editors or interrupted writes can produce transient parse failures
- reload behavior can be noisy or confusing

Recommended action:

- write to a temporary file in the same directory
- `fsync` if needed
- rename into place atomically

### 9. `processing` state is not handled cleanly by toggle behavior

The pipeline has a `processing` state for LLM cleanup, but the daemon toggle logic does not treat it explicitly.

Impact:

- users can receive a success response from `toggle` while nothing happens
- the runtime state model is harder to reason about from the CLI

Recommended action:

- define toggle behavior for `processing`
- expose that state clearly in `status`

### 10. Clipboard backend overwrites the clipboard

Clipboard-based injection replaces the user clipboard and does not restore it afterward.

Impact:

- this is disruptive in normal desktop use
- it makes clipboard fallback feel unreliable even when injection succeeds

Recommended action:

- snapshot and restore clipboard contents when possible
- document the behavior if restoration is unavailable on a given platform

## Prioritized Remediation Plan

### Phase 1: Privacy and secret handling

1. Remove transcript and processed-text payloads from default logs.
2. Enforce `0600` config files and `0700` config directories.
3. Save config atomically.
4. Decide whether plaintext config storage remains acceptable or whether provider secrets should move to environment-only or keyring storage.

### Phase 2: Runtime correctness

1. Split the global pipeline timeout into stage-specific timeouts.
2. Fix pipeline monitor goroutine lifetime.
3. Deep-copy config returned by the manager.
4. Bound daemon command reads.

### Phase 3: Supply-chain and artifact trust

1. Add checksums to whisper model metadata.
2. Verify downloads before marking a model installed.
3. Improve error reporting for failed verification.

### Phase 4: Usability improvements

1. Add a `hyprvoice doctor` command to validate runtime dependencies and environment.
2. Preserve and restore clipboard contents for clipboard injection mode.
3. Improve state reporting so `recording`, `transcribing`, `processing`, and `injecting` are all surfaced consistently.
4. Make onboarding and `configure` explain backend tradeoffs more clearly, especially privacy vs. quality and clipboard side effects.

## Suggested Near-Term Deliverables

The highest-value short-term changes would be:

1. log redaction for transcript content
2. stricter config file permissions
3. atomic config saves
4. stage-specific timeouts
5. a `doctor` command

These changes would reduce privacy risk immediately and make the package easier to operate and debug.
